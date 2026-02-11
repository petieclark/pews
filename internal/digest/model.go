package digest

import (
	"time"
)

type DigestSettings struct {
	ID         string   `json:"id"`
	TenantID   string   `json:"tenant_id"`
	Enabled    bool     `json:"enabled"`
	SendDay    string   `json:"send_day"`
	Recipients []string `json:"recipients"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DigestHistory struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	WeekStart  time.Time `json:"week_start"`
	WeekEnd    time.Time `json:"week_end"`
	SentAt     time.Time `json:"sent_at"`
	Recipients []string  `json:"recipients"`
	CreatedAt  time.Time `json:"created_at"`
}

type WeeklyDigest struct {
	TenantName       string           `json:"tenant_name"`
	WeekStart        time.Time        `json:"week_start"`
	WeekEnd          time.Time        `json:"week_end"`
	Attendance       AttendanceStats  `json:"attendance"`
	Members          MemberStats      `json:"members"`
	Giving           GivingStats      `json:"giving"`
	UpcomingServices []UpcomingService `json:"upcoming_services"`
	PrayerRequests   []PrayerRequest  `json:"prayer_requests"`
	Volunteers       []VolunteerSchedule `json:"volunteers"`
}

type AttendanceStats struct {
	ThisWeek  int     `json:"this_week"`
	LastWeek  int     `json:"last_week"`
	Change    int     `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

type MemberStats struct {
	NewThisWeek int `json:"new_this_week"`
	TotalActive int `json:"total_active"`
}

type GivingStats struct {
	ThisWeekCents int     `json:"this_week_cents"`
	YearToDateCents int   `json:"year_to_date_cents"`
	ThisWeekDisplay string `json:"this_week_display"`
	YearToDateDisplay string `json:"year_to_date_display"`
}

type UpcomingService struct {
	Name        string    `json:"name"`
	ServiceDate time.Time `json:"service_date"`
	ServiceTime string    `json:"service_time"`
}

type PrayerRequest struct {
	PersonName string    `json:"person_name"`
	Request    string    `json:"request"`
	CreatedAt  time.Time `json:"created_at"`
}

type VolunteerSchedule struct {
	ServiceName string    `json:"service_name"`
	ServiceDate time.Time `json:"service_date"`
	PersonName  string    `json:"person_name"`
	Role        string    `json:"role"`
}
