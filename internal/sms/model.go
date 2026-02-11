package sms

import "time"

// Message represents an SMS message
type Message struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	ToPhone      string     `json:"to_phone"`
	FromPhone    string     `json:"from_phone"`
	Body         string     `json:"body"`
	Status       string     `json:"status"` // queued, sent, delivered, failed
	TwilioSID    string     `json:"twilio_sid,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Template represents an SMS template with merge fields
type Template struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	Variables string    `json:"variables,omitempty"` // comma-separated
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Settings represents tenant-specific Twilio settings
type Settings struct {
	ID                       string     `json:"id"`
	TenantID                 string     `json:"tenant_id"`
	TwilioAccountSID        string     `json:"twilio_account_sid,omitempty"`
	TwilioAuthTokenEncrypted string     `json:"-"` // never expose in JSON
	TwilioFromNumber        string     `json:"twilio_from_number,omitempty"`
	IsEnabled               bool       `json:"is_enabled"`
	LastTestedAt            *time.Time `json:"last_tested_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// SendRequest represents a request to send SMS to one person
type SendRequest struct {
	ToPhone     string            `json:"to_phone"`
	Body        string            `json:"body"`
	TemplateID  string            `json:"template_id,omitempty"`
	MergeFields map[string]string `json:"merge_fields,omitempty"`
}

// BulkSendRequest represents a request to send SMS to multiple recipients
type BulkSendRequest struct {
	Body        string            `json:"body"`
	TemplateID  string            `json:"template_id,omitempty"`
	MergeFields map[string]string `json:"merge_fields,omitempty"` // common fields for all
	TargetType  string            `json:"target_type"`            // person_ids, group_id, all
	TargetIDs   []string          `json:"target_ids,omitempty"`   // list of person IDs or single group ID
}

// HistoryFilter represents filters for message history
type HistoryFilter struct {
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}
