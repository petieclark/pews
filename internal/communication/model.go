package communication

import (
	"encoding/json"
	"time"
)

// MessageTemplate represents a reusable message template
type MessageTemplate struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject,omitempty"`
	Body      string    `json:"body"`
	Channel   string    `json:"channel"` // email, sms
	Category  string    `json:"category,omitempty"`
	Variables string    `json:"variables,omitempty"` // comma-separated
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Campaign represents a bulk messaging campaign
type Campaign struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	Name           string     `json:"name"`
	TemplateID     *string    `json:"template_id,omitempty"`
	Channel        string     `json:"channel"`
	Subject        string     `json:"subject,omitempty"`
	Body           string     `json:"body"`
	Status         string     `json:"status"` // draft, scheduled, sending, sent, failed
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	RecipientCount int        `json:"recipient_count"`
	OpenedCount    int        `json:"opened_count"`
	ClickedCount   int        `json:"clicked_count"`
	TargetType     string     `json:"target_type"` // all, tag, group, list, manual
	TargetID       string     `json:"target_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CampaignRecipient represents a recipient of a campaign
type CampaignRecipient struct {
	ID         string     `json:"id"`
	CampaignID string     `json:"campaign_id"`
	PersonID   string     `json:"person_id"`
	Status     string     `json:"status"` // pending, sent, delivered, opened, clicked, bounced, failed
	SentAt     *time.Time `json:"sent_at,omitempty"`
	OpenedAt   *time.Time `json:"opened_at,omitempty"`
	ClickedAt  *time.Time `json:"clicked_at,omitempty"`
}

// Journey represents an automated message journey
type Journey struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	TriggerType  string    `json:"trigger_type"`  // first_visit, tag_added, group_joined, manual, checkin_first_time
	TriggerValue string    `json:"trigger_value,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Steps        []JourneyStep `json:"steps,omitempty"`
	EnrollmentCount int    `json:"enrollment_count,omitempty"`
}

// JourneyStep represents a step in a journey
type JourneyStep struct {
	ID         string          `json:"id"`
	JourneyID  string          `json:"journey_id"`
	Position   int             `json:"position"`
	StepType   string          `json:"step_type"` // send_email, send_sms, wait, add_tag, add_to_group
	DelayDays  int             `json:"delay_days"`
	DelayHours int             `json:"delay_hours"`
	TemplateID *string         `json:"template_id,omitempty"`
	Config     json.RawMessage `json:"config"`
	CreatedAt  time.Time       `json:"created_at"`
}

// JourneyEnrollment represents a person enrolled in a journey
type JourneyEnrollment struct {
	ID          string     `json:"id"`
	JourneyID   string     `json:"journey_id"`
	PersonID    string     `json:"person_id"`
	CurrentStep int        `json:"current_step"`
	Status      string     `json:"status"` // active, completed, paused, exited
	EnrolledAt  time.Time  `json:"enrolled_at"`
	NextStepAt  *time.Time `json:"next_step_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	PersonName  string     `json:"person_name,omitempty"` // for listing
}

// ConnectionCard represents a digital connection card submission
type ConnectionCard struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name,omitempty"`
	Email         string    `json:"email,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	IsFirstVisit  bool      `json:"is_first_visit"`
	HowHeard      string    `json:"how_heard,omitempty"`
	PrayerRequest string    `json:"prayer_request,omitempty"`
	InterestedIn  string    `json:"interested_in,omitempty"`
	SubmittedAt   time.Time `json:"submitted_at"`
	Processed     bool      `json:"processed"`
	PersonID      *string   `json:"person_id,omitempty"`
}

// CommunicationStats represents communication statistics
type CommunicationStats struct {
	EmailsSentThisMonth int     `json:"emails_sent_this_month"`
	SMSSentThisMonth    int     `json:"sms_sent_this_month"`
	OpenRate            float64 `json:"open_rate"`
	ActiveJourneys      int     `json:"active_journeys"`
	UnprocessedCards    int     `json:"unprocessed_cards"`
}
