package calendar

import "time"

type Event struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	AllDay      bool      `json:"all_day"`
	Recurring   string    `json:"recurring"`   // none, weekly, monthly
	EventType   string    `json:"event_type"`  // service, meeting, class, social, outreach, other
	Color       string    `json:"color"`
	RoomID      *string   `json:"room_id,omitempty"`
	RoomName    *string   `json:"room_name,omitempty"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ServiceID       *string   `json:"service_id,omitempty"`
	AttendanceCount *int      `json:"attendance_count,omitempty"`
}

// EventTypeColors maps event types to their default colors
var EventTypeColors = map[string]string{
	"service":  "#4A8B8C", // Teal
	"meeting":  "#1B3A4B", // Navy
	"class":    "#8B5CF6", // Purple
	"social":   "#F59E0B", // Amber
	"outreach": "#10B981", // Emerald
	"other":    "#6B7280", // Gray
}
