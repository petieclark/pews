package services

import (
	"encoding/json"
	"time"
)

type ServiceType struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	DefaultTime string    `json:"default_time,omitempty"`
	DefaultDay  string    `json:"default_day,omitempty"`
	Color       string    `json:"color"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChurchService struct {
	ID            string         `json:"id"`
	TenantID      string         `json:"tenant_id"`
	ServiceTypeID string         `json:"service_type_id"`
	Name          string         `json:"name,omitempty"`
	ServiceDate   time.Time      `json:"service_date"`
	ServiceTime   string         `json:"service_time,omitempty"`
	Notes         string         `json:"notes,omitempty"`
	Status        string         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	ServiceType   *ServiceType   `json:"service_type,omitempty"`
	Items         []ServiceItem  `json:"items,omitempty"`
	Team          []ServiceTeam  `json:"team,omitempty"`
}

type Song struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist,omitempty"`
	DefaultKey string    `json:"default_key,omitempty"`
	Tempo      int       `json:"tempo,omitempty"`
	CCLINumber string    `json:"ccli_number,omitempty"`
	Lyrics     string    `json:"lyrics,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	Tags       string    `json:"tags,omitempty"`
	LastUsed   *time.Time `json:"last_used,omitempty"`
	TimesUsed  int       `json:"times_used"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ServiceItem struct {
	ID              string  `json:"id"`
	ServiceID       string  `json:"service_id"`
	ItemType        string  `json:"item_type"`
	Title           string  `json:"title"`
	SongID          *string `json:"song_id,omitempty"`
	SongKey         string  `json:"song_key,omitempty"`
	Position        int     `json:"position"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	AssignedTo      string  `json:"assigned_to,omitempty"`
	Song            *Song   `json:"song,omitempty"`
}

type ServiceTeam struct {
	ID               string     `json:"id"`
	ServiceID        string     `json:"service_id"`
	PersonID         string     `json:"person_id"`
	TeamID           *string    `json:"team_id,omitempty"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	Notes            string     `json:"notes,omitempty"`
	NotifiedAt       *time.Time `json:"notified_at,omitempty"`
	RespondedAt      *time.Time `json:"responded_at,omitempty"`
	NotificationSent bool       `json:"notification_sent"`
	// Optionally include person details
	PersonFirstName string `json:"person_first_name,omitempty"`
	PersonLastName  string `json:"person_last_name,omitempty"`
	PersonEmail     string `json:"person_email,omitempty"`
	TeamName        string `json:"team_name,omitempty"`
}

type VolunteerTeam struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Optionally include member count
	MemberCount int `json:"member_count,omitempty"`
}

type TeamMember struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	PersonID    string    `json:"person_id"`
	Role        string    `json:"role,omitempty"`
	IsActive    bool      `json:"is_active"`
	AddedAt     time.Time `json:"added_at"`
	// Person details
	PersonFirstName string `json:"person_first_name,omitempty"`
	PersonLastName  string `json:"person_last_name,omitempty"`
	PersonEmail     string `json:"person_email,omitempty"`
	// Team details
	TeamName  string `json:"team_name,omitempty"`
	TeamColor string `json:"team_color,omitempty"`
}

type VolunteerAvailability struct {
	ID        string     `json:"id"`
	PersonID  string     `json:"person_id"`
	TeamID    *string    `json:"team_id,omitempty"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
	Reason    string     `json:"reason,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	// Person details
	PersonFirstName string `json:"person_first_name,omitempty"`
	PersonLastName  string `json:"person_last_name,omitempty"`
}

type ServiceTemplate struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	TemplateData json.RawMessage `json:"template_data"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type SchedulingConflict struct {
	PersonID     string `json:"person_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	ServiceCount int    `json:"service_count"`
}

type AutoScheduleRequest struct {
	ServiceID string `json:"service_id"`
	TeamID    string `json:"team_id"`
	Role      string `json:"role"`
}

type SongUsage struct {
	ServiceID   string    `json:"service_id"`
	ServiceName string    `json:"service_name"`
	ServiceDate time.Time `json:"service_date"`
	ServiceTime string    `json:"service_time"`
	SongKey     string    `json:"song_key,omitempty"`
	Position    int       `json:"position"`
}

type SongAttachment struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	SongID       string     `json:"song_id"`
	Filename     string     `json:"filename"`
	OriginalName string     `json:"original_name"`
	ContentType  string     `json:"content_type"`
	FileData     []byte     `json:"-"` // Excluded from JSON
	FileSize     int        `json:"file_size"`
	UploadedBy   *string    `json:"uploaded_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
