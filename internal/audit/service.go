package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// Log creates an immutable audit log entry
func (s *Service) Log(ctx context.Context, tenantID, userID *string, ipAddress, userAgent *string, entry LogEntry) error {
	var oldValueJSON, newValueJSON, metadataJSON []byte
	var err error

	if entry.OldValue != nil {
		oldValueJSON, err = json.Marshal(entry.OldValue)
		if err != nil {
			return fmt.Errorf("failed to marshal old value: %w", err)
		}
	}

	if entry.NewValue != nil {
		newValueJSON, err = json.Marshal(entry.NewValue)
		if err != nil {
			return fmt.Errorf("failed to marshal new value: %w", err)
		}
	}

	if entry.Metadata != nil {
		metadataJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO audit_logs (tenant_id, user_id, action, entity_type, entity_id, ip_address, user_agent, old_value, new_value, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = s.db.ExecContext(ctx, query, tenantID, userID, entry.Action, entry.EntityType, entry.EntityID, ipAddress, userAgent, oldValueJSON, newValueJSON, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetLogs retrieves paginated audit logs with optional filters
func (s *Service) GetLogs(ctx context.Context, tenantID string, userID *string, action *string, entityType *string, limit, offset int) ([]AuditLog, int, error) {
	logs := []AuditLog{}
	var totalCount int

	// Build WHERE clause dynamically
	whereClause := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argCount := 1

	if userID != nil && *userID != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *userID)
	}

	if action != nil && *action != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, *action)
	}

	if entityType != nil && *entityType != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND entity_type = $%d", argCount)
		args = append(args, *entityType)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	err := s.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get paginated results
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, timestamp, action, entity_type, entity_id, 
		       ip_address, user_agent, old_value, new_value, metadata
		FROM audit_logs
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitArg, offsetArg)

	args = append(args, limit, offset)

	err = s.db.SelectContext(ctx, &logs, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	return logs, totalCount, nil
}

// GetLogsByUser retrieves audit logs for a specific user
func (s *Service) GetLogsByUser(ctx context.Context, tenantID, userID string, limit, offset int) ([]AuditLog, int, error) {
	return s.GetLogs(ctx, tenantID, &userID, nil, nil, limit, offset)
}

// LogFailedLogin records a failed login attempt
func (s *Service) LogFailedLogin(ctx context.Context, tenantID, email string, ipAddress, userAgent *string) error {
	query := `
		INSERT INTO failed_login_attempts (tenant_id, email, ip_address, user_agent)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.ExecContext(ctx, query, tenantID, email, ipAddress, userAgent)
	if err != nil {
		return fmt.Errorf("failed to log failed login: %w", err)
	}

	return nil
}

// CreateSession creates a new user session
func (s *Service) CreateSession(ctx context.Context, tenantID, userID string, ipAddress, userAgent *string) (string, error) {
	var sessionID string
	query := `
		INSERT INTO user_sessions (tenant_id, user_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := s.db.GetContext(ctx, &sessionID, query, tenantID, userID, ipAddress, userAgent)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// UpdateSessionActivity updates the last activity timestamp
func (s *Service) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	query := `UPDATE user_sessions SET last_activity = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, sessionID)
	return err
}

// EndSession marks a session as inactive
func (s *Service) EndSession(ctx context.Context, sessionID string) error {
	query := `UPDATE user_sessions SET is_active = FALSE WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, sessionID)
	return err
}

// GetSecurityDashboard retrieves security dashboard data
func (s *Service) GetSecurityDashboard(ctx context.Context, tenantID string) (*SecurityDashboard, error) {
	dashboard := &SecurityDashboard{}

	// Active sessions count
	err := s.db.GetContext(ctx, &dashboard.ActiveSessionsCount, `
		SELECT COUNT(*) FROM user_sessions 
		WHERE tenant_id = $1 AND is_active = TRUE
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}

	// Failed logins in last 24h
	err = s.db.GetContext(ctx, &dashboard.FailedLoginsLast24h, `
		SELECT COUNT(*) FROM failed_login_attempts
		WHERE tenant_id = $1 AND attempted_at > NOW() - INTERVAL '24 hours'
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count failed logins: %w", err)
	}

	// Users without 2FA (totp_secret is NULL)
	err = s.db.GetContext(ctx, &dashboard.UsersWithout2FA, `
		SELECT COUNT(*) FROM users
		WHERE tenant_id = $1 AND (totp_secret IS NULL OR totp_secret = '')
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count users without 2FA: %w", err)
	}

	// Recent failed logins (last 10)
	dashboard.RecentFailedLogins = []FailedLoginAttempt{}
	err = s.db.SelectContext(ctx, &dashboard.RecentFailedLogins, `
		SELECT id, tenant_id, email, ip_address, user_agent, attempted_at
		FROM failed_login_attempts
		WHERE tenant_id = $1
		ORDER BY attempted_at DESC
		LIMIT 10
	`, tenantID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch recent failed logins: %w", err)
	}

	// Unusual activities (logins from new IPs)
	dashboard.UnusualActivities = []UnusualActivity{}
	err = s.db.SelectContext(ctx, &dashboard.UnusualActivities, `
		WITH user_ips AS (
			SELECT DISTINCT user_id, ip_address
			FROM audit_logs
			WHERE tenant_id = $1 AND action = 'auth.login' AND timestamp < NOW() - INTERVAL '7 days'
		)
		SELECT DISTINCT
			al.user_id,
			u.email,
			al.action,
			al.ip_address,
			al.timestamp
		FROM audit_logs al
		JOIN users u ON al.user_id = u.id
		LEFT JOIN user_ips ui ON al.user_id = ui.user_id AND al.ip_address = ui.ip_address
		WHERE al.tenant_id = $1
			AND al.action = 'auth.login'
			AND al.timestamp > NOW() - INTERVAL '7 days'
			AND ui.user_id IS NULL
		ORDER BY al.timestamp DESC
		LIMIT 10
	`, tenantID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch unusual activities: %w", err)
	}

	// Add "reason" to unusual activities
	for i := range dashboard.UnusualActivities {
		dashboard.UnusualActivities[i].Reason = "Login from new IP address"
	}

	// User password changes
	dashboard.UserPasswordChanges = []UserPasswordChange{}
	err = s.db.SelectContext(ctx, &dashboard.UserPasswordChanges, `
		SELECT 
			id as user_id,
			email,
			password_changed_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY password_changed_at DESC NULLS LAST
	`, tenantID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch password changes: %w", err)
	}

	// Calculate days since last password change
	now := time.Now()
	for i := range dashboard.UserPasswordChanges {
		if dashboard.UserPasswordChanges[i].PasswordChangedAt != nil {
			days := int(now.Sub(*dashboard.UserPasswordChanges[i].PasswordChangedAt).Hours() / 24)
			dashboard.UserPasswordChanges[i].DaysSinceChange = &days
		}
	}

	return dashboard, nil
}
