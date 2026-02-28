package notification

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InAppService handles in-app database notifications (separate from email service)
type InAppService struct {
	db *pgxpool.Pool
}

func NewInAppService(db *pgxpool.Pool) *InAppService {
	return &InAppService{db: db}
}

// List returns paginated list of notifications for a user
func (s *InAppService) List(ctx context.Context, tenantID, userID string, params ListNotificationsParams) (*NotificationList, error) {
	// Get total count
	var total int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2`,
		tenantID, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification count: %w", err)
	}

	// Build query with filters
	query := `
		SELECT id, tenant_id, user_id, title, message, type, read, link, created_at
		FROM notifications 
		WHERE tenant_id = $1 AND user_id = $2`
	
	args := []interface{}{tenantID, userID}
	
	if params.Unread {
		query += " AND NOT read"
	}
	
	query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	args = append(args, params.Limit, params.Offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var link *string
		err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Title, &n.Message, 
			&n.Type, &n.Read, &link, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		n.Link = link
		notifications = append(notifications, n)
	}

	return &NotificationList{
		Notifications: notifications,
		Total:         total,
		Limit:         params.Limit,
		Offset:        params.Offset,
	}, nil
}

// MarkAsRead marks a single notification as read
func (s *InAppService) MarkAsRead(ctx context.Context, tenantID, userID, notificationID string) error {
	result, err := s.db.Exec(ctx, `
		UPDATE notifications SET read = true WHERE id = $1 AND tenant_id = $2 AND user_id = $3`,
		notificationID, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *InAppService) MarkAllAsRead(ctx context.Context, tenantID, userID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE notifications SET read = true WHERE id IN (
			SELECT id FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND NOT read
		)`, tenantID, userID)
	return err
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *InAppService) GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications 
		WHERE tenant_id = $1 AND user_id = $2 AND NOT read`,
		tenantID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}
