package checkins

import (
	"time"
)

type Station struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Event struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	EventDate  string    `json:"event_date"`
	ServiceID  *string   `json:"service_id,omitempty"`
	StationID  *string   `json:"station_id,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	CheckinCount int     `json:"checkin_count,omitempty"`
}

type Checkin struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	EventID      string     `json:"event_id"`
	PersonID     string     `json:"person_id"`
	StationID    *string    `json:"station_id,omitempty"`
	FirstTime    bool       `json:"first_time"`
	CheckedInAt  time.Time  `json:"checked_in_at"`
	CheckedOutAt *time.Time `json:"checked_out_at,omitempty"`
	Notes        string     `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	// Joined fields
	PersonName   string     `json:"person_name,omitempty"`
	PersonEmail  string     `json:"person_email,omitempty"`
	StationName  string     `json:"station_name,omitempty"`
}

type MedicalAlert struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	PersonID    string    `json:"person_id"`
	AlertType   string    `json:"alert_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AuthorizedPickup struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	ChildID        string    `json:"child_id"`
	PickupPersonID string    `json:"pickup_person_id"`
	Relationship   string    `json:"relationship"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Joined fields
	PickupPersonName string `json:"pickup_person_name,omitempty"`
}

type CheckinResult struct {
	Checkin       *Checkin       `json:"checkin"`
	FirstTime     bool           `json:"first_time"`
	MedicalAlerts []MedicalAlert `json:"medical_alerts,omitempty"`
}

type Stats struct {
	TotalCheckins  int            `json:"total_checkins"`
	FirstTimers    int            `json:"first_timers"`
	ByStation      []StationStat  `json:"by_station"`
}

type StationStat struct {
	StationID   string `json:"station_id"`
	StationName string `json:"station_name"`
	Count       int    `json:"count"`
}

type PersonSearchResult struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
}
