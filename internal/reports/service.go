package reports

import (
	"context"
	"fmt"
	"math"
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

var brandColors = []string{
	"#4A8B8C", "#1B3A4B", "#8FBCB0", "#6BA3A4", "#2D5F6E",
	"#A7D0C4", "#3C7A7B", "#1D4D5E", "#B5D9CE", "#5A9B9C",
}

func (s *Service) setTenant(ctx context.Context, tenantID uuid.UUID) error {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	return err
}

func parseRange(rangeStr string) (time.Time, string) {
	now := time.Now()
	switch rangeStr {
	case "1m":
		return now.AddDate(0, -1, 0), "1 month"
	case "3m":
		return now.AddDate(0, -3, 0), "3 months"
	case "6m":
		return now.AddDate(0, -6, 0), "6 months"
	case "1y", "12m":
		return now.AddDate(-1, 0, 0), "12 months"
	case "12w":
		return now.AddDate(0, 0, -84), "12 weeks"
	default:
		return now.AddDate(-1, 0, 0), "12 months"
	}
}

// ---- Attendance ----

func (s *Service) GetAttendanceReport(ctx context.Context, tenantID uuid.UUID, rangeStr string) (*AttendanceReport, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	since, _ := parseRange(rangeStr)
	if rangeStr == "" {
		since, _ = parseRange("12w")
	}

	// Weekly attendance trend
	rows, err := s.db.Query(ctx, `
		SELECT DATE_TRUNC('week', checked_in_at)::DATE as week, COUNT(DISTINCT id) as cnt
		FROM checkins WHERE tenant_id = $1 AND checked_in_at >= $2
		GROUP BY week ORDER BY week`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("attendance weekly: %w", err)
	}
	defer rows.Close()

	labels := []string{}
	data := []float64{}
	var total float64
	var peak float64
	for rows.Next() {
		var week time.Time
		var cnt int
		if err := rows.Scan(&week, &cnt); err != nil {
			return nil, err
		}
		labels = append(labels, week.Format("Jan 2"))
		v := float64(cnt)
		data = append(data, v)
		total += v
		if v > peak {
			peak = v
		}
	}

	weeklyTrend := ChartData{
		Labels: labels,
		Datasets: []ChartDataset{{
			Label:           "Weekly Attendance",
			Data:            data,
			BorderColor:     "#4A8B8C",
			BackgroundColor: "rgba(74,139,140,0.15)",
			Tension:         0.3,
			Fill:            true,
		}},
	}

	// By service type
	rows2, err := s.db.Query(ctx, `
		SELECT COALESCE(st.name, 'Unknown') as type_name, COUNT(c.id)
		FROM checkins c
		LEFT JOIN checkin_events ce ON ce.id = c.event_id AND ce.tenant_id = c.tenant_id
		LEFT JOIN services sv ON sv.id = ce.service_id AND sv.tenant_id = c.tenant_id
		LEFT JOIN service_types st ON st.id = sv.service_type_id AND st.tenant_id = c.tenant_id
		WHERE c.tenant_id = $1 AND c.checked_in_at >= $2
		GROUP BY type_name ORDER BY COUNT(c.id) DESC`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("attendance by type: %w", err)
	}
	defer rows2.Close()

	stLabels := []string{}
	stData := []float64{}
	stColors := []string{}
	i := 0
	for rows2.Next() {
		var name string
		var cnt int
		if err := rows2.Scan(&name, &cnt); err != nil {
			return nil, err
		}
		stLabels = append(stLabels, name)
		stData = append(stData, float64(cnt))
		stColors = append(stColors, brandColors[i%len(brandColors)])
		i++
	}

	byServiceType := ChartData{
		Labels: stLabels,
		Datasets: []ChartDataset{{
			Label:           "Attendance",
			Data:            stData,
			BackgroundColor: stColors,
			BorderWidth:     1,
		}},
	}

	avg := 0.0
	growth := 0.0
	if len(data) > 0 {
		avg = total / float64(len(data))
		if len(data) >= 2 && data[0] > 0 {
			growth = ((data[len(data)-1] - data[0]) / data[0]) * 100
		}
	}

	trend := "flat"
	if growth > 1 {
		trend = "up"
	} else if growth < -1 {
		trend = "down"
	}

	return &AttendanceReport{
		WeeklyTrend:   weeklyTrend,
		ByServiceType: byServiceType,
		KPIs: []KPI{
			{Label: "Average Attendance", Value: fmt.Sprintf("%.0f", avg)},
			{Label: "Peak Attendance", Value: fmt.Sprintf("%.0f", peak)},
			{Label: "Growth Rate", Value: fmt.Sprintf("%.1f%%", growth), Change: growth, Trend: trend},
		},
	}, nil
}

// ---- Giving ----

func (s *Service) GetGivingReport(ctx context.Context, tenantID uuid.UUID, rangeStr string) (*GivingReport, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	since, _ := parseRange(rangeStr)
	if rangeStr == "" {
		since, _ = parseRange("12m")
	}

	// Monthly totals
	rows, err := s.db.Query(ctx, `
		SELECT TO_CHAR(donated_at, 'YYYY-MM') as month, SUM(amount_cents)
		FROM donations WHERE tenant_id = $1 AND donated_at >= $2 AND status = 'completed'
		GROUP BY month ORDER BY month`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("giving monthly: %w", err)
	}
	defer rows.Close()

	mLabels := []string{}
	mData := []float64{}
	for rows.Next() {
		var month string
		var cents int64
		if err := rows.Scan(&month, &cents); err != nil {
			return nil, err
		}
		mLabels = append(mLabels, month)
		mData = append(mData, float64(cents)/100.0)
	}

	monthlyTotals := ChartData{
		Labels: mLabels,
		Datasets: []ChartDataset{{
			Label:           "Monthly Giving",
			Data:            mData,
			BackgroundColor: "#4A8B8C",
			BorderColor:     "#1B3A4B",
			BorderWidth:     1,
		}},
	}

	// By fund
	rows2, err := s.db.Query(ctx, `
		SELECT COALESCE(f.name, 'General') as fund_name, SUM(d.amount_cents)
		FROM donations d LEFT JOIN funds f ON f.id = d.fund_id AND f.tenant_id = d.tenant_id
		WHERE d.tenant_id = $1 AND d.donated_at >= $2 AND d.status = 'completed'
		GROUP BY fund_name ORDER BY SUM(d.amount_cents) DESC`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("giving by fund: %w", err)
	}
	defer rows2.Close()

	fLabels := []string{}
	fData := []float64{}
	fColors := []string{}
	i := 0
	for rows2.Next() {
		var name string
		var cents int64
		if err := rows2.Scan(&name, &cents); err != nil {
			return nil, err
		}
		fLabels = append(fLabels, name)
		fData = append(fData, float64(cents)/100.0)
		fColors = append(fColors, brandColors[i%len(brandColors)])
		i++
	}

	byFund := ChartData{
		Labels: fLabels,
		Datasets: []ChartDataset{{
			Label:           "By Fund",
			Data:            fData,
			BackgroundColor: fColors,
			BorderWidth:     1,
		}},
	}

	// Donor count trend
	rows3, err := s.db.Query(ctx, `
		SELECT TO_CHAR(donated_at, 'YYYY-MM') as month, COUNT(DISTINCT person_id)
		FROM donations WHERE tenant_id = $1 AND donated_at >= $2 AND status = 'completed' AND person_id IS NOT NULL
		GROUP BY month ORDER BY month`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("donor trend: %w", err)
	}
	defer rows3.Close()

	dLabels := []string{}
	dData := []float64{}
	for rows3.Next() {
		var month string
		var cnt int
		if err := rows3.Scan(&month, &cnt); err != nil {
			return nil, err
		}
		dLabels = append(dLabels, month)
		dData = append(dData, float64(cnt))
	}

	donorTrend := ChartData{
		Labels: dLabels,
		Datasets: []ChartDataset{{
			Label:       "Unique Donors",
			Data:        dData,
			BorderColor: "#8FBCB0",
			Tension:     0.3,
		}},
	}

	// KPIs
	year := time.Now().Year()
	var ytdCents int64
	var avgCents int64
	var donorCount int
	var retentionRate float64

	s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents),0), COALESCE(AVG(amount_cents),0)::BIGINT, COUNT(DISTINCT person_id)
		FROM donations WHERE tenant_id = $1 AND EXTRACT(YEAR FROM donated_at) = $2 AND status = 'completed'`,
		tenantID, year).Scan(&ytdCents, &avgCents, &donorCount)

	// Retention: donors who gave both this year and last year / donors last year
	s.db.QueryRow(ctx, `
		WITH this_year AS (SELECT DISTINCT person_id FROM donations WHERE tenant_id=$1 AND EXTRACT(YEAR FROM donated_at)=$2 AND status='completed' AND person_id IS NOT NULL),
		     last_year AS (SELECT DISTINCT person_id FROM donations WHERE tenant_id=$1 AND EXTRACT(YEAR FROM donated_at)=$3 AND status='completed' AND person_id IS NOT NULL)
		SELECT CASE WHEN COUNT(ly.person_id) = 0 THEN 0 ELSE (COUNT(ty.person_id)::FLOAT / COUNT(ly.person_id)::FLOAT * 100) END
		FROM last_year ly LEFT JOIN this_year ty ON ty.person_id = ly.person_id`,
		tenantID, year, year-1).Scan(&retentionRate)

	monthTotal := 0.0
	if len(mData) > 0 {
		monthTotal = mData[len(mData)-1]
	}

	return &GivingReport{
		MonthlyTotals: monthlyTotals,
		ByFund:        byFund,
		DonorTrend:    donorTrend,
		KPIs: []KPI{
			{Label: "This Month", Value: fmt.Sprintf("$%.0f", monthTotal)},
			{Label: "Year to Date", Value: fmt.Sprintf("$%s", formatMoney(float64(ytdCents)/100.0))},
			{Label: "Average Gift", Value: fmt.Sprintf("$%.2f", float64(avgCents)/100.0)},
			{Label: "Donor Retention", Value: fmt.Sprintf("%.0f%%", retentionRate)},
		},
	}, nil
}

// ---- Growth ----

func (s *Service) GetGrowthReport(ctx context.Context, tenantID uuid.UUID, rangeStr string) (*GrowthReport, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	since, _ := parseRange(rangeStr)
	if rangeStr == "" {
		since, _ = parseRange("12m")
	}

	// Cumulative membership growth
	rows, err := s.db.Query(ctx, `
		WITH months AS (
			SELECT DATE_TRUNC('month', d)::DATE as month
			FROM generate_series($2::timestamp, NOW(), INTERVAL '1 month') d
		)
		SELECT TO_CHAR(m.month, 'YYYY-MM'), COUNT(p.id)
		FROM months m LEFT JOIN people p ON p.tenant_id = $1 AND p.created_at <= m.month + INTERVAL '1 month'
		GROUP BY m.month ORDER BY m.month`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("growth cumulative: %w", err)
	}
	defer rows.Close()

	gLabels := []string{}
	gData := []float64{}
	for rows.Next() {
		var month string
		var cnt int
		if err := rows.Scan(&month, &cnt); err != nil {
			return nil, err
		}
		gLabels = append(gLabels, month)
		gData = append(gData, float64(cnt))
	}

	membershipGrowth := ChartData{
		Labels: gLabels,
		Datasets: []ChartDataset{{
			Label:           "Total Members",
			Data:            gData,
			BorderColor:     "#4A8B8C",
			BackgroundColor: "rgba(74,139,140,0.15)",
			Tension:         0.3,
			Fill:            true,
		}},
	}

	// New members by month
	rows2, err := s.db.Query(ctx, `
		SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*)
		FROM people WHERE tenant_id = $1 AND created_at >= $2
		GROUP BY month ORDER BY month`, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("growth new: %w", err)
	}
	defer rows2.Close()

	nLabels := []string{}
	nData := []float64{}
	for rows2.Next() {
		var month string
		var cnt int
		if err := rows2.Scan(&month, &cnt); err != nil {
			return nil, err
		}
		nLabels = append(nLabels, month)
		nData = append(nData, float64(cnt))
	}

	newByMonth := ChartData{
		Labels: nLabels,
		Datasets: []ChartDataset{{
			Label:           "New Members",
			Data:            nData,
			BackgroundColor: "#8FBCB0",
			BorderColor:     "#1B3A4B",
			BorderWidth:     1,
		}},
	}

	// Funnel: visitor → regular → member
	var visitors, regulars, members int
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE tenant_id=$1 AND membership_status='visitor'`, tenantID).Scan(&visitors)
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE tenant_id=$1 AND membership_status='regular'`, tenantID).Scan(&regulars)
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE tenant_id=$1 AND membership_status='member'`, tenantID).Scan(&members)

	totalPeople := visitors + regulars + members
	pct := func(n int) float64 {
		if totalPeople == 0 {
			return 0
		}
		return float64(n) / float64(totalPeople) * 100
	}

	funnel := []FunnelStep{
		{Label: "Visitor", Count: visitors, Pct: pct(visitors)},
		{Label: "Regular", Count: regulars, Pct: pct(regulars)},
		{Label: "Member", Count: members, Pct: pct(members)},
	}

	totalNow := 0.0
	if len(gData) > 0 {
		totalNow = gData[len(gData)-1]
	}
	growthRate := 0.0
	if len(gData) >= 2 && gData[0] > 0 {
		growthRate = ((gData[len(gData)-1] - gData[0]) / gData[0]) * 100
	}

	return &GrowthReport{
		MembershipGrowth: membershipGrowth,
		NewByMonth:       newByMonth,
		Funnel:           funnel,
		KPIs: []KPI{
			{Label: "Total Members", Value: fmt.Sprintf("%.0f", totalNow)},
			{Label: "Growth Rate", Value: fmt.Sprintf("%.1f%%", growthRate)},
			{Label: "New This Month", Value: fmt.Sprintf("%.0f", lastVal(nData))},
		},
	}, nil
}

