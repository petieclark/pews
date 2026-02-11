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

// ListEvents returns events within an optional date range
func (s *Service) ListEvents(ctx context.Context, tenantID, fromDate, toDate string, page, limit int) ([]Event, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	query := `SELECT id, tenant_id, title, description, location, start_time, end_time, 
	          all_day, recurring, color, created_by, created_at, updated_at
	          FROM events WHERE tenant_id = $1`
	
	args := []interface{}{tenantID}
	argCount := 1

	if fromDate != "" {
		argCount++
		query += fmt.Sprintf(" AND start_time >= $%d", argCount)
		args = append(args, fromDate)
	}

	if toDate != "" {
		argCount++
		query += fmt.Sprintf(" AND start_time <= $%d", argCount)
		args = append(args, toDate)
	}

	query += " ORDER BY start_time ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
			&e.StartTime, &e.EndTime, &e.AllDay, &e.Recurring, &e.Color,
			&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM events WHERE tenant_id = $1`
	countArgs := []interface{}{tenantID}
	if fromDate != "" {
		countQuery += " AND start_time >= $2"
		countArgs = append(countArgs, fromDate)
		if toDate != "" {
			countQuery += " AND start_time <= $3"
			countArgs = append(countArgs, toDate)
		}
	} else if toDate != "" {
		countQuery += " AND start_time <= $2"
		countArgs = append(countArgs, toDate)
	}

	err = s.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	return events, total, nil
}

// GetEventByID retrieves a single event
func (s *Service) GetEventByID(ctx context.Context, tenantID, eventID string) (*Event, error) {
	query := `SELECT id, tenant_id, title, description, location, start_time, end_time,
	          all_day, recurring, color, created_by, created_at, updated_at
	          FROM events WHERE id = $1 AND tenant_id = $2`

	var e Event
	err := s.db.QueryRow(ctx, query, eventID, tenantID).Scan(
		&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
		&e.StartTime, &e.EndTime, &e.AllDay, &e.Recurring, &e.Color,
		&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &e, nil
}

// CreateEvent creates a new event
func (s *Service) CreateEvent(ctx context.Context, tenantID, userID string, event *Event) (*Event, error) {
	query := `INSERT INTO events (tenant_id, title, description, location, start_time, end_time,
	          all_day, recurring, color, created_by)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	          RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		tenantID, event.Title, event.Description, event.Location,
		event.StartTime, event.EndTime, event.AllDay, event.Recurring,
		event.Color, userID,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	event.TenantID = tenantID
	event.CreatedBy = userID
	return event, nil
}

// UpdateEvent updates an existing event
func (s *Service) UpdateEvent(ctx context.Context, tenantID, eventID string, event *Event) (*Event, error) {
	query := `UPDATE events SET title = $1, description = $2, location = $3,
	          start_time = $4, end_time = $5, all_day = $6, recurring = $7, color = $8,
	          updated_at = NOW()
	          WHERE id = $9 AND tenant_id = $10
	          RETURNING id, tenant_id, title, description, location, start_time, end_time,
	          all_day, recurring, color, created_by, created_at, updated_at`

	var e Event
	err := s.db.QueryRow(ctx, query,
		event.Title, event.Description, event.Location,
		event.StartTime, event.EndTime, event.AllDay, event.Recurring, event.Color,
		eventID, tenantID,
	).Scan(&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
		&e.StartTime, &e.EndTime, &e.AllDay, &e.Recurring, &e.Color,
		&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return &e, nil
}

// DeleteEvent deletes an event
func (s *Service) DeleteEvent(ctx context.Context, tenantID, eventID string) error {
	query := `DELETE FROM events WHERE id = $1 AND tenant_id = $2`
	_, err := s.db.Exec(ctx, query, eventID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// GetUpcomingEvents returns the next N upcoming events
func (s *Service) GetUpcomingEvents(ctx context.Context, tenantID string, limit int) ([]Event, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	query := `SELECT id, tenant_id, title, description, location, start_time, end_time,
	          all_day, recurring, color, created_by, created_at, updated_at
	          FROM events WHERE tenant_id = $1 AND start_time >= NOW()
	          ORDER BY start_time ASC LIMIT $2`

	rows, err := s.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(&e.ID, &e.TenantID, &e.Title, &e.Description, &e.Location,
			&e.StartTime, &e.EndTime, &e.AllDay, &e.Recurring, &e.Color,
			&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

// GenerateICal generates an iCalendar format (.ics) for all tenant events
func (s *Service) GenerateICal(ctx context.Context, tenantID string) (string, error) {
	events, _, err := s.ListEvents(ctx, tenantID, "", "", 1, 1000)
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

		// Handle recurring events
		if event.Recurring == "weekly" {
			ical.WriteString("RRULE:FREQ=WEEKLY\r\n")
		} else if event.Recurring == "monthly" {
			ical.WriteString("RRULE:FREQ=MONTHLY\r\n")
		}

		ical.WriteString("END:VEVENT\r\n")
	}

	ical.WriteString("END:VCALENDAR\r\n")
	return ical.String(), nil
}

// escapeICalText escapes special characters in iCal text fields
func escapeICalText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}

// CreateEventFromService creates a calendar event from a service
func (s *Service) CreateEventFromService(ctx context.Context, tenantID, userID, serviceID, title, location string, startTime, endTime time.Time) (*Event, error) {
	event := &Event{
		Title:       title,
		Description: fmt.Sprintf("Created from service %s", serviceID),
		Location:    location,
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      false,
		Recurring:   "none",
		Color:       "#4A8B8C",
	}

	return s.CreateEvent(ctx, tenantID, userID, event)
}
