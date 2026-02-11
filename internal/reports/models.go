package reports

// ---- Chart.js-ready response structures ----

// ChartDataset is a generic Chart.js dataset
type ChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor interface{} `json:"backgroundColor,omitempty"` // string or []string
	BorderColor     interface{} `json:"borderColor,omitempty"`     // string or []string
	BorderWidth     int       `json:"borderWidth,omitempty"`
	Tension         float64   `json:"tension,omitempty"`
	Fill            bool      `json:"fill,omitempty"`
	Type            string    `json:"type,omitempty"`
}

// ChartData is a generic Chart.js data object
type ChartData struct {
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

// KPI is a key performance indicator card
type KPI struct {
	Label  string  `json:"label"`
	Value  string  `json:"value"`
	Change float64 `json:"change,omitempty"` // percentage change
	Trend  string  `json:"trend,omitempty"`  // "up", "down", "flat"
}

// ---- Attendance ----

type AttendanceReport struct {
	WeeklyTrend      ChartData `json:"weekly_trend"`
	ByServiceType    ChartData `json:"by_service_type"`
	KPIs             []KPI     `json:"kpis"`
}

// ---- Giving ----

type GivingReport struct {
	MonthlyTotals  ChartData `json:"monthly_totals"`
	ByFund         ChartData `json:"by_fund"`
	DonorTrend     ChartData `json:"donor_trend"`
	KPIs           []KPI     `json:"kpis"`
}

// ---- Growth ----

type GrowthReport struct {
	MembershipGrowth ChartData `json:"membership_growth"`
	NewByMonth       ChartData `json:"new_by_month"`
	Funnel           []FunnelStep `json:"funnel"`
	KPIs             []KPI     `json:"kpis"`
}

type FunnelStep struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Pct   float64 `json:"pct"`
}

// ---- Songs ----

type SongsReport struct {
	TopSongs     ChartData    `json:"top_songs"`
	ByKey        ChartData    `json:"by_key"`
	UnusedSongs  []UnusedSong `json:"unused_songs"`
	KPIs         []KPI        `json:"kpis"`
}

type UnusedSong struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	LastUsed string `json:"last_used"`
}

// ---- Engagement ----

type EngagementReport struct {
	Distribution   ChartData `json:"distribution"`
	Trend          ChartData `json:"trend"`
	KPIs           []KPI     `json:"kpis"`
}
