package rooms

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (s *Service) ListRooms(ctx context.Context, tenantID string, activeOnly bool) ([]Room, error) {
	query := `SELECT id, tenant_id, name, capacity, COALESCE(description, ''), COALESCE(color, '#4A8B8C'), amenities, is_active, created_at, updated_at 
		      FROM rooms WHERE tenant_id = $1`
	if activeOnly {
		query += " AND is_active = TRUE"
	}
	query += " ORDER BY name ASC"

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var r Room
		var amenitiesJSON []byte
		err := rows.Scan(&r.ID, &r.TenantID, &r.Name, &r.Capacity, &r.Description, 
			&r.Color, &amenitiesJSON, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if amenitiesJSON != nil {
			json.Unmarshal(amenitiesJSON, &r.Amenities)
		} else {
			r.Amenities = []string{}
		}
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

func (s *Service) GetRoom(ctx context.Context, tenantID, roomID string) (*Room, error) {
	var r Room
	var amenitiesJSON []byte
	err := s.db.QueryRow(ctx,
		`SELECT id, tenant_id, name, capacity, COALESCE(description, ''), COALESCE(color, '#4A8B8C'), amenities, is_active, created_at, updated_at 
		 FROM rooms WHERE id = $1 AND tenant_id = $2`,
		roomID, tenantID,
	).Scan(&r.ID, &r.TenantID, &r.Name, &r.Capacity, &r.Description, 
		&r.Color, &amenitiesJSON, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if amenitiesJSON != nil {
		json.Unmarshal(amenitiesJSON, &r.Amenities)
	}
	return &r, nil
}

func (s *Service) CreateRoom(ctx context.Context, tenantID, name, description, color string, capacity *int, amenities []string) (*Room, error) {
	room := &Room{
		ID: uuid.New().String(), TenantID: tenantID, Name: name, Capacity: capacity,
		Description: description, Color: color, Amenities: amenities, IsActive: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	amenitiesJSON, _ := json.Marshal(amenities)
	_, err := s.db.Exec(ctx,
		`INSERT INTO rooms (id, tenant_id, name, capacity, description, color, amenities, is_active) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		room.ID, room.TenantID, room.Name, room.Capacity, room.Description, room.Color, amenitiesJSON, room.IsActive,
	)
	return room, err
}

func (s *Service) UpdateRoom(ctx context.Context, tenantID, roomID, name, description, color string, capacity *int, amenities []string, isActive bool) (*Room, error) {
	amenitiesJSON, _ := json.Marshal(amenities)
	_, err := s.db.Exec(ctx,
		`UPDATE rooms SET name = $1, capacity = $2, description = $3, color = $4, amenities = $5, is_active = $6, updated_at = NOW()
		 WHERE id = $7 AND tenant_id = $8`,
		name, capacity, description, color, amenitiesJSON, isActive, roomID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetRoom(ctx, tenantID, roomID)
}

func (s *Service) DeleteRoom(ctx context.Context, tenantID, roomID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM rooms WHERE id = $1 AND tenant_id = $2`, roomID, tenantID)
	return err
}

func (s *Service) ListBookings(ctx context.Context, tenantID string, roomID *string, startTime, endTime *time.Time) ([]RoomBooking, error) {
	query := `SELECT rb.id, rb.tenant_id, rb.room_id, r.name, rb.event_name, 
	                 rb.booked_by, u.name, rb.start_time, rb.end_time, rb.recurring, rb.status, rb.notes, rb.created_at, rb.updated_at
	          FROM room_bookings rb LEFT JOIN rooms r ON rb.room_id = r.id LEFT JOIN users u ON rb.booked_by = u.id
	          WHERE rb.tenant_id = $1`
	args := []interface{}{tenantID}
	argCount := 1
	if roomID != nil {
		argCount++
		query += fmt.Sprintf(" AND rb.room_id = $%d", argCount)
		args = append(args, *roomID)
	}
	if startTime != nil {
		argCount++
		query += fmt.Sprintf(" AND rb.end_time > $%d", argCount)
		args = append(args, *startTime)
	}
	if endTime != nil {
		argCount++
		query += fmt.Sprintf(" AND rb.start_time < $%d", argCount)
		args = append(args, *endTime)
	}
	query += " ORDER BY rb.start_time ASC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []RoomBooking
	for rows.Next() {
		var b RoomBooking
		var roomName, bookedByName sql.NullString
		err := rows.Scan(&b.ID, &b.TenantID, &b.RoomID, &roomName, &b.EventName,
			&b.BookedBy, &bookedByName, &b.StartTime, &b.EndTime,
			&b.Recurring, &b.Status, &b.Notes, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if roomName.Valid {
			b.RoomName = roomName.String
		}
		if bookedByName.Valid {
			b.BookedByName = &bookedByName.String
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

func (s *Service) GetBooking(ctx context.Context, tenantID, bookingID string) (*RoomBooking, error) {
	var b RoomBooking
	var roomName, bookedByName sql.NullString
	err := s.db.QueryRow(ctx,
		`SELECT rb.id, rb.tenant_id, rb.room_id, r.name, rb.event_name, rb.booked_by, u.name, 
		        rb.start_time, rb.end_time, rb.recurring, rb.status, rb.notes, rb.created_at, rb.updated_at
		 FROM room_bookings rb LEFT JOIN rooms r ON rb.room_id = r.id LEFT JOIN users u ON rb.booked_by = u.id
		 WHERE rb.id = $1 AND rb.tenant_id = $2`,
		bookingID, tenantID,
	).Scan(&b.ID, &b.TenantID, &b.RoomID, &roomName, &b.EventName,
		&b.BookedBy, &bookedByName, &b.StartTime, &b.EndTime,
		&b.Recurring, &b.Status, &b.Notes, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if roomName.Valid {
		b.RoomName = roomName.String
	}
	if bookedByName.Valid {
		b.BookedByName = &bookedByName.String
	}
	return &b, nil
}

func (s *Service) CheckConflict(ctx context.Context, tenantID, roomID string, startTime, endTime time.Time, excludeBookingID *string) (*BookingConflict, error) {
	query := `SELECT id, tenant_id, room_id, event_name, booked_by, start_time, end_time, recurring, status, notes, created_at, updated_at
	          FROM room_bookings WHERE tenant_id = $1 AND room_id = $2 AND status != 'cancelled'
	            AND ((start_time <= $3 AND end_time > $3) OR (start_time < $4 AND end_time >= $4) OR (start_time >= $3 AND end_time <= $4))`
	args := []interface{}{tenantID, roomID, startTime, endTime}
	if excludeBookingID != nil {
		query += " AND id != $5"
		args = append(args, *excludeBookingID)
	}
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var conflicts []RoomBooking
	for rows.Next() {
		var b RoomBooking
		rows.Scan(&b.ID, &b.TenantID, &b.RoomID, &b.EventName, &b.BookedBy, &b.StartTime, &b.EndTime,
			&b.Recurring, &b.Status, &b.Notes, &b.CreatedAt, &b.UpdatedAt)
		conflicts = append(conflicts, b)
	}
	return &BookingConflict{HasConflict: len(conflicts) > 0, Conflicts: conflicts}, rows.Err()
}

func (s *Service) CreateBooking(ctx context.Context, tenantID, roomID, eventName string, bookedBy *string, startTime, endTime time.Time, recurring *string, status string, notes *string) (*RoomBooking, error) {
	conflict, err := s.CheckConflict(ctx, tenantID, roomID, startTime, endTime, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflicts: %w", err)
	}
	if conflict.HasConflict {
		return nil, fmt.Errorf("booking conflicts with existing bookings")
	}
	booking := &RoomBooking{
		ID: uuid.New().String(), TenantID: tenantID, RoomID: roomID, EventName: eventName,
		BookedBy: bookedBy, StartTime: startTime, EndTime: endTime, Recurring: recurring,
		Status: status, Notes: notes, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_, err = s.db.Exec(ctx,
		`INSERT INTO room_bookings (id, tenant_id, room_id, event_name, booked_by, start_time, end_time, recurring, status, notes) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		booking.ID, booking.TenantID, booking.RoomID, booking.EventName, booking.BookedBy,
		booking.StartTime, booking.EndTime, booking.Recurring, booking.Status, booking.Notes,
	)
	if err != nil {
		return nil, err
	}
	return s.GetBooking(ctx, tenantID, booking.ID)
}

func (s *Service) UpdateBooking(ctx context.Context, tenantID, bookingID, eventName string, startTime, endTime time.Time, recurring *string, status string, notes *string) (*RoomBooking, error) {
	existing, err := s.GetBooking(ctx, tenantID, bookingID)
	if err != nil {
		return nil, err
	}
	conflict, err := s.CheckConflict(ctx, tenantID, existing.RoomID, startTime, endTime, &bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflicts: %w", err)
	}
	if conflict.HasConflict {
		return nil, fmt.Errorf("booking conflicts with existing bookings")
	}
	_, err = s.db.Exec(ctx,
		`UPDATE room_bookings SET event_name = $1, start_time = $2, end_time = $3, recurring = $4, status = $5, notes = $6, updated_at = NOW()
		 WHERE id = $7 AND tenant_id = $8`,
		eventName, startTime, endTime, recurring, status, notes, bookingID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetBooking(ctx, tenantID, bookingID)
}

func (s *Service) DeleteBooking(ctx context.Context, tenantID, bookingID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM room_bookings WHERE id = $1 AND tenant_id = $2`, bookingID, tenantID)
	return err
}

func (s *Service) GetRoomAvailability(ctx context.Context, tenantID string, startTime, endTime time.Time) ([]RoomAvailability, error) {
	rows, err := s.db.Query(ctx,
		`SELECT r.id, r.name, r.capacity, r.color,
		        NOT EXISTS (SELECT 1 FROM room_bookings rb WHERE rb.room_id = r.id AND rb.tenant_id = r.tenant_id AND rb.status != 'cancelled'
		            AND ((rb.start_time <= $2 AND rb.end_time > $2) OR (rb.start_time < $3 AND rb.end_time >= $3) OR (rb.start_time >= $2 AND rb.end_time <= $3))
		        ) as available
		 FROM rooms r WHERE r.tenant_id = $1 AND r.is_active = TRUE ORDER BY r.name ASC`,
		tenantID, startTime, endTime,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var availability []RoomAvailability
	for rows.Next() {
		var a RoomAvailability
		rows.Scan(&a.RoomID, &a.RoomName, &a.Capacity, &a.Color, &a.Available)
		availability = append(availability, a)
	}
	return availability, rows.Err()
}
