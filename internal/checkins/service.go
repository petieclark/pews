package checkins

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) setTenant(ctx context.Context, tenantID string) error {
	// RLS removed - no longer needed
	return nil
}

// ========== Stations ==========

func (s *Service) ListStations(ctx context.Context, tenantID string) ([]Station, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT id, tenant_id, name, COALESCE(location, ''), is_active, created_at, updated_at FROM checkin_stations WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list stations: %w", err)
	}
	defer rows.Close()
	stations := []Station{}
	for rows.Next() {
		var st Station
		if err := rows.Scan(&st.ID, &st.TenantID, &st.Name, &st.Location, &st.IsActive, &st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, err
		}
		stations = append(stations, st)
	}
	return stations, nil
}

func (s *Service) CreateStation(ctx context.Context, tenantID string, st *Station) (*Station, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	st.ID = uuid.New().String()
	st.TenantID = tenantID
	err := s.db.QueryRow(ctx, `INSERT INTO checkin_stations (id, tenant_id, name, location, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`,
		st.ID, st.TenantID, st.Name, st.Location, st.IsActive).Scan(&st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create station: %w", err)
	}
	return st, nil
}

func (s *Service) UpdateStation(ctx context.Context, tenantID, stationID string, st *Station) (*Station, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	err := s.db.QueryRow(ctx, `UPDATE checkin_stations SET name = $1, location = $2, is_active = $3 WHERE id = $4 AND tenant_id = $5 RETURNING created_at, updated_at`,
		st.Name, st.Location, st.IsActive, stationID, tenantID).Scan(&st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("station not found")
		}
		return nil, fmt.Errorf("failed to update station: %w", err)
	}
	st.ID = stationID
	st.TenantID = tenantID
	return st, nil
}

// ========== Events ==========

func (s *Service) ListEvents(ctx context.Context, tenantID string) ([]Event, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT e.id, e.tenant_id, e.name, e.event_date, e.service_id, e.station_id, e.is_active, e.created_at, e.updated_at, COUNT(c.id) as checkin_count FROM checkin_events e LEFT JOIN checkins c ON c.event_id = e.id WHERE e.tenant_id = $1 GROUP BY e.id ORDER BY e.event_date DESC, e.created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		var ev Event
		if err := rows.Scan(&ev.ID, &ev.TenantID, &ev.Name, &ev.EventDate, &ev.ServiceID, &ev.StationID, &ev.IsActive, &ev.CreatedAt, &ev.UpdatedAt, &ev.CheckinCount); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, nil
}

