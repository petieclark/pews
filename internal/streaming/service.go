package streaming

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

// Stream operations

func (s *Service) ListStreams(ctx context.Context, tenantID string, status string, page, limit int) ([]Stream, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	sqlQuery := `
		SELECT id, tenant_id, title, description, service_id, status, 
		       scheduled_start, actual_start, actual_end, stream_type, 
		       stream_url, stream_key, embed_url, chat_enabled, giving_enabled, 
		       connection_card_enabled, viewer_count, peak_viewers, 
		       created_at, updated_at
		FROM streams
		WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM streams WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if status != "" {
		statusFilter := fmt.Sprintf(` AND status = $%d`, argPos)
		sqlQuery += statusFilter
		countQuery += statusFilter
		args = append(args, status)
		argPos++
	}

	// Get total count
	var total int
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count streams: %w", err)
	}

	// Add pagination
	sqlQuery += fmt.Sprintf(` ORDER BY scheduled_start DESC LIMIT $%d OFFSET $%d`, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list streams: %w", err)
	}
	defer rows.Close()

	streams := []Stream{}
	for rows.Next() {
		var st Stream
		err := rows.Scan(
			&st.ID, &st.TenantID, &st.Title, &st.Description, &st.ServiceID, &st.Status,
			&st.ScheduledStart, &st.ActualStart, &st.ActualEnd, &st.StreamType,
			&st.StreamURL, &st.StreamKey, &st.EmbedURL, &st.ChatEnabled, &st.GivingEnabled,
			&st.ConnectionCardEnabled, &st.ViewerCount, &st.PeakViewers,
			&st.CreatedAt, &st.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stream: %w", err)
		}
		streams = append(streams, st)
	}

	return streams, total, nil
}

func (s *Service) GetStreamByID(ctx context.Context, tenantID, streamID string) (*Stream, error) {
	var st Stream
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, description, service_id, status, 
		       scheduled_start, actual_start, actual_end, stream_type, 
		       stream_url, stream_key, embed_url, chat_enabled, giving_enabled, 
		       connection_card_enabled, viewer_count, peak_viewers, 
		       created_at, updated_at
		FROM streams
		WHERE id = $1
	`, streamID).Scan(
		&st.ID, &st.TenantID, &st.Title, &st.Description, &st.ServiceID, &st.Status,
		&st.ScheduledStart, &st.ActualStart, &st.ActualEnd, &st.StreamType,
		&st.StreamURL, &st.StreamKey, &st.EmbedURL, &st.ChatEnabled, &st.GivingEnabled,
		&st.ConnectionCardEnabled, &st.ViewerCount, &st.PeakViewers,
		&st.CreatedAt, &st.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("stream not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	return &st, nil
}

func (s *Service) GetStreamByIDPublic(ctx context.Context, streamID string) (*Stream, error) {
	// Public access - no tenant context needed, but we still fetch it
	var st Stream
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, description, service_id, status, 
		       scheduled_start, actual_start, actual_end, stream_type, 
		       stream_url, stream_key, embed_url, chat_enabled, giving_enabled, 
		       connection_card_enabled, viewer_count, peak_viewers, 
		       created_at, updated_at
		FROM streams
		WHERE id = $1
	`, streamID).Scan(
		&st.ID, &st.TenantID, &st.Title, &st.Description, &st.ServiceID, &st.Status,
		&st.ScheduledStart, &st.ActualStart, &st.ActualEnd, &st.StreamType,
		&st.StreamURL, &st.StreamKey, &st.EmbedURL, &st.ChatEnabled, &st.GivingEnabled,
		&st.ConnectionCardEnabled, &st.ViewerCount, &st.PeakViewers,
		&st.CreatedAt, &st.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("stream not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	return &st, nil
}

func (s *Service) CreateStream(ctx context.Context, tenantID string, stream *Stream) error {
	stream.ID = uuid.New().String()
	stream.TenantID = tenantID

	_, err := s.db.Exec(ctx, `
		INSERT INTO streams (
			id, tenant_id, title, description, service_id, status, 
			scheduled_start, stream_type, stream_url, stream_key, embed_url,
			chat_enabled, giving_enabled, connection_card_enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, stream.ID, stream.TenantID, stream.Title, stream.Description, stream.ServiceID,
		stream.Status, stream.ScheduledStart, stream.StreamType, stream.StreamURL,
		stream.StreamKey, stream.EmbedURL, stream.ChatEnabled, stream.GivingEnabled,
		stream.ConnectionCardEnabled)

	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	return nil
}

func (s *Service) UpdateStream(ctx context.Context, tenantID, streamID string, stream *Stream) error {
	_, err := s.db.Exec(ctx, `
		UPDATE streams SET
			title = $1, description = $2, service_id = $3, status = $4,
			scheduled_start = $5, stream_type = $6, stream_url = $7, 
			stream_key = $8, embed_url = $9, chat_enabled = $10, 
			giving_enabled = $11, connection_card_enabled = $12
		WHERE id = $13
	`, stream.Title, stream.Description, stream.ServiceID, stream.Status,
		stream.ScheduledStart, stream.StreamType, stream.StreamURL,
		stream.StreamKey, stream.EmbedURL, stream.ChatEnabled,
		stream.GivingEnabled, stream.ConnectionCardEnabled, streamID)

	if err != nil {
		return fmt.Errorf("failed to update stream: %w", err)
	}

	return nil
}

func (s *Service) DeleteStream(ctx context.Context, tenantID, streamID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM streams WHERE id = $1", streamID)
	if err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	return nil
}

func (s *Service) GoLive(ctx context.Context, tenantID, streamID string) error {
	now := time.Now()
	_, err := s.db.Exec(ctx, `
		UPDATE streams 
		SET status = 'live', actual_start = $1
		WHERE id = $2
	`, now, streamID)

	if err != nil {
		return fmt.Errorf("failed to mark stream as live: %w", err)
	}

	return nil
}

func (s *Service) EndStream(ctx context.Context, tenantID, streamID string) error {
	now := time.Now()
	_, err := s.db.Exec(ctx, `
		UPDATE streams 
		SET status = 'ended', actual_end = $1
		WHERE id = $2
	`, now, streamID)

	if err != nil {
		return fmt.Errorf("failed to end stream: %w", err)
	}

	return nil
}

func (s *Service) GetLiveStream(ctx context.Context, tenantID string) (*Stream, error) {
	var st Stream
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, description, service_id, status, 
		       scheduled_start, actual_start, actual_end, stream_type, 
		       stream_url, stream_key, embed_url, chat_enabled, giving_enabled, 
		       connection_card_enabled, viewer_count, peak_viewers, 
		       created_at, updated_at
		FROM streams
		WHERE status = 'live'
		ORDER BY actual_start DESC
		LIMIT 1
	`).Scan(
		&st.ID, &st.TenantID, &st.Title, &st.Description, &st.ServiceID, &st.Status,
		&st.ScheduledStart, &st.ActualStart, &st.ActualEnd, &st.StreamType,
		&st.StreamURL, &st.StreamKey, &st.EmbedURL, &st.ChatEnabled, &st.GivingEnabled,
		&st.ConnectionCardEnabled, &st.ViewerCount, &st.PeakViewers,
		&st.CreatedAt, &st.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // No live stream
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get live stream: %w", err)
	}

	return &st, nil
}

// Chat operations

func (s *Service) GetChatMessages(ctx context.Context, streamID string, after string, limit int) ([]ChatMessage, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}

	sqlQuery := `
		SELECT id, stream_id, person_id, guest_name, message, 
		       is_pinned, is_deleted, created_at
		FROM stream_chat
		WHERE stream_id = $1 AND is_deleted = false`

	args := []interface{}{streamID}
	argPos := 2

	if after != "" {
		sqlQuery += fmt.Sprintf(` AND created_at > (SELECT created_at FROM stream_chat WHERE id = $%d)`, argPos)
		args = append(args, after)
		argPos++
	}

	sqlQuery += fmt.Sprintf(` ORDER BY created_at ASC LIMIT $%d`, argPos)
	args = append(args, limit)

	rows, err := s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var msg ChatMessage
		err := rows.Scan(
			&msg.ID, &msg.StreamID, &msg.PersonID, &msg.GuestName,
			&msg.Message, &msg.IsPinned, &msg.IsDeleted, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *Service) SendChatMessage(ctx context.Context, streamID string, personID *string, guestName, message string) (*ChatMessage, error) {
	msg := &ChatMessage{
		ID:        uuid.New().String(),
		StreamID:  streamID,
		PersonID:  personID,
		GuestName: guestName,
		Message:   message,
		CreatedAt: time.Now(),
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO stream_chat (id, stream_id, person_id, guest_name, message)
		VALUES ($1, $2, $3, $4, $5)
	`, msg.ID, msg.StreamID, msg.PersonID, msg.GuestName, msg.Message)

	if err != nil {
		return nil, fmt.Errorf("failed to send chat message: %w", err)
	}

	return msg, nil
}

