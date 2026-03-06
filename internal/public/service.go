package public

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AssignmentTokenData holds validated token claims
type AssignmentTokenData struct {
	AssignmentID string
	PersonID     string
	ServiceName  string
	ServiceDate  string
}

// Service handles public response validation and processing
type Service struct {
	db          *pgxpool.Pool
	tokenSecret string
	expiryHours int
}

// NewService creates a new public service
func NewService(db *pgxpool.Pool, jwtSecret string) *Service {
	return &Service{
		db:          db,
		tokenSecret: jwtSecret,
		expiryHours: 168, // 7 days
	}
}

// ValidateAssignmentToken validates a token and returns assignment details
func (s *Service) ValidateAssignmentToken(tokenString string) (*AssignmentTokenData, error) {
	parts := splitToken(tokenString)
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	signature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature
	h := hmac.New(sha256.New, []byte(s.tokenSecret))
	h.Write(payload)
	expectedSig := h.Sum(nil)

	if !hmac.Equal(signature, expectedSig) {
		return nil, errors.New("invalid token signature")
	}

	// Parse payload
	var claims struct {
		AssignmentID string    `json:"assignment_id"`
		PersonID     string    `json:"person_id"`
		ExpiresAt    time.Time `json:"expires_at"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	// Check expiry
	if time.Now().After(claims.ExpiresAt) {
		return nil, errors.New("token has expired")
	}

	// Get assignment details from database
	serviceName, serviceDate, err := s.getAssignmentDetails(s.db, claims.AssignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment details: %w", err)
	}

	return &AssignmentTokenData{
		AssignmentID: claims.AssignmentID,
		PersonID:     claims.PersonID,
		ServiceName:  serviceName,
		ServiceDate:  serviceDate,
	}, nil
}

func (s *Service) getAssignmentDetails(db *pgxpool.Pool, assignmentID string) (string, string, error) {
	var serviceName, serviceDate string

	err := db.QueryRow(context.Background(), `
		SELECT s.name, TO_CHAR(s.service_date, 'Month DD, YYYY')
		FROM service_team_assignments sta
		JOIN services s ON s.id = sta.service_id
		WHERE sta.id = $1`,
		assignmentID).Scan(&serviceName, &serviceDate)

	if err != nil {
		return "", "", fmt.Errorf("failed to query assignment: %w", err)
	}

	return serviceName, serviceDate, nil
}

// ProcessResponse handles accept/decline responses and updates the database
func (s *Service) ProcessResponse(ctx context.Context, assignmentID, personID, action string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update assignment status based on response
	var newStatus string
	var respondedAt *time.Time

	if action == "accept" {
		newStatus = "confirmed"
		now := time.Now()
		respondedAt = &now
	} else {
		newStatus = "declined"
		now := time.Now()
		respondedAt = &now
	}

	result, err := tx.Exec(ctx, `
		UPDATE service_team_assignments 
		SET status = $1, responded_at = $2, updated_at = NOW()
		WHERE id = $3 AND person_id = $4`,
		newStatus, respondedAt, assignmentID, personID,
	)

	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("assignment not found or person mismatch")
	}

	// Mark notification as sent if it wasn't already
	_, _ = tx.Exec(ctx, `
		UPDATE service_team_assignments 
		SET notification_sent = true, notified_at = NOW()
		WHERE id = $1`,
		assignmentID,
	)

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetChurchInfo returns public church information
func (s *Service) GetChurchInfo(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	var name, slug, logo, about string
	err := s.db.QueryRow(ctx,
		`SELECT name, COALESCE(slug, ''), COALESCE(logo, ''), COALESCE(about, '') FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&name, &slug, &logo, &about)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"name": name, "slug": slug, "logo": logo, "about": about,
	}, nil
}

// GetPublicEvents returns upcoming public events
func (s *Service) GetPublicEvents(ctx context.Context, tenantID string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, COALESCE(description, ''), start_time, end_time, COALESCE(location, '')
		 FROM events WHERE tenant_id = $1 AND is_public = true AND start_time >= NOW()
		 ORDER BY start_time LIMIT 20`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	for rows.Next() {
		var id, name, description, location string
		var startTime, endTime time.Time
		if err := rows.Scan(&id, &name, &description, &startTime, &endTime, &location); err != nil {
			return nil, err
		}
		events = append(events, map[string]interface{}{
			"id": id, "name": name, "description": description,
			"start_time": startTime, "end_time": endTime, "location": location,
		})
	}
	return events, nil
}

// GetPublicGroups returns public groups for the group finder
func (s *Service) GetPublicGroups(ctx context.Context, tenantID string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, COALESCE(description, ''), group_type, COALESCE(meeting_day, ''),
		        COALESCE(meeting_time, ''), COALESCE(meeting_location, ''), COALESCE(photo_url, '')
		 FROM groups WHERE tenant_id = $1 AND is_public = true AND is_active = true
		 ORDER BY name`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []map[string]interface{}{}
	for rows.Next() {
		var id, name, description, groupType, meetingDay, meetingTime, meetingLocation, photoURL string
		if err := rows.Scan(&id, &name, &description, &groupType, &meetingDay, &meetingTime, &meetingLocation, &photoURL); err != nil {
			return nil, err
		}
		groups = append(groups, map[string]interface{}{
			"id": id, "name": name, "description": description, "group_type": groupType,
			"meeting_day": meetingDay, "meeting_time": meetingTime, "meeting_location": meetingLocation,
			"photo_url": photoURL,
		})
	}
	return groups, nil
}

// GroupSignup handles a public group signup request
func (s *Service) GroupSignup(ctx context.Context, groupID, name, email, phone string) error {
	// Verify group is public and active
	var tenantID string
	err := s.db.QueryRow(ctx,
		`SELECT tenant_id FROM groups WHERE id = $1 AND is_public = true AND is_active = true`,
		groupID,
	).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("group not found or not accepting signups")
	}

	// Store the signup as a connection card-style record
	_, err = s.db.Exec(ctx,
		`INSERT INTO connection_cards (id, tenant_id, name, email, phone, notes, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())`,
		tenantID, name, email, phone, "Group signup: "+groupID,
	)
	return err
}

func splitToken(token string) []string {
	parts := make([]string, 0, 2)
	current := ""
	inQuotes := false
	
	for _, c := range token {
		if c == '"' {
			inQuotes = !inQuotes
			current += string(c)
		} else if c == '.' && !inQuotes {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}