func (s *Service) CreateEvent(ctx context.Context, tenantID string, ev *Event) (*Event, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	ev.ID = uuid.New().String()
	ev.TenantID = tenantID
	err := s.db.QueryRow(ctx, `INSERT INTO checkin_events (id, tenant_id, name, event_date, service_id, station_id, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`,
		ev.ID, ev.TenantID, ev.Name, ev.EventDate, ev.ServiceID, ev.StationID, ev.IsActive).Scan(&ev.CreatedAt, &ev.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return ev, nil
}

func (s *Service) UpdateEvent(ctx context.Context, tenantID, eventID string, ev *Event) (*Event, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	err := s.db.QueryRow(ctx, `UPDATE checkin_events SET name = $1, event_date = $2, service_id = $3, station_id = $4, is_active = $5 WHERE id = $6 AND tenant_id = $7 RETURNING created_at, updated_at`,
		ev.Name, ev.EventDate, ev.ServiceID, ev.StationID, ev.IsActive, eventID, tenantID).Scan(&ev.CreatedAt, &ev.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to update event: %w", err)
	}
	ev.ID = eventID
	ev.TenantID = tenantID
	return ev, nil
}

func (s *Service) GetEvent(ctx context.Context, tenantID, eventID string) (*Event, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	var ev Event
	err := s.db.QueryRow(ctx, `SELECT e.id, e.tenant_id, e.name, e.event_date, e.service_id, e.station_id, e.is_active, e.created_at, e.updated_at, COUNT(c.id) as checkin_count FROM checkin_events e LEFT JOIN checkins c ON c.event_id = e.id WHERE e.id = $1 AND e.tenant_id = $2 GROUP BY e.id`, eventID, tenantID).Scan(
		&ev.ID, &ev.TenantID, &ev.Name, &ev.EventDate, &ev.ServiceID, &ev.StationID, &ev.IsActive, &ev.CreatedAt, &ev.UpdatedAt, &ev.CheckinCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return &ev, nil
}

// ========== Check-ins ==========

func (s *Service) CheckIn(ctx context.Context, tenantID, eventID, personID string, stationID *string, notes string) (*CheckinResult, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	var previousCount int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM checkins WHERE tenant_id = $1 AND person_id = $2`, tenantID, personID).Scan(&previousCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check history: %w", err)
	}
	firstTime := previousCount == 0
	checkinID := uuid.New().String()
	now := time.Now()
	_, err = s.db.Exec(ctx, `INSERT INTO checkins (id, tenant_id, event_id, person_id, station_id, first_time, checked_in_at, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		checkinID, tenantID, eventID, personID, stationID, firstTime, now, notes)
	if err != nil {
		return nil, fmt.Errorf("failed to check in: %w", err)
	}
	checkin := &Checkin{ID: checkinID, TenantID: tenantID, EventID: eventID, PersonID: personID, StationID: stationID, FirstTime: firstTime, CheckedInAt: now, Notes: notes, CreatedAt: now}
	alerts, _ := s.GetAlerts(ctx, tenantID, personID)
	if alerts == nil {
		alerts = []MedicalAlert{}
	}
	return &CheckinResult{Checkin: checkin, FirstTime: firstTime, MedicalAlerts: alerts}, nil
}

func (s *Service) CheckOut(ctx context.Context, tenantID, eventID, personID string) error {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return err
	}
	now := time.Now()
	result, err := s.db.Exec(ctx, `UPDATE checkins SET checked_out_at = $1 WHERE tenant_id = $2 AND event_id = $3 AND person_id = $4 AND checked_out_at IS NULL`, now, tenantID, eventID, personID)
	if err != nil {
		return fmt.Errorf("failed to check out: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no active check-in found")
	}
	return nil
}

func (s *Service) GetAttendees(ctx context.Context, tenantID, eventID string) ([]Checkin, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT c.id, c.tenant_id, c.event_id, c.person_id, c.station_id, c.first_time, c.checked_in_at, c.checked_out_at, COALESCE(c.notes, ''), c.created_at, COALESCE(p.first_name || ' ' || p.last_name, '') as person_name, COALESCE(p.email, '') as person_email, COALESCE(cs.name, '') as station_name FROM checkins c JOIN people p ON p.id = c.person_id LEFT JOIN checkin_stations cs ON cs.id = c.station_id WHERE c.event_id = $1 AND c.tenant_id = $2 ORDER BY c.checked_in_at DESC`, eventID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendees: %w", err)
	}
	defer rows.Close()
	attendees := []Checkin{}
	for rows.Next() {
		var ci Checkin
		if err := rows.Scan(&ci.ID, &ci.TenantID, &ci.EventID, &ci.PersonID, &ci.StationID, &ci.FirstTime, &ci.CheckedInAt, &ci.CheckedOutAt, &ci.Notes, &ci.CreatedAt, &ci.PersonName, &ci.PersonEmail, &ci.StationName); err != nil {
			return nil, err
		}
		attendees = append(attendees, ci)
	}
	return attendees, nil
}

func (s *Service) GetPersonHistory(ctx context.Context, tenantID, personID string) ([]Checkin, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT c.id, c.tenant_id, c.event_id, c.person_id, c.station_id, c.first_time, c.checked_in_at, c.checked_out_at, COALESCE(c.notes, ''), c.created_at, COALESCE(e.name, '') as person_name, '' as person_email, COALESCE(cs.name, '') as station_name FROM checkins c LEFT JOIN checkin_stations cs ON cs.id = c.station_id LEFT JOIN checkin_events e ON e.id = c.event_id WHERE c.person_id = $1 AND c.tenant_id = $2 ORDER BY c.checked_in_at DESC LIMIT 100`, personID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person history: %w", err)
	}
	defer rows.Close()
	history := []Checkin{}
	for rows.Next() {
		var ci Checkin
		if err := rows.Scan(&ci.ID, &ci.TenantID, &ci.EventID, &ci.PersonID, &ci.StationID, &ci.FirstTime, &ci.CheckedInAt, &ci.CheckedOutAt, &ci.Notes, &ci.CreatedAt, &ci.PersonName, &ci.PersonEmail, &ci.StationName); err != nil {
			return nil, err
		}
		history = append(history, ci)
	}
	return history, nil
}

