package ccli

import "time"

// CCLIReport represents a CCLI license usage report for a given period.
// CCLI requires: Song title, CCLI song number, author/composer, copyright holder,
// number of times performed/projected/printed.
type CCLIReport struct {
	LicenseNumber string          `json:"license_number"`
	Period        string          `json:"period"`
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	Songs         []CCLISongUsage `json:"songs"`
	TotalSongs    int             `json:"total_songs"`
	TotalUses     int             `json:"total_uses"`
	GeneratedAt   time.Time       `json:"generated_at"`
}

// CCLISongUsage represents a single song's usage data for CCLI reporting.
type CCLISongUsage struct {
	Title      string `json:"title"`
	CCLINumber string `json:"ccli_number"`
	Artist     string `json:"artist"`
	TimesUsed  int    `json:"times_used"`
	LastUsed   string `json:"last_used"`
}

// CCLISettings stores a tenant's CCLI license configuration.
type CCLISettings struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	LicenseNumber   string     `json:"license_number"`
	AutoReport      bool       `json:"auto_report"`
	ReportFrequency string     `json:"report_frequency"` // quarterly, semi-annual, annual
	LastReportedAt  *time.Time `json:"last_reported_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CCLIStats provides a quick overview of CCLI-related metrics.
type CCLIStats struct {
	TotalLicensedSongs int             `json:"total_licensed_songs"`
	SongsUsedThisPeriod int            `json:"songs_used_this_period"`
	TotalUsesThisPeriod int            `json:"total_uses_this_period"`
	TopSongs           []CCLISongUsage `json:"top_songs"`
	ReportingStatus    string          `json:"reporting_status"` // current, overdue, not_configured
	LicenseNumber      string          `json:"license_number,omitempty"`
}
