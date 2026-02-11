package reports

import "time"

// AttendanceWeeklyData represents weekly attendance aggregations
type AttendanceWeeklyData struct {
	WeekStartDate string `json:"week_start_date"`
	AttendanceCount int    `json:"attendance_count"`
}

// AttendanceReport contains attendance trends and stats
type AttendanceReport struct {
	WeeklyData       []AttendanceWeeklyData `json:"weekly_data"`
	AverageAttendance float64                `json:"average_attendance"`
	GrowthPercentage  float64                `json:"growth_percentage"`
}

// GivingMonthlyData represents monthly giving aggregations
type GivingMonthlyData struct {
	Month       string  `json:"month"`
	TotalCents  int64   `json:"total_cents"`
	TotalAmount float64 `json:"total_amount"`
}

// GivingReport contains giving trends and stats
type GivingReport struct {
	MonthlyData      []GivingMonthlyData `json:"monthly_data"`
	TotalYTDCents    int64               `json:"total_ytd_cents"`
	TotalYTDAmount   float64             `json:"total_ytd_amount"`
	AverageGiftCents int64               `json:"average_gift_cents"`
	AverageGiftAmount float64            `json:"average_gift_amount"`
	DonorCount       int                 `json:"donor_count"`
}

// MembershipMonthlyData represents monthly membership growth
type MembershipMonthlyData struct {
	Month      string `json:"month"`
	TotalMembers int    `json:"total_members"`
}

// MembershipReport contains membership growth and stats
type MembershipReport struct {
	MonthlyData        []MembershipMonthlyData `json:"monthly_data"`
	NewMembersThisMonth int                    `json:"new_members_this_month"`
	NewMembersThisQuarter int                  `json:"new_members_this_quarter"`
	ActiveMembers      int                    `json:"active_members"`
	InactiveMembers    int                    `json:"inactive_members"`
}

// GroupParticipationReport contains group participation stats
type GroupParticipationReport struct {
	TotalMembers          int     `json:"total_members"`
	MembersInGroups       int     `json:"members_in_groups"`
	MembersNotInGroups    int     `json:"members_not_in_groups"`
	ParticipationRate     float64 `json:"participation_rate"`
	ActiveGroups          int     `json:"active_groups"`
	AverageMembersPerGroup float64 `json:"average_members_per_group"`
}
