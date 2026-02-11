package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// GetAttendanceReport generates weekly attendance trends for last 12 weeks
func (s *Service) GetAttendanceReport(ctx context.Context, tenantID uuid.UUID) (*AttendanceReport, error) {
	// Set tenant context for RLS
	if _, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get weekly attendance data for last 12 weeks
	query := `
		SELECT 
			DATE_TRUNC('week', checked_in_at)::DATE as week_start_date,
			COUNT(DISTINCT id) as attendance_count
		FROM checkins
		WHERE tenant_id = $1
			AND checked_in_at >= NOW() - INTERVAL '12 weeks'
		GROUP BY week_start_date
		ORDER BY week_start_date ASC
	`

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query weekly attendance: %w", err)
	}
	defer rows.Close()

	var weeklyData []AttendanceWeeklyData
	var totalAttendance int
	for rows.Next() {
		var data AttendanceWeeklyData
		if err := rows.Scan(&data.WeekStartDate, &data.AttendanceCount); err != nil {
			return nil, fmt.Errorf("failed to scan weekly data: %w", err)
		}
		weeklyData = append(weeklyData, data)
		totalAttendance += data.AttendanceCount
	}

	// Calculate average and growth percentage
	average := 0.0
	growth := 0.0
	if len(weeklyData) > 0 {
		average = float64(totalAttendance) / float64(len(weeklyData))
		
		// Growth = (last week - first week) / first week * 100
		if len(weeklyData) >= 2 && weeklyData[0].AttendanceCount > 0 {
			firstWeek := float64(weeklyData[0].AttendanceCount)
			lastWeek := float64(weeklyData[len(weeklyData)-1].AttendanceCount)
			growth = ((lastWeek - firstWeek) / firstWeek) * 100
		}
	}

	return &AttendanceReport{
		WeeklyData:       weeklyData,
		AverageAttendance: average,
		GrowthPercentage:  growth,
	}, nil
}

// GetGivingReport generates monthly giving trends for last 12 months
func (s *Service) GetGivingReport(ctx context.Context, tenantID uuid.UUID) (*GivingReport, error) {
	// Set tenant context for RLS
	if _, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get monthly giving data for last 12 months
	query := `
		SELECT 
			TO_CHAR(donated_at, 'YYYY-MM') as month,
			SUM(amount_cents) as total_cents
		FROM donations
		WHERE tenant_id = $1
			AND donated_at >= NOW() - INTERVAL '12 months'
			AND status = 'completed'
		GROUP BY month
		ORDER BY month ASC
	`

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly giving: %w", err)
	}
	defer rows.Close()

	var monthlyData []GivingMonthlyData
	var totalCents int64
	for rows.Next() {
		var data GivingMonthlyData
		if err := rows.Scan(&data.Month, &data.TotalCents); err != nil {
			return nil, fmt.Errorf("failed to scan monthly data: %w", err)
		}
		data.TotalAmount = float64(data.TotalCents) / 100.0
		monthlyData = append(monthlyData, data)
		totalCents += data.TotalCents
	}

	// Get YTD total and donor count
	currentYear := time.Now().Year()
	statsQuery := `
		SELECT 
			COALESCE(SUM(amount_cents), 0) as ytd_total,
			COUNT(DISTINCT person_id) as donor_count,
			COALESCE(AVG(amount_cents), 0)::BIGINT as avg_gift
		FROM donations
		WHERE tenant_id = $1
			AND EXTRACT(YEAR FROM donated_at) = $2
			AND status = 'completed'
	`

	var ytdTotal, avgGift int64
	var donorCount int
	err = s.db.QueryRow(ctx, statsQuery, tenantID, currentYear).Scan(&ytdTotal, &donorCount, &avgGift)
	if err != nil {
		return nil, fmt.Errorf("failed to query giving stats: %w", err)
	}

	return &GivingReport{
		MonthlyData:       monthlyData,
		TotalYTDCents:     ytdTotal,
		TotalYTDAmount:    float64(ytdTotal) / 100.0,
		AverageGiftCents:  avgGift,
		AverageGiftAmount: float64(avgGift) / 100.0,
		DonorCount:        donorCount,
	}, nil
}