// ========== Medical Alerts ==========

func (s *Service) GetAlerts(ctx context.Context, tenantID, personID string) ([]MedicalAlert, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT id, tenant_id, person_id, alert_type, severity, description, created_at, updated_at FROM medical_alerts WHERE tenant_id = $1 AND person_id = $2 ORDER BY severity DESC, created_at DESC`, tenantID, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()
	alerts := []MedicalAlert{}
	for rows.Next() {
		var a MedicalAlert
		if err := rows.Scan(&a.ID, &a.TenantID, &a.PersonID, &a.AlertType, &a.Severity, &a.Description, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (s *Service) CreateAlert(ctx context.Context, tenantID string, a *MedicalAlert) (*MedicalAlert, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	a.ID = uuid.New().String()
	a.TenantID = tenantID
	err := s.db.QueryRow(ctx, `INSERT INTO medical_alerts (id, tenant_id, person_id, alert_type, severity, description) VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`,
		a.ID, a.TenantID, a.PersonID, a.AlertType, a.Severity, a.Description).Scan(&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}
	return a, nil
}

func (s *Service) DeleteAlert(ctx context.Context, tenantID, alertID string) error {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return err
	}
	result, err := s.db.Exec(ctx, "DELETE FROM medical_alerts WHERE id = $1 AND tenant_id = $2", alertID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found")
	}
	return nil
}

// ========== Authorized Pickups ==========

func (s *Service) GetPickups(ctx context.Context, tenantID, childID string) ([]AuthorizedPickup, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT ap.id, ap.tenant_id, ap.child_id, ap.pickup_person_id, ap.relationship, ap.is_active, ap.created_at, ap.updated_at, COALESCE(p.first_name || ' ' || p.last_name, '') as pickup_person_name FROM authorized_pickups ap JOIN people p ON p.id = ap.pickup_person_id WHERE ap.tenant_id = $1 AND ap.child_id = $2 ORDER BY ap.relationship, p.last_name`, tenantID, childID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pickups: %w", err)
	}
	defer rows.Close()
	pickups := []AuthorizedPickup{}
	for rows.Next() {
		var p AuthorizedPickup
		if err := rows.Scan(&p.ID, &p.TenantID, &p.ChildID, &p.PickupPersonID, &p.Relationship, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.PickupPersonName); err != nil {
			return nil, err
		}
		pickups = append(pickups, p)
	}
	return pickups, nil
}

func (s *Service) CreatePickup(ctx context.Context, tenantID string, p *AuthorizedPickup) (*AuthorizedPickup, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	p.ID = uuid.New().String()
	p.TenantID = tenantID
	err := s.db.QueryRow(ctx, `INSERT INTO authorized_pickups (id, tenant_id, child_id, pickup_person_id, relationship, is_active) VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`,
		p.ID, p.TenantID, p.ChildID, p.PickupPersonID, p.Relationship, p.IsActive).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create pickup: %w", err)
	}
	return p, nil
}

func (s *Service) DeletePickup(ctx context.Context, tenantID, pickupID string) error {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return err
	}
	result, err := s.db.Exec(ctx, "DELETE FROM authorized_pickups WHERE id = $1 AND tenant_id = $2", pickupID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete pickup: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("pickup not found")
	}
	return nil
}

// ========== Stats ==========

