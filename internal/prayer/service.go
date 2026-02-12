package prayer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) ListPrayerRequests(ctx context.Context, tenantID string, filter PrayerRequestFilter, userID *string) ([]PrayerRequest, error) {
	query := `
		SELECT pr.id, pr.tenant_id, pr.person_id, COALESCE(p.first_name || ' ' || p.last_name, '') as person_name,
		       pr.name, pr.email, pr.request_text, pr.is_public, pr.status, pr.connection_card_id,
		       pr.notes, pr.submitted_at, pr.updated_at,
		       COUNT(DISTINCT pf.id) as follower_count`
	
	if userID != nil {
		query += `, EXISTS(SELECT 1 FROM prayer_followers pf2 WHERE pf2.prayer_request_id = pr.id AND pf2.user_id = $2) as is_following`
	}
	
	query += `
		FROM prayer_requests pr
		LEFT JOIN people p ON pr.person_id = p.id
		LEFT JOIN prayer_followers pf ON pr.id = pf.prayer_request_id
		WHERE pr.tenant_id = $1`

	args := []interface{}{tenantID}
	argCount := 1

	if userID != nil {
		argCount++
		args = append(args, *userID)
	}

	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND pr.status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	if filter.IsPublic != nil {
		argCount++
		query += fmt.Sprintf(" AND pr.is_public = $%d", argCount)
		args = append(args, *filter.IsPublic)
	}

	query += ` GROUP BY pr.id, pr.tenant_id, pr.person_id, person_name, pr.name, pr.email, pr.request_text,
	           pr.is_public, pr.status, pr.connection_card_id, pr.notes, pr.submitted_at, pr.updated_at`
	query += ` ORDER BY pr.submitted_at DESC`

	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	} else {
		query += " LIMIT 100"
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list prayer requests: %w", err)
	}
	defer rows.Close()

	requests := []PrayerRequest{}
	for rows.Next() {
		var pr PrayerRequest
		var followerCount int
		
		scanArgs := []interface{}{
			&pr.ID, &pr.TenantID, &pr.PersonID, &pr.PersonName,
			&pr.Name, &pr.Email, &pr.RequestText, &pr.IsPublic, &pr.Status,
			&pr.ConnectionCardID, &pr.Notes, &pr.SubmittedAt, &pr.UpdatedAt,
			&followerCount,
		}
		
		if userID != nil {
			var isFollowing bool
			scanArgs = append(scanArgs, &isFollowing)
			if err := rows.Scan(scanArgs...); err != nil {
				return nil, fmt.Errorf("failed to scan prayer request: %w", err)
			}
			pr.IsFollowing = &isFollowing
		} else {
			if err := rows.Scan(scanArgs...); err != nil {
				return nil, fmt.Errorf("failed to scan prayer request: %w", err)
			}
		}
		
		pr.FollowerCount = &followerCount
		requests = append(requests, pr)
	}

	return requests, rows.Err()
}

