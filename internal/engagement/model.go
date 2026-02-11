package engagement

import "time"

type EngagementScore struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	PersonID         string    `json:"person_id"`
	Score            int       `json:"score"`
	AttendanceScore  int       `json:"attendance_score"`
	GivingScore      int       `json:"giving_score"`
	GroupScore       int       `json:"group_score"`
	VolunteerScore   int       `json:"volunteer_score"`
	ConnectionScore  int       `json:"connection_score"`
	CalculatedAt     time.Time `json:"calculated_at"`
	PersonName       string    `json:"person_name,omitempty"`
	PersonEmail      string    `json:"person_email,omitempty"`
}

type DashboardActivity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Icon        string    `json:"icon"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

type EngagementScoreHistory struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	PersonID        string    `json:"person_id"`
	Score           int       `json:"score"`
	AttendanceScore int       `json:"attendance_score"`
	GivingScore     int       `json:"giving_score"`
	GroupScore      int       `json:"group_score"`
	VolunteerScore  int       `json:"volunteer_score"`
	ConnectionScore int       `json:"connection_score"`
	RecordedAt      time.Time `json:"recorded_at"`
}

type AtRiskPerson struct {
	PersonID        string  `json:"person_id"`
	PersonName      string  `json:"person_name"`
	PersonEmail     string  `json:"person_email,omitempty"`
	CurrentScore    int     `json:"current_score"`
	PreviousScore   int     `json:"previous_score"`
	ScoreChange     int     `json:"score_change"`
	PercentChange   float64 `json:"percent_change"`
	LastCalculated  time.Time `json:"last_calculated"`
}

type EngagementDistribution struct {
	High     int `json:"high"`      // 75-100
	Medium   int `json:"medium"`    // 50-74
	Low      int `json:"low"`       // 25-49
	Inactive int `json:"inactive"`  // 0-24
}

type DashboardKPIs struct {
	TotalActiveMembers     int                    `json:"total_active_members"`
	AverageAttendance      float64                `json:"average_attendance"`
	GivingThisMonth        int                    `json:"giving_this_month_cents"`
	GivingLastMonth        int                    `json:"giving_last_month_cents"`
	GivingPercentChange    float64                `json:"giving_percent_change"`
	NewVisitorsThisMonth   int                    `json:"new_visitors_this_month"`
	EngagementDistribution EngagementDistribution `json:"engagement_distribution"`
	AttendanceTrend        []int                  `json:"attendance_trend"`
}

type ActionItems struct {
	OverdueFollowUps       int `json:"overdue_followups"`
	UnreadPrayerRequests   int `json:"unread_prayer_requests"`
	PendingVolunteers      int `json:"pending_volunteers"`
	AtRiskCount            int `json:"at_risk_count"`
}
