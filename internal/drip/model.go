package drip

import "time"

// Campaign represents a drip campaign
type Campaign struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	TriggerEvent string    `json:"trigger_event"` // new_member, connection_card, first_visit
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Steps        []Step    `json:"steps,omitempty"`
	EnrollmentCount int    `json:"enrollment_count,omitempty"`
}

// Step represents a step in a drip campaign
type Step struct {
	ID         string    `json:"id"`
	CampaignID string    `json:"campaign_id"`
	StepOrder  int       `json:"step_order"`
	DelayDays  int       `json:"delay_days"`
	ActionType string    `json:"action_type"` // email, sms, follow_up
	Subject    string    `json:"subject,omitempty"`
	Body       string    `json:"body"`
	TemplateID *string   `json:"template_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Enrollment represents a person enrolled in a drip campaign
type Enrollment struct {
	ID          string     `json:"id"`
	CampaignID  string     `json:"campaign_id"`
	PersonID    string     `json:"person_id"`
	Status      string     `json:"status"` // active, completed, paused, cancelled
	EnrolledAt  time.Time  `json:"enrolled_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	PersonName  string     `json:"person_name,omitempty"` // for listing
	PersonEmail string     `json:"person_email,omitempty"`
}

// StepExecution represents the execution of a drip step
type StepExecution struct {
	ID           string     `json:"id"`
	EnrollmentID string     `json:"enrollment_id"`
	StepID       string     `json:"step_id"`
	Status       string     `json:"status"` // pending, sent, failed
	ScheduledAt  time.Time  `json:"scheduled_at"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// CreateCampaignRequest represents a request to create a campaign
type CreateCampaignRequest struct {
	Name         string `json:"name" binding:"required"`
	TriggerEvent string `json:"trigger_event" binding:"required,oneof=new_member connection_card first_visit"`
	IsActive     bool   `json:"is_active"`
}

// UpdateCampaignRequest represents a request to update a campaign
type UpdateCampaignRequest struct {
	Name         string `json:"name"`
	TriggerEvent string `json:"trigger_event" binding:"omitempty,oneof=new_member connection_card first_visit"`
	IsActive     *bool  `json:"is_active"`
}

// CreateStepRequest represents a request to create a step
type CreateStepRequest struct {
	StepOrder  int     `json:"step_order" binding:"required"`
	DelayDays  int     `json:"delay_days" binding:"min=0"`
	ActionType string  `json:"action_type" binding:"required,oneof=email sms follow_up"`
	Subject    string  `json:"subject"`
	Body       string  `json:"body" binding:"required"`
	TemplateID *string `json:"template_id"`
}

// UpdateStepRequest represents a request to update a step
type UpdateStepRequest struct {
	StepOrder  *int    `json:"step_order"`
	DelayDays  *int    `json:"delay_days" binding:"omitempty,min=0"`
	ActionType string  `json:"action_type" binding:"omitempty,oneof=email sms follow_up"`
	Subject    string  `json:"subject"`
	Body       string  `json:"body"`
	TemplateID *string `json:"template_id"`
}