func (s *Service) PinChatMessage(ctx context.Context, tenantID, messageID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE stream_chat SET is_pinned = true WHERE id = $1
	`, messageID)

	if err != nil {
		return fmt.Errorf("failed to pin message: %w", err)
	}

	return nil
}

func (s *Service) DeleteChatMessage(ctx context.Context, tenantID, messageID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE stream_chat SET is_deleted = true WHERE id = $1
	`, messageID)

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// Viewer operations

func (s *Service) JoinStream(ctx context.Context, streamID string, personID *string, guestName string) (*StreamViewer, error) {
	viewer := &StreamViewer{
		ID:        uuid.New().String(),
		StreamID:  streamID,
		PersonID:  personID,
		GuestName: guestName,
		JoinedAt:  time.Now(),
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO stream_viewers (id, stream_id, person_id, guest_name, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`, viewer.ID, viewer.StreamID, viewer.PersonID, viewer.GuestName, viewer.JoinedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to join stream: %w", err)
	}

	// Update viewer count
	_, _ = s.db.Exec(ctx, `
		UPDATE streams 
		SET viewer_count = (
			SELECT COUNT(*) FROM stream_viewers 
			WHERE stream_id = $1 AND left_at IS NULL
		),
		peak_viewers = GREATEST(peak_viewers, (
			SELECT COUNT(*) FROM stream_viewers 
			WHERE stream_id = $1 AND left_at IS NULL
		))
		WHERE id = $1
	`, streamID)

	return viewer, nil
}

func (s *Service) LeaveStream(ctx context.Context, streamID, viewerID string) error {
	now := time.Now()
	_, err := s.db.Exec(ctx, `
		UPDATE stream_viewers 
		SET left_at = $1,
		    duration_seconds = EXTRACT(EPOCH FROM ($1 - joined_at))::INTEGER
		WHERE id = $2
	`, now, viewerID)

	if err != nil {
		return fmt.Errorf("failed to leave stream: %w", err)
	}

	// Update viewer count
	_, _ = s.db.Exec(ctx, `
		UPDATE streams 
		SET viewer_count = (
			SELECT COUNT(*) FROM stream_viewers 
			WHERE stream_id = $1 AND left_at IS NULL
		)
		WHERE id = $1
	`, streamID)

	return nil
}

func (s *Service) GetViewers(ctx context.Context, tenantID, streamID string) ([]StreamViewer, int, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, stream_id, person_id, guest_name, joined_at, left_at, duration_seconds
		FROM stream_viewers
		WHERE stream_id = $1 AND left_at IS NULL
		ORDER BY joined_at DESC
	`, streamID)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get viewers: %w", err)
	}
	defer rows.Close()

	viewers := []StreamViewer{}
	for rows.Next() {
		var v StreamViewer
		err := rows.Scan(
			&v.ID, &v.StreamID, &v.PersonID, &v.GuestName,
			&v.JoinedAt, &v.LeftAt, &v.DurationSeconds,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan viewer: %w", err)
		}
		viewers = append(viewers, v)
	}

	return viewers, len(viewers), nil
}

