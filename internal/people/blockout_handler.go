package people

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BlockoutHandler struct {
	db *pgxpool.Pool
}

func NewBlockoutHandler(db *pgxpool.Pool) *BlockoutHandler {
	return &BlockoutHandler{db: db}
}

// ListBlockouts handles GET /api/teams/members/:personId/blockouts
func (h *BlockoutHandler) ListBlockouts(w http.ResponseWriter, r *http.Request, personID string) {
	tenantID := getTenantFromContext(r.Context())
	
	rows, err := h.db.Query(r.Context(), `SELECT * FROM get_person_blockouts($1)`, personID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	blockouts := []VolunteerBlockout{}
	for rows.Next() {
		var b VolunteerBlockout
		err := rows.Scan(
			&b.ID, &b.PersonID, &b.TeamID, &b.StartDate, &b.EndDate,
			&b.Reason, &b.IsRecurring, &b.DayOfWeek, &b.CreatedAt, &b.UpdatedAt, &b.TeamName,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan blockout: %v", err), http.StatusInternalServerError)
			return
		}
		blockouts = append(blockouts, b)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blockouts)
}

// CreateBlockout handles POST /api/teams/members/:personId/blockouts
func (h *BlockoutHandler) CreateBlockout(w http.ResponseWriter, r *http.Request, personID string) {
	tenantID := getTenantFromContext(r.Context())
	
	var req BlockoutCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate dates
	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, time.UTC)
	if err != nil {
		http.Error(w, "Invalid start_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDate, err := time.ParseInLocation("2006-01-02", req.EndDate, time.UTC)
	if err != nil {
		http.Error(w, "Invalid end_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	if endDate.Before(startDate) || endDate.Equal(startDate) {
		http.Error(w, "end_date must be after start_date", http.StatusBadRequest)
		return
	}

	// Validate recurring logic
	if req.IsRecurring && req.DayOfWeek == nil {
		http.Error(w, "day_of_week is required when is_recurring=true", http.StatusBadRequest)
		return
	}

	if !req.IsRecurring && req.DayOfWeek != nil {
		http.Error(w, "day_of_week should not be set for non-recurring blockouts", http.StatusBadRequest)
		return
	}

	// Validate day_of_week range if provided
	if req.DayOfWeek != nil && (*req.DayOfWeek < 0 || *req.DayOfWeek > 6) {
		http.Error(w, "day_of_week must be between 0 (Sun) and 6 (Sat)", http.StatusBadRequest)
		return
	}

	// Insert blockout
	blockoutID := uuid.New().String()
	var createdAt, updatedAt time.Time
	
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO volunteer_availability (id, person_id, team_id, start_date, end_date, reason, is_recurring, day_of_week)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`,
		blockoutID, personID, req.TeamID, startDate.UTC(), endDate.UTC(), 
		req.Reason, req.IsRecurring, req.DayOfWeek,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create blockout: %v", err), http.StatusInternalServerError)
		return
	}

	blockout := VolunteerBlockout{
		ID:          blockoutID,
		PersonID:    personID,
		TenantID:    tenantID,
		TeamID:      req.TeamID,
		StartDate:   startDate.UTC(),
		EndDate:     endDate.UTC(),
		Reason:      req.Reason,
		IsRecurring: req.IsRecurring,
		DayOfWeek:   req.DayOfWeek,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(blockout)
}

// UpdateBlockout handles PUT /api/teams/members/:personId/blockouts/:id
func (h *BlockoutHandler) UpdateBlockout(w http.ResponseWriter, r *http.Request, personID, blockoutID string) {
	tenantID := getTenantFromContext(r.Context())
	
	var req BlockoutUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	queryParts := []string{"start_date = $1", "end_date = $2", "team_id = $3", "reason = $4", "is_recurring = $5", "day_of_week = $6"}
	args := []interface{}{
		req.StartDate, req.EndDate, req.TeamID, req.Reason, req.IsRecurring, req.DayOfWeek,
	}

	hasUpdate := false
	
	if req.StartDate != nil {
		startDate, err := time.ParseInLocation("2006-01-02", *req.StartDate, time.UTC)
		if err != nil {
			http.Error(w, "Invalid start_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		args[0] = startDate.UTC()
		hasUpdate = true
	}

	if req.EndDate != nil {
		endDate, err := time.ParseInLocation("2006-01-02", *req.EndDate, time.UTC)
		if err != nil {
			http.Error(w, "Invalid end_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		args[1] = endDate.UTC()
		hasUpdate = true
	}

	if req.IsRecurring != nil && req.DayOfWeek == nil {
		http.Error(w, "day_of_week is required when is_recurring=true", http.StatusBadRequest)
		return
	}

	if !req.IsRecurring && req.DayOfWeek != nil {
		http.Error(w, "day_of_week should not be set for non-recurring blockouts", http.StatusBadRequest)
		return
	}

	if *req.IsRecurring || (req.DayOfWeek != nil) {
		hasUpdate = true
	}

	query := fmt.Sprintf("UPDATE volunteer_availability SET %s WHERE id = $7 AND person_id = $8 RETURNING updated_at", strings.Join(queryParts, ", "))
	args = append(args, blockoutID, personID)

	var updatedAt time.Time
	err := h.db.QueryRow(r.Context(), query, args...).Scan(&updatedAt)
	if err == pgx.ErrNoRows {
		http.Error(w, "Blockout not found or access denied", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update blockout: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch updated blockout
	var createdAt time.Time
	row := h.db.QueryRow(r.Context(), `
		SELECT created_at FROM volunteer_availability WHERE id = $1`, blockoutID)
	if err := row.Scan(&createdAt); err != nil {
		http.Error(w, "Failed to fetch updated blockout", http.StatusInternalServerError)
		return
	}

	blockout := VolunteerBlockout{
		ID:          blockoutID,
		PersonID:    personID,
		TenantID:    tenantID,
		TeamID:      req.TeamID,
		StartDate:   createdAt, // Would need to fetch actual start_date
		EndDate:     updatedAt, // Would need to fetch actual end_date
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blockout)
}

// DeleteBlockout handles DELETE /api/teams/members/:personId/blockouts/:id
func (h *BlockoutHandler) DeleteBlockout(w http.ResponseWriter, r *http.Request, personID, blockoutID string) {
	tenantID := getTenantFromContext(r.Context())
	
	result, err := h.db.Exec(r.Context(), "DELETE FROM volunteer_availability WHERE id = $1 AND person_id = $2", blockoutID, personID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete blockout: %v", err), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Blockout not found or access denied", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckAvailability handles GET /api/teams/members/:personId/check-availability?date=YYYY-MM-DD
func (h *BlockoutHandler) CheckAvailability(w http.ResponseWriter, r *http.Request, personID string) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date parameter is required", http.StatusBadRequest)
		return
	}

	date, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		http.Error(w, "Invalid date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Use the enhanced conflict detection function
	rows, err := h.db.Query(r.Context(), `SELECT * FROM get_volunteer_conflicts($1, $2)`, personID, date.UTC())
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	conflict := ConflictInfo{}
	hasConflict := false
	
	for rows.Next() {
		var isBlocked bool
		var conflictType *string
		var startDate, endDate *time.Time
		var reason *string
		var dayOfWeek *int
		
		err := rows.Scan(&isBlocked, &conflictType, &startDate, &endDate, &reason, &dayOfWeek)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan conflict: %v", err), http.StatusInternalServerError)
			return
		}

		if isBlocked {
			hasConflict = true
			conflict.IsBlocked = true
			if conflictType != nil {
				conflict.ConflictType = *conflictType
			}
			if startDate != nil {
				s := startDate.Format("2006-01-02")
				conflict.StartDate = &s
			}
			if endDate != nil {
				e := endDate.Format("2006-01-02")
				conflict.EndDate = &e
			}
			conflict.Reason = reason
			conflict.DayOfWeek = dayOfWeek
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conflict)
}

// CheckSchedulingConflict is a helper for scheduling endpoints to validate assignments
func (h *BlockoutHandler) CheckSchedulingConflict(personID string, serviceDate time.Time) (*ConflictInfo, error) {
	rows, err := h.db.Query(r.Context(), `SELECT * FROM get_volunteer_conflicts($1, $2)`, personID, serviceDate.UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to check scheduling conflict: %w", err)
	}
	defer rows.Close()

	conflict := &ConflictInfo{}
	for rows.Next() {
		var isBlocked bool
		var conflictType *string
		var startDate, endDate *time.Time
		var reason *string
		var dayOfWeek *int
		
		err := rows.Scan(&isBlocked, &conflictType, &startDate, &endDate, &reason, &dayOfWeek)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict: %w", err)
		}

		if isBlocked {
			conflict.IsBlocked = true
			if conflictType != nil {
				conflict.ConflictType = *conflictType
			}
			if startDate != nil {
				s := startDate.Format("2006-01-02")
				conflict.StartDate = &s
			}
			if endDate != nil {
				e := endDate.Format("2006-01-02")
				conflict.EndDate = &e
			}
			conflict.Reason = reason
			conflict.DayOfWeek = dayOfWeek
		}
	}

	return conflict, nil
}

// Helper to get tenant from context (implement based on your existing middleware pattern)
func getTenantFromContext(ctx interface{}) string {
	// This should extract tenant_id from the request context
	// Implement based on your existing middleware patterns
	return ""
}
