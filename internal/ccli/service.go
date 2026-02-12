package ccli

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

// GenerateCCLIReport queries service_items joined with songs to count how many times
// each CCLI-licensed song was used in services during the given period.
func (s *Service) GenerateCCLIReport(ctx context.Context, tenantID string, startDate, endDate time.Time) (*CCLIReport, error) {
	// Get the tenant's CCLI license number
	settings, err := s.GetSettings(ctx, tenantID)
	if err != nil {
		settings = &CCLISettings{} // proceed without license number
	}

	rows, err := s.db.Query(ctx, `
		SELECT 
			songs.title,
			songs.ccli_number,
			COALESCE(songs.artist, '') as artist,
			COUNT(*) as times_used,
			MAX(sv.service_date) as last_used
		FROM service_items si
		JOIN songs ON songs.id = si.song_id
		JOIN services sv ON sv.id = si.service_id
		WHERE songs.ccli_number IS NOT NULL 
			AND songs.ccli_number != ''
			AND sv.service_date >= $1
			AND sv.service_date <= $2
		GROUP BY songs.id, songs.title, songs.ccli_number, songs.artist
		ORDER BY times_used DESC, songs.title ASC`,
		startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query CCLI song usage: %w", err)
	}
	defer rows.Close()

	var songs []CCLISongUsage
	totalUses := 0

	for rows.Next() {
		var usage CCLISongUsage
		var lastUsed time.Time
		if err := rows.Scan(&usage.Title, &usage.CCLINumber, &usage.Artist, &usage.TimesUsed, &lastUsed); err != nil {
			return nil, fmt.Errorf("failed to scan song usage: %w", err)
		}
		usage.LastUsed = lastUsed.Format("2006-01-02")
		totalUses += usage.TimesUsed
		songs = append(songs, usage)
	}

	if songs == nil {
		songs = []CCLISongUsage{}
	}

	return &CCLIReport{
		LicenseNumber: settings.LicenseNumber,
		Period:        fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:     startDate,
		EndDate:       endDate,
		Songs:         songs,
		TotalSongs:    len(songs),
		TotalUses:     totalUses,
		GeneratedAt:   time.Now(),
	}, nil
}

// GetStats returns quick CCLI stats for the current quarter.
func (s *Service) GetStats(ctx context.Context, tenantID string) (*CCLIStats, error) {
	// Current quarter boundaries
	now := time.Now()
	quarterStart := time.Date(now.Year(), ((now.Month()-1)/3)*3+1, 1, 0, 0, 0, 0, time.UTC)
	quarterEnd := quarterStart.AddDate(0, 3, 0).Add(-time.Second)

	// Total songs with CCLI numbers
	var totalLicensed int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM songs 
		WHERE ccli_number IS NOT NULL AND ccli_number != ''`).Scan(&totalLicensed)
	if err != nil {
		return nil, fmt.Errorf("failed to count licensed songs: %w", err)
	}

	// Songs used this quarter
	rows, err := s.db.Query(ctx, `
		SELECT 
			songs.title,
			songs.ccli_number,
			COALESCE(songs.artist, '') as artist,
			COUNT(*) as times_used,
			MAX(sv.service_date) as last_used
		FROM service_items si
		JOIN songs ON songs.id = si.song_id
		JOIN services sv ON sv.id = si.service_id
		WHERE songs.ccli_number IS NOT NULL 
			AND songs.ccli_number != ''
			AND sv.service_date >= $1
			AND sv.service_date <= $2
		GROUP BY songs.id, songs.title, songs.ccli_number, songs.artist
		ORDER BY times_used DESC
		LIMIT 10`,
		quarterStart, quarterEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to query top songs: %w", err)
	}
	defer rows.Close()

	var topSongs []CCLISongUsage
	songsUsed := 0
	totalUses := 0

	for rows.Next() {
		var usage CCLISongUsage
		var lastUsed time.Time
		if err := rows.Scan(&usage.Title, &usage.CCLINumber, &usage.Artist, &usage.TimesUsed, &lastUsed); err != nil {
			return nil, fmt.Errorf("failed to scan top song: %w", err)
		}
		usage.LastUsed = lastUsed.Format("2006-01-02")
		totalUses += usage.TimesUsed
		songsUsed++
		topSongs = append(topSongs, usage)
	}

	if topSongs == nil {
		topSongs = []CCLISongUsage{}
	}

	// Get total unique songs used (not just top 10)
	var totalSongsUsed int
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT songs.id)
		FROM service_items si
		JOIN songs ON songs.id = si.song_id
		JOIN services sv ON sv.id = si.service_id
		WHERE songs.ccli_number IS NOT NULL 
			AND songs.ccli_number != ''
			AND sv.service_date >= $1
			AND sv.service_date <= $2`,
		quarterStart, quarterEnd).Scan(&totalSongsUsed)
	if err != nil {
		totalSongsUsed = songsUsed
	}

	// Determine reporting status
	settings, _ := s.GetSettings(ctx, tenantID)
	status := "not_configured"
	if settings != nil && settings.LicenseNumber != "" {
		status = "current"
		if settings.LastReportedAt != nil && settings.LastReportedAt.Before(quarterStart) {
			status = "overdue"
		}
	}

	licenseNum := ""
	if settings != nil {
		licenseNum = settings.LicenseNumber
	}

	return &CCLIStats{
		TotalLicensedSongs:  totalLicensed,
		SongsUsedThisPeriod: totalSongsUsed,
		TotalUsesThisPeriod: totalUses,
		TopSongs:            topSongs,
		ReportingStatus:     status,
		LicenseNumber:       licenseNum,
	}, nil
}

// GetSettings retrieves CCLI settings for a tenant.
func (s *Service) GetSettings(ctx context.Context, tenantID string) (*CCLISettings, error) {
	var settings CCLISettings
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, COALESCE(license_number, ''), auto_report, 
		       COALESCE(report_frequency, 'quarterly'), last_reported_at, created_at, updated_at
		FROM ccli_settings
		WHERE tenant_id = $1`, tenantID).Scan(
		&settings.ID, &settings.TenantID, &settings.LicenseNumber,
		&settings.AutoReport, &settings.ReportFrequency,
		&settings.LastReportedAt, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("CCLI settings not found: %w", err)
	}
	return &settings, nil
}

// SaveSettings creates or updates CCLI settings for a tenant.
func (s *Service) SaveSettings(ctx context.Context, tenantID string, licenseNumber string, autoReport bool, reportFrequency string) (*CCLISettings, error) {
	if reportFrequency == "" {
		reportFrequency = "quarterly"
	}

	var settings CCLISettings
	err := s.db.QueryRow(ctx, `
		INSERT INTO ccli_settings (id, tenant_id, license_number, auto_report, report_frequency)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id) DO UPDATE SET
			license_number = EXCLUDED.license_number,
			auto_report = EXCLUDED.auto_report,
			report_frequency = EXCLUDED.report_frequency,
			updated_at = NOW()
		RETURNING id, tenant_id, license_number, auto_report, report_frequency, last_reported_at, created_at, updated_at`,
		uuid.New().String(), tenantID, licenseNumber, autoReport, reportFrequency).Scan(
		&settings.ID, &settings.TenantID, &settings.LicenseNumber,
		&settings.AutoReport, &settings.ReportFrequency,
		&settings.LastReportedAt, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save CCLI settings: %w", err)
	}
	return &settings, nil
}
