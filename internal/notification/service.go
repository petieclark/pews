package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Create creates a new notification for a specific user
func (s *Service) Create(ctx context.Context, tenantID string, req CreateNotificationRequest) (*Notification, error) {
	notif := &Notification{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		UserID:    req.UserID,
		Title:     req.Title,
		Message:   req.Message,
		Type:      req.Type,
		Read:      false,
		Link:      req.Link,
		CreatedAt: time.Now(),
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO notifications (id, tenant_id, user_id, title, message, type, read, link, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		notif.ID, notif.TenantID, notif.UserID, notif.Title, notif.Message, notif.Type, notif.Read, notif.Link, notif.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notif, nil
}

// CreateForAllAdmins creates a notification for all admin users in a tenant
func (s *Service) CreateForAllAdmins(ctx context.Context, tenantID, title, message string, notifType NotificationType, link *string) error {
	// Get all admin users for this tenant
	rows, err := s.db.Query(ctx,
		`SELECT id FROM users WHERE tenant_id = $1 AND role = 'admin'`,
		tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch admin users: %w", err)
	}
	defer rows.Close()

	var adminIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		adminIDs = append(adminIDs, id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Create notification for each admin
	for _, adminID := range adminIDs {
		_, err := s.Create(ctx, tenantID, CreateNotificationRequest{
			UserID:  adminID,
			Title:   title,
			Message: message,
			Type:    notifType,
			Link:    link,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// List returns paginated notifications for a user
func (s *Service) List(ctx context.Context, tenantID, userID string, params ListNotificationsParams) (*NotificationList, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	// Build query
	query := `SELECT id, tenant_id, user_id, title, message, type, read, link, created_at 
	          FROM notifications 
	          WHERE tenant_id = $1 AND user_id = $2`
	countQuery := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2`

	args := []interface{}{tenantID, userID}
	if params.Unread {
		query += ` AND read = FALSE`
		countQuery += ` AND read = FALSE`
	}

	query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	args = append(args, params.Limit, params.Offset)

	// Get total count
	var total int
	err := s.db.QueryRow(ctx, countQuery, tenantID, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get notifications
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.Read, &n.Link, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &NotificationList{
		Notifications: notifications,
		Total:         total,
		Limit:         params.Limit,
		Offset:        params.Offset,
	}, nil
}

// MarkAsRead marks a specific notification as read
func (s *Service) MarkAsRead(ctx context.Context, tenantID, userID, notificationID string) error {
	result, err := s.db.Exec(ctx,
		`UPDATE notifications SET read = TRUE 
		 WHERE id = $1 AND tenant_id = $2 AND user_id = $3`,
		notificationID, tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *Service) MarkAllAsRead(ctx context.Context, tenantID, userID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE notifications SET read = TRUE 
		 WHERE tenant_id = $1 AND user_id = $2 AND read = FALSE`,
		tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *Service) GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications 
		 WHERE tenant_id = $1 AND user_id = $2 AND read = FALSE`,
		tenantID, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}
