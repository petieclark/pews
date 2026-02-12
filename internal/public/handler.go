package public

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Rate limiter: simple IP-based, 10 req/min
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{requests: make(map[string][]time.Time)}
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// Clean old entries
	times := rl.requests[ip]
	clean := make([]time.Time, 0, len(times))
	for _, t := range times {
		if t.After(cutoff) {
			clean = append(clean, t)
		}
	}

	if len(clean) >= 10 {
		rl.requests[ip] = clean
		return false
	}

	rl.requests[ip] = append(clean, now)
	return true
}

type Handler struct {
	db      *pgxpool.Pool
	limiter *rateLimiter
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db, limiter: newRateLimiter()}
}

func (h *Handler) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func (h *Handler) rateLimit(w http.ResponseWriter, r *http.Request) bool {
	if !h.limiter.Allow(h.getClientIP(r)) {
		http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
		return false
	}
	return true
}

func (h *Handler) getTenantBySlug(ctx context.Context, slug string) (*tenantInfo, error) {
	var t tenantInfo
	err := h.db.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(logo, ''), COALESCE(about, ''),
		       COALESCE(address_line1, ''), COALESCE(address_line2, ''),
		       COALESCE(city, ''), COALESCE(state, ''), COALESCE(zip, ''),
		       COALESCE(phone, ''), COALESCE(website, ''), COALESCE(email, '')
		FROM tenants WHERE slug = $1`, slug).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Logo, &t.About,
		&t.AddressLine1, &t.AddressLine2, &t.City, &t.State, &t.Zip,
		&t.Phone, &t.Website, &t.Email,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type tenantInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Logo         string `json:"logo,omitempty"`
	About        string `json:"about,omitempty"`
	AddressLine1 string `json:"address_line1,omitempty"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Zip          string `json:"zip,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
}

type serviceTime struct {
	Name        string `json:"name"`
	DefaultDay  string `json:"default_day,omitempty"`
	DefaultTime string `json:"default_time,omitempty"`
}

type publicGroup struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	GroupType       string `json:"group_type"`
	MeetingDay      string `json:"meeting_day,omitempty"`
	MeetingTime     string `json:"meeting_time,omitempty"`
	MeetingLocation string `json:"meeting_location,omitempty"`
	PhotoURL        string `json:"photo_url,omitempty"`
	MemberCount     int    `json:"member_count"`
	LeaderName      string `json:"leader_name,omitempty"`
}

type publicEvent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	AllDay      bool      `json:"all_day"`
	EventType   string    `json:"event_type"`
}

// GetChurchInfo returns public church information
func (h *Handler) GetChurchInfo(w http.ResponseWriter, r *http.Request) {
	if !h.rateLimit(w, r) {
		return
	}

	slug := r.URL.Query().Get("tenant_slug")
	if slug == "" {
		http.Error(w, `{"error":"tenant_slug required"}`, http.StatusBadRequest)
		return
	}

	t, err := h.getTenantBySlug(r.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
			return
		}
		log.Printf("GetChurchInfo error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// Get service times
	rows, err := h.db.Query(r.Context(), `
		SELECT name, COALESCE(default_day, ''), COALESCE(default_time, '')
		FROM service_types WHERE tenant_id = $1 AND is_active = true
		ORDER BY name`, t.ID)
	if err != nil {
		log.Printf("GetChurchInfo service_types error: %v", err)
	}

	var serviceTimes []serviceTime
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var st serviceTime
			if err := rows.Scan(&st.Name, &st.DefaultDay, &st.DefaultTime); err == nil {
				serviceTimes = append(serviceTimes, st)
			}
		}
	}

	resp := map[string]interface{}{
		"church":        t,
		"service_times": serviceTimes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetEvents returns upcoming public events
func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	if !h.rateLimit(w, r) {
		return
	}

	slug := r.URL.Query().Get("tenant_slug")
	if slug == "" {
		http.Error(w, `{"error":"tenant_slug required"}`, http.StatusBadRequest)
		return
	}

	t, err := h.getTenantBySlug(r.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	now := time.Now()
	end := now.AddDate(0, 0, 30)

	rows, err := h.db.Query(r.Context(), `
		SELECT id, title, COALESCE(description, ''), COALESCE(location, ''),
		       start_time, end_time, all_day, event_type
		FROM events
		WHERE tenant_id = $1 AND start_time >= $2 AND start_time <= $3
		ORDER BY start_time`, t.ID, now, end)
	if err != nil {
		log.Printf("GetEvents error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []publicEvent{}
	for rows.Next() {
		var e publicEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Location,
			&e.StartTime, &e.EndTime, &e.AllDay, &e.EventType); err != nil {
			log.Printf("GetEvents scan error: %v", err)
			continue
		}
		events = append(events, e)
	}

	resp := map[string]interface{}{
		"events": events,
		"church": map[string]string{"name": t.Name, "slug": t.Slug, "logo": t.Logo},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPublicGroups returns public groups with leader info
func (h *Handler) GetPublicGroups(w http.ResponseWriter, r *http.Request) {
	if !h.rateLimit(w, r) {
		return
	}

	slug := r.URL.Query().Get("tenant_slug")
	if slug == "" {
		http.Error(w, `{"error":"tenant_slug required"}`, http.StatusBadRequest)
		return
	}

	t, err := h.getTenantBySlug(r.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"church not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT g.id, g.name, COALESCE(g.description, ''), g.group_type,
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''),
		       COALESCE(g.meeting_location, ''), COALESCE(g.photo_url, ''),
		       COUNT(DISTINCT gm.id) as member_count,
		       COALESCE(
		           (SELECT p.first_name || ' ' || p.last_name
		            FROM group_members lm
		            JOIN people p ON p.id = lm.person_id
		            WHERE lm.group_id = g.id AND lm.role = 'leader'
		            LIMIT 1), '') as leader_name
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id
		WHERE g.tenant_id = $1 AND g.is_public = TRUE AND g.is_active = TRUE
		GROUP BY g.id ORDER BY g.name`, t.ID)
	if err != nil {
		log.Printf("GetPublicGroups error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []publicGroup{}
	for rows.Next() {
		var g publicGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.GroupType,
			&g.MeetingDay, &g.MeetingTime, &g.MeetingLocation, &g.PhotoURL,
			&g.MemberCount, &g.LeaderName); err != nil {
			log.Printf("GetPublicGroups scan error: %v", err)
			continue
		}
		groups = append(groups, g)
	}

	resp := map[string]interface{}{
		"groups": groups,
		"church": map[string]string{"name": t.Name, "slug": t.Slug, "logo": t.Logo},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type groupSignupRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Interest   string `json:"interest"`
}

// GroupSignup handles public group signup
func (h *Handler) GroupSignup(w http.ResponseWriter, r *http.Request) {
	if !h.rateLimit(w, r) {
		return
	}

	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		http.Error(w, `{"error":"group id required"}`, http.StatusBadRequest)
		return
	}

	var req groupSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		http.Error(w, `{"error":"first_name, last_name, and email are required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get group and verify it's public
	var tenantID, groupName string
	var isPublic, isActive bool
	err := h.db.QueryRow(ctx, `
		SELECT tenant_id, name, is_public, is_active FROM groups WHERE id = $1`,
		groupID).Scan(&tenantID, &groupName, &isPublic, &isActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"group not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	if !isPublic || !isActive {
		http.Error(w, `{"error":"group not available for signup"}`, http.StatusBadRequest)
		return
	}

	// Check if person exists by email, create if not
	var personID string
	err = h.db.QueryRow(ctx, `SELECT id FROM people WHERE tenant_id = $1 AND LOWER(email) = $2`,
		tenantID, req.Email).Scan(&personID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create new person
			personID = uuid.New().String()
			_, err = h.db.Exec(ctx, `
				INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, membership_status)
				VALUES ($1, $2, $3, $4, $5, $6, 'visitor')`,
				personID, tenantID, req.FirstName, req.LastName, req.Email, req.Phone)
			if err != nil {
				log.Printf("GroupSignup create person error: %v", err)
				http.Error(w, `{"error":"failed to process signup"}`, http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
	}

	// Check if already a member
	var alreadyMember bool
	err = h.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND person_id = $2)`,
		groupID, personID).Scan(&alreadyMember)
	if err == nil && alreadyMember {
		// Still show success - don't leak membership info
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Thanks! The group leader will reach out soon."})
		return
	}

	// Add as pending member
	memberID := uuid.New().String()
	_, err = h.db.Exec(ctx, `
		INSERT INTO group_members (id, group_id, person_id, role) VALUES ($1, $2, $3, 'pending')`,
		memberID, groupID, personID)
	if err != nil {
		log.Printf("GroupSignup add member error: %v", err)
		http.Error(w, `{"error":"failed to process signup"}`, http.StatusInternalServerError)
		return
	}

	// Find group leader to assign follow-up
	var leaderUserID *string
	_ = h.db.QueryRow(ctx, `
		SELECT u.id FROM group_members gm
		JOIN people p ON p.id = gm.person_id
		JOIN users u ON LOWER(u.email) = LOWER(p.email)
		WHERE gm.group_id = $1 AND gm.role = 'leader'
		LIMIT 1`, groupID).Scan(&leaderUserID)

	// Create follow-up
	followUpID := uuid.New().String()
	title := fmt.Sprintf("Group signup: %s %s → %s", req.FirstName, req.LastName, groupName)
	notes := ""
	if req.Interest != "" {
		notes = fmt.Sprintf("Interest: %s", req.Interest)
	}
	_, err = h.db.Exec(ctx, `
		INSERT INTO follow_ups (id, tenant_id, person_id, assigned_to, title, type, priority, status)
		VALUES ($1, $2, $3, $4, $5, 'general', 'medium', 'new')`,
		followUpID, tenantID, personID, leaderUserID, title)
	if err != nil {
		log.Printf("GroupSignup create follow-up error: %v", err)
		// Don't fail the signup over this
	}

	// Add note to follow-up if interest provided
	if notes != "" && err == nil {
		noteID := uuid.New().String()
		h.db.Exec(ctx, `INSERT INTO follow_up_notes (id, follow_up_id, note) VALUES ($1, $2, $3)`,
			noteID, followUpID, notes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Thanks! The group leader will reach out soon."})
}