func (s *Service) GetTodayStats(ctx context.Context, tenantID string) (*Stats, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	stats := &Stats{}
	err := s.db.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(CASE WHEN first_time THEN 1 ELSE 0 END), 0) FROM checkins c JOIN checkin_events e ON e.id = c.event_id WHERE c.tenant_id = $1 AND e.event_date = $2`, tenantID, today).Scan(&stats.TotalCheckins, &stats.FirstTimers)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	rows, err := s.db.Query(ctx, `SELECT COALESCE(c.station_id::text, 'unassigned'), COALESCE(cs.name, 'Unassigned'), COUNT(*) FROM checkins c JOIN checkin_events e ON e.id = c.event_id LEFT JOIN checkin_stations cs ON cs.id = c.station_id WHERE c.tenant_id = $1 AND e.event_date = $2 GROUP BY c.station_id, cs.name ORDER BY COUNT(*) DESC`, tenantID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get station stats: %w", err)
	}
	defer rows.Close()
	stats.ByStation = []StationStat{}
	for rows.Next() {
		var ss StationStat
		if err := rows.Scan(&ss.StationID, &ss.StationName, &ss.Count); err != nil {
			return nil, err
		}
		stats.ByStation = append(stats.ByStation, ss)
	}
	return stats, nil
}

// ========== Search ==========

func (s *Service) SearchPeople(ctx context.Context, tenantID, query string) ([]PersonSearchResult, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT id, first_name, last_name, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(photo_url, '') FROM people WHERE tenant_id = $1 AND (first_name ILIKE $2 OR last_name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2 OR (first_name || ' ' || last_name) ILIKE $2) ORDER BY last_name, first_name LIMIT 20`, tenantID, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search people: %w", err)
	}
	defer rows.Close()
	results := []PersonSearchResult{}
	for rows.Next() {
		var p PersonSearchResult
		if err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Email, &p.Phone, &p.PhotoURL); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, nil
}

// ========== Attendance Tracking ==========

func (s *Service) GetAttendanceTrends(ctx context.Context, tenantID, period string, limit int) ([]AttendanceTrend, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	
	var query string
	if period == "weekly" {
		query = `SELECT 
			to_char(e.event_date::date, 'IYYY-IW') as period,
			COUNT(c.id) as count
		FROM checkin_events e
		LEFT JOIN checkins c ON c.event_id = e.id
		WHERE e.tenant_id = $1 AND e.event_date >= CURRENT_DATE - INTERVAL '12 weeks'
		GROUP BY period
		ORDER BY period DESC
		LIMIT $2`
	} else {
		query = `SELECT 
			to_char(e.event_date::date, 'YYYY-MM') as period,
			COUNT(c.id) as count
		FROM checkin_events e
		LEFT JOIN checkins c ON c.event_id = e.id
		WHERE e.tenant_id = $1 AND e.event_date >= CURRENT_DATE - INTERVAL '12 months'
		GROUP BY period
		ORDER BY period DESC
		LIMIT $2`
	}
	
	rows, err := s.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance trends: %w", err)
	}
	defer rows.Close()
	
	trends := []AttendanceTrend{}
	for rows.Next() {
		var t AttendanceTrend
		if err := rows.Scan(&t.Period, &t.Count); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	return trends, nil
}

func (s *Service) GetPersonAttendance(ctx context.Context, tenantID, personID string) (*PersonAttendance, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	
	attendance := &PersonAttendance{}
	
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM checkins WHERE tenant_id = $1 AND person_id = $2`, tenantID, personID).Scan(&attendance.TotalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	
	var lastAttended *time.Time
	err = s.db.QueryRow(ctx, `SELECT MAX(checked_in_at) FROM checkins WHERE tenant_id = $1 AND person_id = $2`, tenantID, personID).Scan(&lastAttended)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get last attended: %w", err)
	}
	if lastAttended != nil {
		formatted := lastAttended.Format("2006-01-02")
		attendance.LastAttended = &formatted
	}
	
	attendance.Streak = s.calculateStreak(ctx, tenantID, personID)
	
	rows, err := s.db.Query(ctx, `
		SELECT 
			DATE(c.checked_in_at) as date,
			COUNT(*) as count
		FROM checkins c
		WHERE c.tenant_id = $1 AND c.person_id = $2 
		AND c.checked_in_at >= CURRENT_DATE - INTERVAL '365 days'
		GROUP BY DATE(c.checked_in_at)
		ORDER BY date DESC
	`, tenantID, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance data: %w", err)
	}
	defer rows.Close()
	
	attendance.AttendanceData = []AttendanceTrend{}
	for rows.Next() {
		var t AttendanceTrend
		if err := rows.Scan(&t.Period, &t.Count); err != nil {
			return nil, err
		}
		attendance.AttendanceData = append(attendance.AttendanceData, t)
	}
	
	attendance.RecentCheckins, err = s.GetPersonHistory(ctx, tenantID, personID)
	if err != nil {
		return nil, err
	}
	if len(attendance.RecentCheckins) > 10 {
		attendance.RecentCheckins = attendance.RecentCheckins[:10]
	}
	
	return attendance, nil
}

func (s *Service) calculateStreak(ctx context.Context, tenantID, personID string) int {
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT DATE(checked_in_at) as date
		FROM checkins
		WHERE tenant_id = $1 AND person_id = $2
		ORDER BY date DESC
	`, tenantID, personID)
	if err != nil {
		return 0
	}
	defer rows.Close()
	
	dates := []time.Time{}
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return 0
		}
		dates = append(dates, date)
	}
	
	if len(dates) == 0 {
		return 0
	}
	
	streak := 1
	for i := 0; i < len(dates)-1; i++ {
		diff := dates[i].Sub(dates[i+1]).Hours() / 24
		if diff <= 7 {
			streak++
		} else {
			break
		}
	}
	
	return streak
}