func (s *Service) ListPublicPrayerRequests(ctx context.Context, tenantID string, limit int) ([]PrayerRequest, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT pr.id, pr.tenant_id, pr.name, pr.request_text, pr.submitted_at
		FROM prayer_requests pr
		WHERE pr.tenant_id = $1 AND pr.is_public = TRUE AND pr.status != 'archived'
		ORDER BY pr.submitted_at DESC
		LIMIT $2`

	rows, err := s.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list public prayer requests: %w", err)
	}
	defer rows.Close()

	requests := []PrayerRequest{}
	for rows.Next() {
		var pr PrayerRequest
		if err := rows.Scan(&pr.ID, &pr.TenantID, &pr.Name, &pr.RequestText, &pr.SubmittedAt); err != nil {
			return nil, fmt.Errorf("failed to scan public prayer request: %w", err)
		}
		requests = append(requests, pr)
	}

	return requests, rows.Err()
}

func (s *Service) CreatePrayerRequest(ctx context.Context, tenantID string, input CreatePrayerRequestInput, personID *string) (*PrayerRequest, error) {
	pr := &PrayerRequest{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		PersonID:    personID,
		Name:        input.Name,
		Email:       input.Email,
		RequestText: input.RequestText,
		IsPublic:    input.IsPublic,
		Status:      "pending",
		SubmittedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO prayer_requests (id, tenant_id, person_id, name, email, request_text, is_public, status, submitted_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		pr.ID, pr.TenantID, pr.PersonID, pr.Name, pr.Email, pr.RequestText, pr.IsPublic, pr.Status, pr.SubmittedAt, pr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create prayer request: %w", err)
	}

	return pr, nil
}

func (s *Service) GetPrayerRequest(ctx context.Context, tenantID, requestID string, userID *string) (*PrayerRequest, error) {
	query := `
		SELECT pr.id, pr.tenant_id, pr.person_id, COALESCE(p.first_name || ' ' || p.last_name, '') as person_name,
		       pr.name, pr.email, pr.request_text, pr.is_public, pr.status, pr.connection_card_id,
		       pr.notes, pr.submitted_at, pr.updated_at,
		       COUNT(DISTINCT pf.id) as follower_count`
	
	if userID != nil {
		query += `, EXISTS(SELECT 1 FROM prayer_followers pf2 WHERE pf2.prayer_request_id = pr.id AND pf2.user_id = $3) as is_following`
	}
	
	query += `
		FROM prayer_requests pr
		LEFT JOIN people p ON pr.person_id = p.id
		LEFT JOIN prayer_followers pf ON pr.id = pf.prayer_request_id
		WHERE pr.tenant_id = $1 AND pr.id = $2
		GROUP BY pr.id, pr.tenant_id, pr.person_id, person_name, pr.name, pr.email, pr.request_text,
		         pr.is_public, pr.status, pr.connection_card_id, pr.notes, pr.submitted_at, pr.updated_at`

	var pr PrayerRequest
	var followerCount int
	
	args := []interface{}{tenantID, requestID}
	if userID != nil {
		args = append(args, *userID)
	}
	
	scanArgs := []interface{}{
		&pr.ID, &pr.TenantID, &pr.PersonID, &pr.PersonName,
		&pr.Name, &pr.Email, &pr.RequestText, &pr.IsPublic, &pr.Status,
		&pr.ConnectionCardID, &pr.Notes, &pr.SubmittedAt, &pr.UpdatedAt,
		&followerCount,
	}
	
	if userID != nil {
		var isFollowing bool
		scanArgs = append(scanArgs, &isFollowing)
		if err := s.db.QueryRow(ctx, query, args...).Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to get prayer request: %w", err)
		}
		pr.IsFollowing = &isFollowing
	} else {
		if err := s.db.QueryRow(ctx, query, args...).Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to get prayer request: %w", err)
		}
	}
	
	pr.FollowerCount = &followerCount
	return &pr, nil
}

func (s *Service) UpdatePrayerRequest(ctx context.Context, tenantID, requestID string, input UpdatePrayerRequestInput) (*PrayerRequest, error) {
	updates := []string{}
	args := []interface{}{tenantID, requestID}
	argCount := 2

	if input.Status != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("status = $%d", argCount))
		args = append(args, input.Status)
	}

	if input.Notes != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("notes = $%d", argCount))
		args = append(args, *input.Notes)
	}

	if len(updates) == 0 {
		return s.GetPrayerRequest(ctx, tenantID, requestID, nil)
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf(`UPDATE prayer_requests SET %s WHERE tenant_id = $1 AND id = $2`, strings.Join(updates, ", "))

	_, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update prayer request: %w", err)
	}

	return s.GetPrayerRequest(ctx, tenantID, requestID, nil)
}

func (s *Service) FollowPrayerRequest(ctx context.Context, tenantID, requestID, userID string) error {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM prayer_requests WHERE id = $1 AND tenant_id = $2)`,
		requestID, tenantID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify prayer request: %w", err)
	}
	if !exists {
		return fmt.Errorf("prayer request not found")
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO prayer_followers (id, prayer_request_id, user_id, followed_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (prayer_request_id, user_id) DO NOTHING`,
		uuid.New().String(), requestID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to follow prayer request: %w", err)
	}

	return nil
}

func (s *Service) UnfollowPrayerRequest(ctx context.Context, tenantID, requestID, userID string) error {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM prayer_requests WHERE id = $1 AND tenant_id = $2)`,
		requestID, tenantID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify prayer request: %w", err)
	}
	if !exists {
		return fmt.Errorf("prayer request not found")
	}

	_, err = s.db.Exec(ctx,
		`DELETE FROM prayer_followers WHERE prayer_request_id = $1 AND user_id = $2`,
		requestID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to unfollow prayer request: %w", err)
	}

	return nil
}

func (s *Service) ListFollowers(ctx context.Context, tenantID, requestID string) ([]PrayerFollower, error) {
	query := `
		SELECT pf.id, pf.prayer_request_id, pf.user_id, u.name as user_name, pf.followed_at
		FROM prayer_followers pf
		JOIN users u ON pf.user_id = u.id
		JOIN prayer_requests pr ON pf.prayer_request_id = pr.id
		WHERE pr.tenant_id = $1 AND pf.prayer_request_id = $2
		ORDER BY pf.followed_at DESC`

	rows, err := s.db.Query(ctx, query, tenantID, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to list followers: %w", err)
	}
	defer rows.Close()

	followers := []PrayerFollower{}
	for rows.Next() {
		var pf PrayerFollower
		if err := rows.Scan(&pf.ID, &pf.PrayerRequestID, &pf.UserID, &pf.UserName, &pf.FollowedAt); err != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", err)
		}
		followers = append(followers, pf)
	}

	return followers, rows.Err()
}

func (s *Service) ImportFromConnectionCard(ctx context.Context, tenantID, connectionCardID string) (*PrayerRequest, error) {
	var name, email, prayerRequestText string
	var personID *string
	
	err := s.db.QueryRow(ctx,
		`SELECT first_name || ' ' || COALESCE(last_name, ''), COALESCE(email, ''), prayer_request, person_id
		 FROM connection_cards
		 WHERE id = $1 AND tenant_id = $2 AND prayer_request IS NOT NULL AND prayer_request != ''`,
		connectionCardID, tenantID,
	).Scan(&name, &email, &prayerRequestText, &personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection card: %w", err)
	}

	var existingID string
	err = s.db.QueryRow(ctx,
		`SELECT id FROM prayer_requests WHERE connection_card_id = $1`,
		connectionCardID,
	).Scan(&existingID)
	if err == nil {
		return s.GetPrayerRequest(ctx, tenantID, existingID, nil)
	}

	pr := &PrayerRequest{
		ID:               uuid.New().String(),
		TenantID:         tenantID,
		PersonID:         personID,
		Name:             name,
		Email:            &email,
		RequestText:      prayerRequestText,
		IsPublic:         false,
		Status:           "pending",
		ConnectionCardID: &connectionCardID,
		SubmittedAt:      time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO prayer_requests (id, tenant_id, person_id, name, email, request_text, is_public, status, connection_card_id, submitted_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		pr.ID, pr.TenantID, pr.PersonID, pr.Name, pr.Email, pr.RequestText, pr.IsPublic, pr.Status, pr.ConnectionCardID, pr.SubmittedAt, pr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create prayer request from connection card: %w", err)
	}

	return pr, nil
}