// ---- Songs ----

func (s *Service) GetSongsReport(ctx context.Context, tenantID uuid.UUID) (*SongsReport, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	// Top 20 songs by usage
	rows, err := s.db.Query(ctx, `
		SELECT s.title, COUNT(si.id) as usage_count
		FROM songs s
		LEFT JOIN service_items si ON si.song_id = s.id
		WHERE s.tenant_id = $1
		GROUP BY s.id, s.title
		ORDER BY usage_count DESC
		LIMIT 20`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("top songs: %w", err)
	}
	defer rows.Close()

	tLabels := []string{}
	tData := []float64{}
	for rows.Next() {
		var title string
		var cnt int
		if err := rows.Scan(&title, &cnt); err != nil {
			return nil, err
		}
		tLabels = append(tLabels, title)
		tData = append(tData, float64(cnt))
	}

	topSongs := ChartData{
		Labels: tLabels,
		Datasets: []ChartDataset{{
			Label:           "Times Used",
			Data:            tData,
			BackgroundColor: "#4A8B8C",
			BorderWidth:     1,
		}},
	}

	// Songs by key
	rows2, err := s.db.Query(ctx, `
		SELECT COALESCE(NULLIF(default_key, ''), 'Unknown') as song_key, COUNT(*)
		FROM songs WHERE tenant_id = $1
		GROUP BY song_key ORDER BY COUNT(*) DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("songs by key: %w", err)
	}
	defer rows2.Close()

	kLabels := []string{}
	kData := []float64{}
	kColors := []string{}
	i := 0
	for rows2.Next() {
		var key string
		var cnt int
		if err := rows2.Scan(&key, &cnt); err != nil {
			return nil, err
		}
		kLabels = append(kLabels, key)
		kData = append(kData, float64(cnt))
		kColors = append(kColors, brandColors[i%len(brandColors)])
		i++
	}

	byKey := ChartData{
		Labels: kLabels,
		Datasets: []ChartDataset{{
			Label:           "Songs",
			Data:            kData,
			BackgroundColor: kColors,
			BorderWidth:     1,
		}},
	}

	// Unused songs (not used in 6+ months)
	rows3, err := s.db.Query(ctx, `
		SELECT s.id, s.title, COALESCE(s.artist, '') as artist,
			COALESCE(MAX(sv.service_date)::TEXT, 'Never')
		FROM songs s
		LEFT JOIN service_items si ON si.song_id = s.id
		LEFT JOIN services sv ON sv.id = si.service_id
		WHERE s.tenant_id = $1
		GROUP BY s.id, s.title, s.artist
		HAVING MAX(sv.service_date) IS NULL OR MAX(sv.service_date) < NOW() - INTERVAL '6 months'
		ORDER BY MAX(sv.service_date) ASC NULLS FIRST
		LIMIT 50`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("unused songs: %w", err)
	}
	defer rows3.Close()

	unused := []UnusedSong{}
	for rows3.Next() {
		var u UnusedSong
		if err := rows3.Scan(&u.ID, &u.Title, &u.Artist, &u.LastUsed); err != nil {
			return nil, err
		}
		unused = append(unused, u)
	}

	// Stats
	var totalSongs, uniqueUsed int
	var avgPerService float64
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM songs WHERE tenant_id=$1`, tenantID).Scan(&totalSongs)
	s.db.QueryRow(ctx, `SELECT COUNT(DISTINCT si.song_id) FROM service_items si JOIN services sv ON sv.id = si.service_id WHERE sv.tenant_id=$1 AND si.song_id IS NOT NULL`, tenantID).Scan(&uniqueUsed)
	s.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(cnt), 0) FROM (
			SELECT si.service_id, COUNT(*) as cnt FROM service_items si
			JOIN services sv ON sv.id = si.service_id
			WHERE sv.tenant_id=$1 AND si.song_id IS NOT NULL GROUP BY si.service_id
		) sub`, tenantID).Scan(&avgPerService)

	return &SongsReport{
		TopSongs:    topSongs,
		ByKey:       byKey,
		UnusedSongs: unused,
		KPIs: []KPI{
			{Label: "Total Songs", Value: fmt.Sprintf("%d", totalSongs)},
			{Label: "Unique Songs Used", Value: fmt.Sprintf("%d", uniqueUsed)},
			{Label: "Avg Songs/Service", Value: fmt.Sprintf("%.1f", avgPerService)},
			{Label: "Unused (6+ months)", Value: fmt.Sprintf("%d", len(unused))},
		},
	}, nil
}

// ---- Engagement ----

func (s *Service) GetEngagementReport(ctx context.Context, tenantID uuid.UUID) (*EngagementReport, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	// Distribution by tier
	rows, err := s.db.Query(ctx, `
		SELECT 
			CASE 
				WHEN score >= 80 THEN 'Core'
				WHEN score >= 60 THEN 'Active'
				WHEN score >= 40 THEN 'Attending'
				WHEN score >= 20 THEN 'Occasional'
				ELSE 'Inactive'
			END as tier,
			COUNT(*)
		FROM engagement_scores WHERE tenant_id = $1
		GROUP BY tier
		ORDER BY MIN(score) DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("engagement dist: %w", err)
	}
	defer rows.Close()

	tierOrder := []string{"Core", "Active", "Attending", "Occasional", "Inactive"}
	tierCounts := make(map[string]float64)
	var totalPeople int
	for rows.Next() {
		var tier string
		var cnt int
		if err := rows.Scan(&tier, &cnt); err != nil {
			return nil, err
		}
		tierCounts[tier] = float64(cnt)
		totalPeople += cnt
	}

	distLabels := []string{}
	distData := []float64{}
	distColors := []string{"#1B3A4B", "#4A8B8C", "#8FBCB0", "#6BA3A4", "#B5D9CE"}
	for _, t := range tierOrder {
		distLabels = append(distLabels, t)
		distData = append(distData, tierCounts[t])
	}

	distribution := ChartData{
		Labels: distLabels,
		Datasets: []ChartDataset{{
			Label:           "People",
			Data:            distData,
			BackgroundColor: distColors,
			BorderWidth:     1,
		}},
	}

	// Trend (monthly avg engagement score) - last 12 months
	rows2, err := s.db.Query(ctx, `
		SELECT TO_CHAR(calculated_at, 'YYYY-MM') as month, AVG(score)
		FROM engagement_scores WHERE tenant_id = $1 AND calculated_at >= NOW() - INTERVAL '12 months'
		GROUP BY month ORDER BY month`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("engagement trend: %w", err)
	}
	defer rows2.Close()

	eLabels := []string{}
	eData := []float64{}
	for rows2.Next() {
		var month string
		var avg float64
		if err := rows2.Scan(&month, &avg); err != nil {
			return nil, err
		}
		eLabels = append(eLabels, month)
		eData = append(eData, math.Round(avg*10)/10)
	}

	trend := ChartData{
		Labels: eLabels,
		Datasets: []ChartDataset{{
			Label:       "Avg Engagement Score",
			Data:        eData,
			BorderColor: "#4A8B8C",
			Tension:     0.3,
		}},
	}

	avgScore := 0.0
	if len(eData) > 0 {
		avgScore = eData[len(eData)-1]
	}

	coreCount := tierCounts["Core"] + tierCounts["Active"]

	return &EngagementReport{
		Distribution: distribution,
		Trend:        trend,
		KPIs: []KPI{
			{Label: "Avg Score", Value: fmt.Sprintf("%.1f", avgScore)},
			{Label: "Tracked People", Value: fmt.Sprintf("%d", totalPeople)},
			{Label: "Core + Active", Value: fmt.Sprintf("%.0f", coreCount)},
		},
	}, nil
}

// helpers

func formatMoney(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func lastVal(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	return data[len(data)-1]
}
