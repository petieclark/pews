package digest

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed templates/*.html
var templatesFS embed.FS

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// GetSettings retrieves digest settings for a tenant
func (s *Service) GetSettings(ctx context.Context, tenantID string) (*DigestSettings, error) {
	var settings DigestSettings
	recipients := []string{}

	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, enabled, send_day, recipients, created_at, updated_at
		FROM digest_settings
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&settings.ID,
		&settings.TenantID,
		&settings.Enabled,
		&settings.SendDay,
		&recipients,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		// Create default settings
		return s.CreateSettings(ctx, tenantID)
	}
	if err != nil {
		return nil, err
	}

	settings.Recipients = recipients
	return &settings, nil
}

// CreateSettings creates default digest settings for a tenant
func (s *Service) CreateSettings(ctx context.Context, tenantID string) (*DigestSettings, error) {
	var settings DigestSettings
	recipients := []string{}

	err := s.db.QueryRow(ctx, `
		INSERT INTO digest_settings (tenant_id, enabled, send_day, recipients)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, enabled, send_day, recipients, created_at, updated_at
	`, tenantID, true, "monday", []string{}).Scan(
		&settings.ID,
		&settings.TenantID,
		&settings.Enabled,
		&settings.SendDay,
		&recipients,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	settings.Recipients = recipients
	return &settings, nil
}

// UpdateSettings updates digest settings
func (s *Service) UpdateSettings(ctx context.Context, tenantID string, enabled bool, sendDay string, recipients []string) (*DigestSettings, error) {
	var settings DigestSettings
	recipientsArray := []string{}

	err := s.db.QueryRow(ctx, `
		UPDATE digest_settings
		SET enabled = $2, send_day = $3, recipients = $4, updated_at = NOW()
		WHERE tenant_id = $1
		RETURNING id, tenant_id, enabled, send_day, recipients, created_at, updated_at
	`, tenantID, enabled, sendDay, recipients).Scan(
		&settings.ID,
		&settings.TenantID,
		&settings.Enabled,
		&settings.SendDay,
		&recipientsArray,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	settings.Recipients = recipientsArray
	return &settings, nil
}

// GenerateWeeklyDigest compiles all digest data for the past week
func (s *Service) GenerateWeeklyDigest(ctx context.Context, tenantID string) (*WeeklyDigest, error) {
	// Calculate week boundaries (last 7 days)
	weekEnd := time.Now()
	weekStart := weekEnd.AddDate(0, 0, -7)

	digest := &WeeklyDigest{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
	}

	// Get tenant name
	err := s.db.QueryRow(ctx, "SELECT name FROM tenants WHERE id = $1", tenantID).Scan(&digest.TenantName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant name: %w", err)
	}

	// Get attendance stats
	attendance, err := s.getAttendanceStats(ctx, tenantID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance: %w", err)
	}
	digest.Attendance = attendance

	// Get member stats
	members, err := s.getMemberStats(ctx, tenantID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get member stats: %w", err)
	}
	digest.Members = members

	// Get giving stats
	giving, err := s.getGivingStats(ctx, tenantID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get giving stats: %w", err)
	}
	digest.Giving = giving

	// Get upcoming services
	upcoming, err := s.getUpcomingServices(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming services: %w", err)
	}
	digest.UpcomingServices = upcoming

	// Get new prayer requests (from communication module if available)
	prayers, err := s.getNewPrayerRequests(ctx, tenantID, weekStart)
	if err != nil {
		// Non-critical, just log and continue
		digest.PrayerRequests = []PrayerRequest{}
	} else {
		digest.PrayerRequests = prayers
	}

	// Get volunteer schedule for next week
	volunteers, err := s.getVolunteerSchedule(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volunteer schedule: %w", err)
	}
	digest.Volunteers = volunteers

	return digest, nil
}

func (s *Service) getAttendanceStats(ctx context.Context, tenantID string, weekStart, weekEnd time.Time) (AttendanceStats, error) {
	var stats AttendanceStats

	// This week's attendance
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(COUNT(DISTINCT person_id), 0)
		FROM checkin_records
		WHERE tenant_id = $1
		AND checked_in_at >= $2
		AND checked_in_at < $3
	`, tenantID, weekStart, weekEnd).Scan(&stats.ThisWeek)

	if err != nil {
		return stats, err
	}

	// Last week's attendance
	lastWeekStart := weekStart.AddDate(0, 0, -7)
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(COUNT(DISTINCT person_id), 0)
		FROM checkin_records
		WHERE tenant_id = $1
		AND checked_in_at >= $2
		AND checked_in_at < $3
	`, tenantID, lastWeekStart, weekStart).Scan(&stats.LastWeek)

	if err != nil {
		return stats, err
	}

	stats.Change = stats.ThisWeek - stats.LastWeek
	if stats.LastWeek > 0 {
		stats.ChangePercent = (float64(stats.Change) / float64(stats.LastWeek)) * 100
	}

	return stats, nil
}

func (s *Service) getMemberStats(ctx context.Context, tenantID string, weekStart time.Time) (MemberStats, error) {
	var stats MemberStats

	// New members this week
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*), 0)
		FROM people
		WHERE tenant_id = $1
		AND created_at >= $2
	`, tenantID, weekStart).Scan(&stats.NewThisWeek)

	if err != nil {
		return stats, err
	}

	// Total active members
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*), 0)
		FROM people
		WHERE tenant_id = $1
		AND is_active = TRUE
	`, tenantID).Scan(&stats.TotalActive)

	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (s *Service) getGivingStats(ctx context.Context, tenantID string, weekStart, weekEnd time.Time) (GivingStats, error) {
	var stats GivingStats

	// This week's giving
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents), 0)
		FROM donations
		WHERE tenant_id = $1
		AND donated_at >= $2
		AND donated_at < $3
		AND status = 'completed'
	`, tenantID, weekStart, weekEnd).Scan(&stats.ThisWeekCents)

	if err != nil {
		return stats, err
	}

	// Year to date
	yearStart := time.Date(weekEnd.Year(), 1, 1, 0, 0, 0, 0, weekEnd.Location())
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents), 0)
		FROM donations
		WHERE tenant_id = $1
		AND donated_at >= $2
		AND donated_at < $3
		AND status = 'completed'
	`, tenantID, yearStart, weekEnd).Scan(&stats.YearToDateCents)

	if err != nil {
		return stats, err
	}

	// Format display strings
	stats.ThisWeekDisplay = formatCents(stats.ThisWeekCents)
	stats.YearToDateDisplay = formatCents(stats.YearToDateCents)

	return stats, nil
}

func (s *Service) getUpcomingServices(ctx context.Context, tenantID string) ([]UpcomingService, error) {
	rows, err := s.db.Query(ctx, `
		SELECT s.name, s.service_date, s.service_time
		FROM church_services s
		WHERE s.tenant_id = $1
		AND s.service_date >= CURRENT_DATE
		AND s.service_date <= CURRENT_DATE + INTERVAL '7 days'
		ORDER BY s.service_date, s.service_time
		LIMIT 5
	`, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []UpcomingService{}
	for rows.Next() {
		var svc UpcomingService
		var serviceTime *string
		err := rows.Scan(&svc.Name, &svc.ServiceDate, &serviceTime)
		if err != nil {
			return nil, err
		}
		if serviceTime != nil {
			svc.ServiceTime = *serviceTime
		}
		services = append(services, svc)
	}

	return services, nil
}

func (s *Service) getNewPrayerRequests(ctx context.Context, tenantID string, weekStart time.Time) ([]PrayerRequest, error) {
	// Prayer requests might be in connection_cards or a dedicated table
	// For now, check connection cards
	rows, err := s.db.Query(ctx, `
		SELECT 
			COALESCE(c.first_name || ' ' || c.last_name, 'Anonymous'),
			c.prayer_requests,
			c.submitted_at
		FROM connection_cards c
		WHERE c.tenant_id = $1
		AND c.prayer_requests IS NOT NULL
		AND c.prayer_requests != ''
		AND c.submitted_at >= $2
		ORDER BY c.submitted_at DESC
		LIMIT 10
	`, tenantID, weekStart)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []PrayerRequest{}
	for rows.Next() {
		var req PrayerRequest
		err := rows.Scan(&req.PersonName, &req.Request, &req.CreatedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *Service) getVolunteerSchedule(ctx context.Context, tenantID string) ([]VolunteerSchedule, error) {
	// Get service team assignments for next week
	rows, err := s.db.Query(ctx, `
		SELECT 
			s.name,
			s.service_date,
			COALESCE(p.first_name || ' ' || p.last_name, 'Unassigned'),
			st.role
		FROM service_team st
		JOIN church_services s ON s.id = st.service_id
		LEFT JOIN people p ON p.id = st.person_id
		WHERE s.tenant_id = $1
		AND s.service_date >= CURRENT_DATE
		AND s.service_date <= CURRENT_DATE + INTERVAL '7 days'
		ORDER BY s.service_date, st.role
		LIMIT 20
	`, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	volunteers := []VolunteerSchedule{}
	for rows.Next() {
		var vol VolunteerSchedule
		err := rows.Scan(&vol.ServiceName, &vol.ServiceDate, &vol.PersonName, &vol.Role)
		if err != nil {
			return nil, err
		}
		volunteers = append(volunteers, vol)
	}

	return volunteers, nil
}

// RenderDigestHTML generates the HTML email from template
func (s *Service) RenderDigestHTML(digest *WeeklyDigest) (string, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/weekly.html")
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, digest)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Helper to format cents as dollars
func formatCents(cents int) string {
	dollars := float64(cents) / 100.0
	return fmt.Sprintf("$%.2f", dollars)
}
