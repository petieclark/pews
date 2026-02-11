package notification

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	TypeInfo    NotificationType = "info"
	TypeWarning NotificationType = "warning"
	TypeSuccess NotificationType = "success"
)

// Notification represents an in-app notification
type Notification struct {
	ID        string           `json:"id"`
	TenantID  string           `json:"tenant_id"`
	UserID    string           `json:"user_id"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Type      NotificationType `json:"type"`
	Read      bool             `json:"read"`
	Link      *string          `json:"link,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID  string           `json:"user_id"`
	Title   string           `json:"title"`
	Message string           `json:"message"`
	Type    NotificationType `json:"type"`
	Link    *string          `json:"link,omitempty"`
}

// ListNotificationsParams represents query parameters for listing notifications
type ListNotificationsParams struct {
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	Unread bool `json:"unread"`
}

// NotificationList represents a paginated list of notifications
type NotificationList struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
}

// UnreadCountResponse represents the count of unread notifications
type UnreadCountResponse struct {
	Count int `json:"count"`
}
