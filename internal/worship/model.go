package worship

import "time"

type ServicePlan struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	ServiceID string    `json:"service_id"`
	CreatedBy string    `json:"created_by"`
	Notes     string    `json:"notes,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Items     []ServicePlanItem `json:"items,omitempty"`
}

type ServicePlanItem struct {
	ID              string    `json:"id"`
	PlanID          string    `json:"plan_id"`
	ItemOrder       int       `json:"item_order"`
	ItemType        string    `json:"item_type"`
	Title           string    `json:"title"`
	DurationMinutes *int      `json:"duration_minutes,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	SongID          *string   `json:"song_id,omitempty"`
	AssignedTo      *string   `json:"assigned_to,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// Populated fields
	SongTitle       string `json:"song_title,omitempty"`
	AssignedToName  string `json:"assigned_to_name,omitempty"`
}
