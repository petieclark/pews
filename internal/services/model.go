package services

import "time"

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
	ServiceDate   string         `json:"service_date"` // DATE format YYYY-MM-DD
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
	LastUsed   *string   `json:"last_used,omitempty"`
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
	ID        string `json:"id"`
	ServiceID string `json:"service_id"`
	PersonID  string `json:"person_id"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	Notes     string `json:"notes,omitempty"`
	// Optionally include person details
	PersonFirstName string `json:"person_first_name,omitempty"`
	PersonLastName  string `json:"person_last_name,omitempty"`
}
