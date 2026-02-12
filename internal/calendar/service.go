package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

const selectFields = `e.id, e.tenant_id, e.title, COALESCE(e.description, ''), COALESCE(e.location, ''),
	e.start_time, e.end_time, e.all_day, COALESCE(e.recurring, 'none'), COALESCE(e.event_type, 'other'),
	COALESCE(e.color, '#4A8B8C'), e.room_id, r.name as room_name,
	COALESCE(e.created_by::text, ''), e.created_at, e.updated_at`

func scanEvent(scan func(dest ...interface{}) error) (Event, error) {
	var e Event
	err := scan(&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
		&e.StartTime, &e.EndTime, &e.AllDay, &e.Recurring, &e.EventType,
		&e.Color, &e.RoomID, &e.RoomName,
		&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

// GetServicesAsEvents returns services within a date range as calendar events
func (s *Service) GetServicesAsEvents(ctx context.Context, tenantID, fromDate, toDate string) ([]Event, error) {
	query := `
		SELECT s.id, s.tenant_id, COALESCE(s.name, st.name) as title,
		       COALESCE(s.notes, '') as description, '' as location,
		       (s.service_date + COALESCE(s.service_time::time, '10:00'::time)) as start_time,
		       (s.service_date + COALESCE(s.service_time::time, '10:00'::time) + INTERVAL '1 hour') as end_time,
		       s.service_date, s.id as service_id,
		       COALESCE(st.color, '#4A8B8C') as color,
		       (SELECT COUNT(*) FROM checkins cr
		        JOIN checkin_events ce ON cr.event_id = ce.id
		        WHERE ce.service_id = s.id) as attendance_count
		FROM services s
		LEFT JOIN service_types st ON s.service_type_id = st.id
		WHERE s.tenant_id = $1`

	args := []interface{}{tenantID}
	argN := 1

	if fromDate != "" {
		argN++
		query += fmt.Sprintf(" AND s.service_date >= $%d::date", argN)
		args = append(args, fromDate)
	}
	if toDate != "" {
		argN++
		query += fmt.Sprintf(" AND s.service_date <= $%d::date", argN)
		args = append(args, toDate)
	}

	query += " ORDER BY s.service_date ASC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query services as events: %w", err)
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var e Event
		var serviceDate time.Time
		var serviceID string
		var attendCount int
		err := rows.Scan(&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
			&e.StartTime, &e.EndTime, &serviceDate, &serviceID, &e.Color, &attendCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service event: %w", err)
		}
		e.EventType = "service"
		e.Recurring = "none"
		e.ServiceID = &serviceID
		if attendCount > 0 {
			e.AttendanceCount = &attendCount
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// GetEventAttendanceCounts enriches events with check-in attendance counts
func (s *Service) GetEventAttendanceCounts(ctx context.Context, tenantID string, events []Event) []Event {
	for i := range events {
		var count int
		err := s.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM checkins cr
			JOIN checkin_events ce ON cr.event_id = ce.id
			WHERE ce.tenant_id = $1 AND ce.name = $2
			AND ce.event_date = $3::date`,
			tenantID, events[i].Title, events[i].StartTime).Scan(&count)
		if err == nil && count > 0 {
			events[i].AttendanceCount = &count
		}
	}
	return events
}

// GenerateRecurringInstances expands recurring events into individual instances for a date range
func (s *Service) GenerateRecurringInstances(events []Event, from, to time.Time) []Event {
	result := []Event{}
	for _, e := range events {
		if e.Recurring == "none" || e.Recurring == "" {
			result = append(result, e)
			continue
		}

		// Add original
		if !e.StartTime.Before(from) && !e.StartTime.After(to) {
			result = append(result, e)
		}

		var interval time.Duration
		switch e.Recurring {
		case "weekly":
			interval = 7 * 24 * time.Hour
		case "biweekly":
			interval = 14 * 24 * time.Hour
		case "monthly":
			// Handle monthly separately
			duration := e.EndTime.Sub(e.StartTime)
			current := e.StartTime
			for {
				current = current.AddDate(0, 1, 0)
				if current.After(to) {
					break
				}
				if current.Before(from) {
					continue
				}
				instance := e
				instance.StartTime = current
				instance.EndTime = current.Add(duration)
				result = append(result, instance)
			}
			continue
		default:
			result = append(result, e)
			continue
		}

		// Weekly/biweekly
		duration := e.EndTime.Sub(e.StartTime)
		current := e.StartTime
		for {
			current = current.Add(interval)
			if current.After(to) {
				break
			}
			if current.Before(from) {
				continue
			}
			instance := e
			instance.StartTime = current
			instance.EndTime = current.Add(duration)
			result = append(result, instance)
		}
	}
	return result
}

// ListEvents returns events within an optional date range
func (s *Service) ListEvents(ctx context.Context, tenantID, fromDate, toDate, eventType string, page, limit int) ([]Event, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`SELECT %s FROM events e LEFT JOIN rooms r ON e.room_id = r.id WHERE e.tenant_id = $1`, selectFields)
	args := []interface{}{tenantID}
	argCount := 1

	if fromDate != "" {
		argCount++
		query += fmt.Sprintf(" AND e.start_time >= $%d", argCount)
		args = append(args, fromDate)
	}
	if toDate != "" {
		argCount++
		query += fmt.Sprintf(" AND e.start_time <= $%d", argCount)
		args = append(args, toDate)
	}
	if eventType != "" {
		argCount++
		query += fmt.Sprintf(" AND e.event_type = $%d", argCount)
		args = append(args, eventType)
	}

	query += " ORDER BY e.start_time ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		e, err := scanEvent(rows.Scan)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	// Count
	countQuery := `SELECT COUNT(*) FROM events e WHERE e.tenant_id = $1`
	countArgs := []interface{}{tenantID}
	cN := 1
	if fromDate != "" {
		cN++
		countQuery += fmt.Sprintf(" AND e.start_time >= $%d", cN)
		countArgs = append(countArgs, fromDate)
	}
	if toDate != "" {
		cN++
		countQuery += fmt.Sprintf(" AND e.start_time <= $%d", cN)
		countArgs = append(countArgs, toDate)
	}
	if eventType != "" {
		cN++
		countQuery += fmt.Sprintf(" AND e.event_type = $%d", cN)
		countArgs = append(countArgs, eventType)
	}

	var total int
	_ = s.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)

	return events, total, nil
}

func (s *Service) GetEventByID(ctx context.Context, tenantID, eventID string) (*Event, error) {
	query := fmt.Sprintf(`SELECT %s FROM events e LEFT JOIN rooms r ON e.room_id = r.id WHERE e.id = $1 AND e.tenant_id = $2`, selectFields)
	e, err := scanEvent(s.db.QueryRow(ctx, query, eventID, tenantID).Scan)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return &e, nil
}

func (s *Service) CreateEvent(ctx context.Context, tenantID, userID string, event *Event) (*Event, error) {
	if event.Color == "" {
		if c, ok := EventTypeColors[event.EventType]; ok {
			event.Color = c
		} else {
			event.Color = "#4A8B8C"
		}
	}
	if event.EventType == "" {
		event.EventType = "other"
	}

	query := `INSERT INTO events (tenant_id, title, description, location, start_time, end_time,
	          all_day, recurring, event_type, color, room_id, created_by)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	          RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		tenantID, event.Title, event.Description, event.Location,
		event.StartTime, event.EndTime, event.AllDay, event.Recurring,
		event.EventType, event.Color, event.RoomID, userID,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	event.TenantID = tenantID
	event.CreatedBy = userID
	return event, nil
}

func (s *Service) UpdateEvent(ctx context.Context, tenantID, eventID string, event *Event) (*Event, error) {
	if event.Color == "" {
		if c, ok := EventTypeColors[event.EventType]; ok {
			event.Color = c
		} else {
			event.Color = "#4A8B8C"
		}
	}

	query := `UPDATE events SET title = $1, description = $2, location = $3,
	          start_time = $4, end_time = $5, all_day = $6, recurring = $7,
	          event_type = $8, color = $9, room_id = $10, updated_at = NOW()
	          WHERE id = $11 AND tenant_id = $12
	          RETURNING id, tenant_id, title, COALESCE(description, ''), COALESCE(location, ''),
	          start_time, end_time, all_day, COALESCE(recurring, 'none'), COALESCE(event_type, 'other'),
	          COALESCE(color, '#4A8B8C'), room_id, COALESCE(created_by::text, ''), created_at, updated_at`

	var e Event
	err := s.db.QueryRow(ctx, query,
		event.Title, event.Description, event.Location,
		event.StartTime, event.EndTime, event.AllDay, event.Recurring,
		event.EventType, event.Color, event.RoomID,
		eventID, tenantID,
	).Scan(&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
		&e.StartTime, &e.EndTime, &e.AllDay, &e.Recurring, &e.EventType,
		&e.Color, &e.RoomID, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return &e, nil
}

func (s *Service) DeleteEvent(ctx context.Context, tenantID, eventID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM events WHERE id = $1 AND tenant_id = $2`, eventID, tenantID)
	return err
}

func (s *Service) GetUpcomingEvents(ctx context.Context, tenantID string, limit int) ([]Event, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	query := fmt.Sprintf(`SELECT %s FROM events e LEFT JOIN rooms r ON e.room_id = r.id
	          WHERE e.tenant_id = $1 AND e.start_time >= NOW()
	          ORDER BY e.start_time ASC LIMIT $2`, selectFields)

	rows, err := s.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming events: %w", err)
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		e, err := scanEvent(rows.Scan)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func (s *Service) GenerateICal(ctx context.Context, tenantID string) (string, error) {
	events, _, err := s.ListEvents(ctx, tenantID, "", "", "", 1, 1000)
	if err != nil {
		return "", err
	}

	var ical strings.Builder
	ical.WriteString("BEGIN:VCALENDAR\r\n")
	ical.WriteString("VERSION:2.0\r\n")
	ical.WriteString("PRODID:-//Pews//Calendar//EN\r\n")
	ical.WriteString("CALSCALE:GREGORIAN\r\n")
	ical.WriteString("METHOD:PUBLISH\r\n")

	for _, event := range events {
		ical.WriteString("BEGIN:VEVENT\r\n")
		ical.WriteString(fmt.Sprintf("UID:%s\r\n", event.ID))
		ical.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", event.CreatedAt.UTC().Format("20060102T150405Z")))

		if event.AllDay {
			ical.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", event.StartTime.Format("20060102")))
			ical.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", event.EndTime.Format("20060102")))
		} else {
			ical.WriteString(fmt.Sprintf("DTSTART:%s\r\n", event.StartTime.UTC().Format("20060102T150405Z")))
			ical.WriteString(fmt.Sprintf("DTEND:%s\r\n", event.EndTime.UTC().Format("20060102T150405Z")))
		}

		ical.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(event.Title)))
		if event.Description != "" {
			ical.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(event.Description)))
		}
		if event.Location != "" {
			ical.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICalText(event.Location)))
		}

		if event.Recurring == "weekly" {
			ical.WriteString("RRULE:FREQ=WEEKLY\r\n")
		} else if event.Recurring == "biweekly" {
			ical.WriteString("RRULE:FREQ=WEEKLY;INTERVAL=2\r\n")
		} else if event.Recurring == "monthly" {
			ical.WriteString("RRULE:FREQ=MONTHLY\r\n")
		}

		ical.WriteString("END:VEVENT\r\n")
	}

	ical.WriteString("END:VCALENDAR\r\n")
	return ical.String(), nil
}

func escapeICalText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}

func (s *Service) CreateEventFromService(ctx context.Context, tenantID, userID, serviceID, title, location string, startTime, endTime time.Time) (*Event, error) {
	event := &Event{
		Title:       title,
		Description: fmt.Sprintf("Created from service %s", serviceID),
		Location:    location,
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      false,
		Recurring:   "none",
		EventType:   "service",
		Color:       "#4A8B8C",
	}

	return s.CreateEvent(ctx, tenantID, userID, event)
}

// ListAvailableRooms returns rooms not booked during the given time range
func (s *Service) ListAvailableRooms(ctx context.Context, tenantID, startTime, endTime string) ([]map[string]interface{}, error) {
	query := `
		SELECT r.id, r.name, r.capacity, COALESCE(r.description, '')
		FROM rooms r
		WHERE r.tenant_id = $1 AND r.is_active = true
		AND r.id NOT IN (
			SELECT DISTINCT e.room_id FROM events e
			WHERE e.tenant_id = $1 AND e.room_id IS NOT NULL
			AND e.start_time < $3 AND e.end_time > $2
		)
		ORDER BY r.name`

	rows, err := s.db.Query(ctx, query, tenantID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := []map[string]interface{}{}
	for rows.Next() {
		var id, name, desc string
		var capacity *int
		if err := rows.Scan(&id, &name, &capacity, &desc); err != nil {
			return nil, err
		}
		room := map[string]interface{}{"id": id, "name": name, "description": desc}
		if capacity != nil {
			room["capacity"] = *capacity
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}
