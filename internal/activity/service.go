package activity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// LogActivity records an activity log entry
func (s *Service) LogActivity(ctx context.Context, tenantID, action, entityType string, userID, entityID, ipAddress *string, details interface{}) error {
	var detailsJSON []byte
	var err error
	
	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	query := `
		INSERT INTO activity_log (tenant_id, user_id, action, entity_type, entity_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = s.db.Exec(ctx, query, tenantID, userID, action, entityType, entityID, detailsJSON, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// ListActivity returns activity logs with optional filtering
func (s *Service) ListActivity(ctx context.Context, params ListActivityParams) ([]ActivityLog, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 50
	}
	offset := (params.Page - 1) * params.Limit

	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", params.TenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Build query with filters
	sqlQuery := `
		SELECT 
			al.id, 
			al.tenant_id, 
			al.user_id, 
			COALESCE(u.email, '') as user_email,
			al.action, 
			al.entity_type, 
			al.entity_id, 
			al.details, 
			al.ip_address, 
			al.created_at
		FROM activity_log al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM activity_log WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	// Add filters
	if params.EntityType != "" {
		sqlQuery += fmt.Sprintf(" AND al.entity_type = $%d", argPos)
		countQuery += fmt.Sprintf(" AND entity_type = $%d", argPos)
		args = append(args, params.EntityType)
		argPos++
	}

	if params.UserID != "" {
		sqlQuery += fmt.Sprintf(" AND al.user_id = $%d", argPos)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, params.UserID)
		argPos++
	}

	if params.StartDate != nil {
		sqlQuery += fmt.Sprintf(" AND al.created_at >= $%d", argPos)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, params.StartDate)
		argPos++
	}

	if params.EndDate != nil {
		sqlQuery += fmt.Sprintf(" AND al.created_at <= $%d", argPos)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, params.EndDate)
		argPos++
	}

	// Get total count
	var total int
	err = s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count activity logs: %w", err)
	}

	// Add pagination and ordering
	sqlQuery += fmt.Sprintf(" ORDER BY al.created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, params.Limit, offset)

	// Execute query
	rows, err := s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query activity logs: %w", err)
	}
	defer rows.Close()

	logs := []ActivityLog{}
	for rows.Next() {
		var log ActivityLog
		err := rows.Scan(
			&log.ID,
			&log.TenantID,
			&log.UserID,
			&log.UserEmail,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.Details,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan activity log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
