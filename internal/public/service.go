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