// Notes operations

func (s *Service) GetStreamNotes(ctx context.Context, tenantID, streamID, personID string) (*StreamNote, error) {
	var note StreamNote
	err := s.db.QueryRow(ctx, `
		SELECT id, stream_id, person_id, content, created_at, updated_at
		FROM stream_notes
		WHERE stream_id = $1 AND person_id = $2
	`, streamID, personID).Scan(
		&note.ID, &note.StreamID, &note.PersonID, &note.Content,
		&note.CreatedAt, &note.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // No notes yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	return &note, nil
}

func (s *Service) SaveStreamNotes(ctx context.Context, tenantID, streamID, personID, content string) (*StreamNote, error) {
	// Check if notes already exist
	existing, err := s.GetStreamNotes(ctx, tenantID, streamID, personID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update existing notes
		_, err = s.db.Exec(ctx, `
			UPDATE stream_notes SET content = $1 WHERE id = $2
		`, content, existing.ID)

		if err != nil {
			return nil, fmt.Errorf("failed to update notes: %w", err)
		}

		existing.Content = content
		return existing, nil
	}

	// Create new notes
	note := &StreamNote{
		ID:       uuid.New().String(),
		StreamID: streamID,
		PersonID: &personID,
		Content:  content,
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO stream_notes (id, stream_id, person_id, content)
		VALUES ($1, $2, $3, $4)
	`, note.ID, note.StreamID, note.PersonID, note.Content)

	if err != nil {
		return nil, fmt.Errorf("failed to create notes: %w", err)
	}

	return note, nil
}
