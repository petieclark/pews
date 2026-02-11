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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// ========== Stations ==========

func (s *Service) ListStations(ctx context.Context, tenantID string) ([]Station, error) {
	if err := s.setTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT id, tenant_id, name, location, is_active, created_at, updated_at FROM checkin_stations ORDER BY name`)
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
	err := s.db.QueryRow(ctx, `UPDATE checkin_stations SET name = $1, location = $2, is_active = $3 WHERE id = $4 RETURNING created_at, updated_at`,
		st.Name, st.Location, st.IsActive, stationID).Scan(&st.CreatedAt, &st.UpdatedAt)
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
	rows, err := s.db.Query(ctx, `SELECT e.id, e.tenant_id, e.name, e.event_date, e.service_id, e.station_id, e.is_active, e.created_at, e.updated_at, COUNT(c.id) as checkin_count FROM checkin_events e LEFT JOIN checkins c ON c.event_id = e.id GROUP BY e.id ORDER BY e.event_date DESC, e.created_at DESC`)
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
	err := s.db.QueryRow(ctx, `UPDATE checkin_events SET name = $1, event_date = $2, service_id = $3, station_id = $4, is_active = $5 WHERE id = $6 RETURNING created_at, updated_at`,
		ev.Name, ev.EventDate, ev.ServiceID, ev.StationID, ev.IsActive, eventID).Scan(&ev.CreatedAt, &ev.UpdatedAt)
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
	err := s.db.QueryRow(ctx, `SELECT e.id, e.tenant_id, e.name, e.event_date, e.service_id, e.station_id, e.is_active, e.created_at, e.updated_at, COUNT(c.id) as checkin_count FROM checkin_events e LEFT JOIN checkins c ON c.event_id = e.id WHERE e.id = $1 GROUP BY e.id`, eventID).Scan(
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
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM checkins WHERE person_id = $1`, personID).Scan(&previousCount)
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
	result, err := s.db.Exec(ctx, `UPDATE checkins SET checked_out_at = $1 WHERE event_id = $2 AND person_id = $3 AND checked_out_at IS NULL`, now, eventID, personID)
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
	rows, err := s.db.Query(ctx, `SELECT c.id, c.tenant_id, c.event_id, c.person_id, c.station_id, c.first_time, c.checked_in_at, c.checked_out_at, c.notes, c.created_at, COALESCE(p.first_name || ' ' || p.last_name, '') as person_name, COALESCE(p.email, '') as person_email, COALESCE(cs.name, '') as station_name FROM checkins c JOIN people p ON p.id = c.person_id LEFT JOIN checkin_stations cs ON cs.id = c.station_id WHERE c.event_id = $1 ORDER BY c.checked_in_at DESC`, eventID)
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
	rows, err := s.db.Query(ctx, `SELECT c.id, c.tenant_id, c.event_id, c.person_id, c.station_id, c.first_time, c.checked_in_at, c.checked_out_at, c.notes, c.created_at, '' as person_name, '' as person_email, COALESCE(cs.name, '') as station_name FROM checkins c LEFT JOIN checkin_stations cs ON cs.id = c.station_id WHERE c.person_id = $1 ORDER BY c.checked_in_at DESC LIMIT 100`, personID)
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
	rows, err := s.db.Query(ctx, `SELECT id, tenant_id, person_id, alert_type, severity, description, created_at, updated_at FROM medical_alerts WHERE person_id = $1 ORDER BY severity DESC, created_at DESC`, personID)
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
	result, err := s.db.Exec(ctx, "DELETE FROM medical_alerts WHERE id = $1", alertID)
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
	rows, err := s.db.Query(ctx, `SELECT ap.id, ap.tenant_id, ap.child_id, ap.pickup_person_id, ap.relationship, ap.is_active, ap.created_at, ap.updated_at, COALESCE(p.first_name || ' ' || p.last_name, '') as pickup_person_name FROM authorized_pickups ap JOIN people p ON p.id = ap.pickup_person_id WHERE ap.child_id = $1 ORDER BY ap.relationship, p.last_name`, childID)
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
	result, err := s.db.Exec(ctx, "DELETE FROM authorized_pickups WHERE id = $1", pickupID)
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
	err := s.db.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(CASE WHEN first_time THEN 1 ELSE 0 END), 0) FROM checkins c JOIN checkin_events e ON e.id = c.event_id WHERE e.event_date = $1`, today).Scan(&stats.TotalCheckins, &stats.FirstTimers)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	rows, err := s.db.Query(ctx, `SELECT COALESCE(c.station_id::text, 'unassigned'), COALESCE(cs.name, 'Unassigned'), COUNT(*) FROM checkins c JOIN checkin_events e ON e.id = c.event_id LEFT JOIN checkin_stations cs ON cs.id = c.station_id WHERE e.event_date = $1 GROUP BY c.station_id, cs.name ORDER BY COUNT(*) DESC`, today)
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
	rows, err := s.db.Query(ctx, `SELECT id, first_name, last_name, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(photo_url, '') FROM people WHERE (first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1 OR (first_name || ' ' || last_name) ILIKE $1) ORDER BY last_name, first_name LIMIT 20`, "%"+query+"%")
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