func (s *Service) GetServiceAttendance(ctx context.Context, tenantID, serviceID string) (*ServiceAttendance, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	
	sa := &ServiceAttendance{ServiceID: serviceID}
	
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(c.id)
		FROM checkins c
		JOIN checkin_events e ON e.id = c.event_id
		WHERE c.tenant_id = $1 AND e.service_id = $2
	`, tenantID, serviceID).Scan(&sa.AttendanceCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance count: %w", err)
	}
	
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(event_count), 0)
		FROM (
			SELECT e.id, COUNT(c.id) as event_count
			FROM checkin_events e
			LEFT JOIN checkins c ON c.event_id = e.id
			WHERE e.tenant_id = $1 AND e.service_id = $2
			GROUP BY e.id
		) subq
	`, tenantID, serviceID).Scan(&sa.AverageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get average: %w", err)
	}
	
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT ON (c.person_id)
			c.id, c.tenant_id, c.event_id, c.person_id, c.station_id, 
			c.first_time, c.checked_in_at, c.checked_out_at, COALESCE(c.notes, ''), c.created_at,
			COALESCE(p.first_name || ' ' || p.last_name, '') as person_name,
			COALESCE(p.email, '') as person_email,
			COALESCE(cs.name, '') as station_name
		FROM checkins c
		JOIN checkin_events e ON e.id = c.event_id
		JOIN people p ON p.id = c.person_id
		LEFT JOIN checkin_stations cs ON cs.id = c.station_id
		WHERE c.tenant_id = $1 AND e.service_id = $2
		ORDER BY c.person_id, c.checked_in_at DESC
	`, tenantID, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendees: %w", err)
	}
	defer rows.Close()
	
	sa.Attendees = []Checkin{}
	for rows.Next() {
		var ci Checkin
		if err := rows.Scan(&ci.ID, &ci.TenantID, &ci.EventID, &ci.PersonID, &ci.StationID, 
			&ci.FirstTime, &ci.CheckedInAt, &ci.CheckedOutAt, &ci.Notes, &ci.CreatedAt,
			&ci.PersonName, &ci.PersonEmail, &ci.StationName); err != nil {
			return nil, err
		}
		sa.Attendees = append(sa.Attendees, ci)
	}
	
	return sa, nil
}

func (s *Service) GetFirstTimersThisWeek(ctx context.Context, tenantID string) ([]FirstTimer, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	
	rows, err := s.db.Query(ctx, `
		SELECT 
			c.id, c.person_id, 
			COALESCE(p.first_name || ' ' || p.last_name, '') as person_name,
			COALESCE(p.email, '') as person_email,
			e.name as event_name,
			c.checked_in_at
		FROM checkins c
		JOIN people p ON p.id = c.person_id
		JOIN checkin_events e ON e.id = c.event_id
		WHERE c.tenant_id = $1 AND c.first_time = true
		AND c.checked_in_at >= date_trunc('week', CURRENT_DATE)
		ORDER BY c.checked_in_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get first timers: %w", err)
	}
	defer rows.Close()
	
	firstTimers := []FirstTimer{}
	for rows.Next() {
		var ft FirstTimer
		if err := rows.Scan(&ft.ID, &ft.PersonID, &ft.PersonName, &ft.PersonEmail, 
			&ft.EventName, &ft.CheckedInAt); err != nil {
			return nil, err
		}
		firstTimers = append(firstTimers, ft)
	}
	
	return firstTimers, nil
}