// GetMembershipReport generates membership growth trends
func (s *Service) GetMembershipReport(ctx context.Context, tenantID uuid.UUID) (*MembershipReport, error) {
	// Set tenant context for RLS
	if _, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get monthly membership growth (cumulative count)
	// We'll approximate by counting people created up to each month
	query := `
		WITH months AS (
			SELECT DATE_TRUNC('month', d)::DATE as month
			FROM generate_series(
				NOW() - INTERVAL '12 months',
				NOW(),
				INTERVAL '1 month'
			) d
		)
		SELECT 
			TO_CHAR(m.month, 'YYYY-MM') as month,
			COUNT(p.id) as total_members
		FROM months m
		LEFT JOIN people p ON p.tenant_id = $1 AND p.created_at <= m.month + INTERVAL '1 month'
		GROUP BY m.month
		ORDER BY m.month ASC
	`

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query membership growth: %w", err)
	}
	defer rows.Close()

	var monthlyData []MembershipMonthlyData
	for rows.Next() {
		var data MembershipMonthlyData
		if err := rows.Scan(&data.Month, &data.TotalMembers); err != nil {
			return nil, fmt.Errorf("failed to scan membership data: %w", err)
		}
		monthlyData = append(monthlyData, data)
	}

	// Get new members this month and quarter
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	quarterStart := time.Date(now.Year(), ((now.Month()-1)/3)*3+1, 1, 0, 0, 0, 0, now.Location())

	var newThisMonth, newThisQuarter int
	err = s.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE created_at >= $2) as new_this_month,
			COUNT(*) FILTER (WHERE created_at >= $3) as new_this_quarter
		FROM people
		WHERE tenant_id = $1
	`, tenantID, monthStart, quarterStart).Scan(&newThisMonth, &newThisQuarter)
	if err != nil {
		return nil, fmt.Errorf("failed to query new members: %w", err)
	}

	// Get active vs inactive (simplified: anyone with recent checkin is active)
	var activeCount, inactiveCount int
	err = s.db.QueryRow(ctx, `
		SELECT 
			COUNT(DISTINCT p.id) FILTER (WHERE c.id IS NOT NULL) as active_members,
			COUNT(DISTINCT p.id) FILTER (WHERE c.id IS NULL) as inactive_members
		FROM people p
		LEFT JOIN checkins c ON c.person_id = p.id 
			AND c.checked_in_at >= NOW() - INTERVAL '90 days'
		WHERE p.tenant_id = $1
	`, tenantID).Scan(&activeCount, &inactiveCount)
	if err != nil {
		return nil, fmt.Errorf("failed to query active members: %w", err)
	}

	return &MembershipReport{
		MonthlyData:           monthlyData,
		NewMembersThisMonth:   newThisMonth,
		NewMembersThisQuarter: newThisQuarter,
		ActiveMembers:         activeCount,
		InactiveMembers:       inactiveCount,
	}, nil
}

// GetGroupParticipationReport generates group participation stats
func (s *Service) GetGroupParticipationReport(ctx context.Context, tenantID uuid.UUID) (*GroupParticipationReport, error) {
	// Set tenant context for RLS
	if _, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var report GroupParticipationReport

	// Get total members, members in groups, and members not in groups
	err := s.db.QueryRow(ctx, `
		SELECT 
			COUNT(DISTINCT p.id) as total_members,
			COUNT(DISTINCT gm.person_id) as members_in_groups,
			COUNT(DISTINCT p.id) - COUNT(DISTINCT gm.person_id) as members_not_in_groups
		FROM people p
		LEFT JOIN group_members gm ON gm.person_id = p.id
		WHERE p.tenant_id = $1
	`, tenantID).Scan(&report.TotalMembers, &report.MembersInGroups, &report.MembersNotInGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to query group participation: %w", err)
	}

	// Calculate participation rate
	if report.TotalMembers > 0 {
		report.ParticipationRate = (float64(report.MembersInGroups) / float64(report.TotalMembers)) * 100
	}

	// Get active groups and average members per group
	err = s.db.QueryRow(ctx, `
		SELECT 
			COUNT(DISTINCT g.id) as active_groups,
			COALESCE(AVG(member_count), 0) as avg_members_per_group
		FROM groups g
		LEFT JOIN (
			SELECT group_id, COUNT(*) as member_count
			FROM group_members
			GROUP BY group_id
		) gm ON gm.group_id = g.id
		WHERE g.tenant_id = $1 AND g.is_active = true
	`, tenantID).Scan(&report.ActiveGroups, &report.AverageMembersPerGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to query active groups: %w", err)
	}

	return &report, nil
}
