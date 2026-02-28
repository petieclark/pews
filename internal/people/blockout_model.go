package people

import (
	"time"
)

// VolunteerBlockout represents a volunteer's unavailable dates (blockout/blacklist)
type VolunteerBlockout struct {
	ID          string     `json:"id"`
	PersonID    string     `json:"_person_id,omitempty"` // internal field, not exposed to API
	TenantID    string     `json:"-"`                    // RLS handles tenant isolation
	TeamID      *string    `json:"team_id,omitempty"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	Reason      *string    `json:"reason,omitempty"`
	IsRecurring bool       `json:"is_recurring"`
	DayOfWeek   *int       `json:"day_of_week,omitempty"` // 0=Sun, 6=Sat
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	TeamName string `json:"team_name,omitempty"` // joined from volunteer_teams
}

// BlockoutCreateRequest represents the API payload for creating a blockout
type BlockoutCreateRequest struct {
	TeamID      *string `json:"team_id,omitempty"`
	StartDate   string  `json:"start_date"` // YYYY-MM-DD format
	EndDate     string  `json:"end_date"`   // YYYY-MM-DD format (must be >= start_date)
	Reason      *string `json:"reason,omitempty"`
	IsRecurring bool    `json:"is_recurring"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"` // required if is_recurring=true
}

// BlockoutUpdateRequest represents the API payload for updating a blockout
type BlockoutUpdateRequest struct {
	TeamID      *string `json:"team_id,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	Reason      *string `json:"reason,omitempty"`
	IsRecurring *bool   `json:"is_recurring,omitempty"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"`
}

// ConflictInfo represents conflict detection result for scheduling
type ConflictInfo struct {
	IsBlocked     bool    `json:"is_blocked"`
	ConflictType  string  `json:"conflict_type,omitempty"` // "date_range" or "recurring"
	StartDate     *string `json:"start_date,omitempty"`
	EndDate       *string `json:"end_date,omitempty"`
	Reason        *string `json:"reason,omitempty"`
	DayOfWeek     *int    `json:"day_of_week,omitempty"`
	MatchingBlock string  `json:"matching_blockout_id,omitempty"` // ID of blocking record
}

// DayOfWeek returns the day of week (0=Sun, 6=Sat) for a date
func DayOfWeek(date time.Time) int {
	dow := int(date.Weekday())
	if dow == 0 {
		return 6 // Go's Sunday = 0, but we want 0
	}
	return dow - 1
}

// ParseDate parses a YYYY-MM-DD string to a time.Time (UTC midnight)
func ParseDate(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, time.UTC)
}
