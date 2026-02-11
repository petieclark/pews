package rooms

import (
	"time"
)

type Room struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Capacity    *int     `json:"capacity"`
	Description string   `json:"description"`
	Color       string   `json:"color"`
	Amenities   []string `json:"amenities"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomBooking struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	RoomID      string     `json:"room_id"`
	RoomName    string     `json:"room_name,omitempty"`
	EventName   string     `json:"event_name"`
	BookedBy    *string    `json:"booked_by"`
	BookedByName *string   `json:"booked_by_name,omitempty"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	Recurring   *string    `json:"recurring,omitempty"`
	Status      string     `json:"status"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type RoomAvailability struct {
	RoomID    string `json:"room_id"`
	RoomName  string `json:"room_name"`
	Available bool   `json:"available"`
	Capacity  *int   `json:"capacity"`
	Color     string `json:"color"`
}

type BookingConflict struct {
	HasConflict bool          `json:"has_conflict"`
	Conflicts   []RoomBooking `json:"conflicts,omitempty"`
}
