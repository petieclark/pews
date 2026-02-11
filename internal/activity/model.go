package activity

import (
	"encoding/json"
	"time"
)

type ActivityLog struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenant_id"`
	UserID     *string         `json:"user_id,omitempty"`
	UserEmail  string          `json:"user_email,omitempty"`
	UserName   string          `json:"user_name,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ListActivityParams struct {
	TenantID   string
	EntityType string
	UserID     string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	Limit      int
}
