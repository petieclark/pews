package care

import (
	"context"
	"fmt"
	"strings"
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

func (s *Service) ListFollowUps(ctx context.Context, tenantID string, status, typ, priority string) ([]FollowUp, error) {
	query := `
		SELECT f.id, f.tenant_id, f.person_id,
		       COALESCE(p.first_name || ' ' || p.last_name, '') as person_name,
		       f.assigned_to, COALESCE(u.name, '') as assigned_name,
		       f.title, f.type, f.priority, f.status,
		       CASE WHEN f.due_date IS NOT NULL THEN to_char(f.due_date, 'YYYY-MM-DD') ELSE '' END as due_date,
		       f.completed_at, f.created_at, f.updated_at
		FROM follow_ups f
		LEFT JOIN people p ON f.person_id = p.id
		LEFT JOIN users u ON f.assigned_to = u.id
		WHERE f.tenant_id = $1`

	args := []interface{}{tenantID}
	argN := 1

	if status != "" {
		argN++
		query += fmt.Sprintf(" AND f.status = $%d", argN)
		args = append(args, status)
	}
	if typ != "" {
		argN++
		query += fmt.Sprintf(" AND f.type = $%d", argN)
		args = append(args, typ)
	}
	if priority != "" {
		argN++
		query += fmt.Sprintf(" AND f.priority = $%d", argN)
		args = append(args, priority)
	}

	query += " ORDER BY CASE f.priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END, f.due_date ASC NULLS LAST, f.created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list follow-ups: %w", err)
	}
	defer rows.Close()

	var items []FollowUp
	for rows.Next() {
		var f FollowUp
		var dueStr string
		err := rows.Scan(&f.ID, &f.TenantID, &f.PersonID, &f.PersonName,
			&f.AssignedTo, &f.AssignedName, &f.Title, &f.Type, &f.Priority, &f.Status,
			&dueStr, &f.CompletedAt, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follow-up: %w", err)
		}
		if dueStr != "" {
			f.DueDate = &dueStr
		}
		items = append(items, f)
	}

	return items, rows.Err()
}

func (s *Service) GetFollowUp(ctx context.Context, tenantID, id string) (*FollowUp, error) {
	query := `
		SELECT f.id, f.tenant_id, f.person_id,
		       COALESCE(p.first_name || ' ' || p.last_name, '') as person_name,
		       f.assigned_to, COALESCE(u.name, '') as assigned_name,
		       f.title, f.type, f.priority, f.status,
		       CASE WHEN f.due_date IS NOT NULL THEN to_char(f.due_date, 'YYYY-MM-DD') ELSE '' END as due_date,
		       f.completed_at, f.created_at, f.updated_at
		FROM follow_ups f
		LEFT JOIN people p ON f.person_id = p.id
		LEFT JOIN users u ON f.assigned_to = u.id
		WHERE f.tenant_id = $1 AND f.id = $2`

	var f FollowUp
	var dueStr string
	err := s.db.QueryRow(ctx, query, tenantID, id).Scan(
		&f.ID, &f.TenantID, &f.PersonID, &f.PersonName,
		&f.AssignedTo, &f.AssignedName, &f.Title, &f.Type, &f.Priority, &f.Status,
		&dueStr, &f.CompletedAt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("follow-up not found: %w", err)
	}
	if dueStr != "" {
		f.DueDate = &dueStr
	}

	// Load notes
	notes, err := s.ListNotes(ctx, tenantID, id)
	if err == nil {
		f.Notes = notes
	}

	return &f, nil
}

func (s *Service) CreateFollowUp(ctx context.Context, tenantID, userID string, input CreateFollowUpInput) (*FollowUp, error) {
	id := uuid.New().String()

	var dueDate interface{}
	if input.DueDate != nil && *input.DueDate != "" {
		dueDate = *input.DueDate
	}

	// Validate type
	validTypes := map[string]bool{"first_time_visitor": true, "hospital_visit": true, "counseling": true, "general": true, "membership": true}
	if !validTypes[input.Type] {
		input.Type = "general"
	}

	// Validate priority
	validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
	if !validPriorities[input.Priority] {
		input.Priority = "medium"
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO follow_ups (id, tenant_id, person_id, assigned_to, title, type, priority, status, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'new', $8)`,
		id, tenantID, input.PersonID, input.AssignedTo, input.Title, input.Type, input.Priority, dueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create follow-up: %w", err)
	}

	// Add initial note if provided
	if input.Notes != "" {
		s.AddNote(ctx, tenantID, id, userID, input.Notes)
	}

	return s.GetFollowUp(ctx, tenantID, id)
}

func (s *Service) UpdateFollowUp(ctx context.Context, tenantID, id string, input UpdateFollowUpInput) (*FollowUp, error) {
	updates := []string{}
	args := []interface{}{tenantID, id}
	n := 2

	if input.Title != nil {
		n++
		updates = append(updates, fmt.Sprintf("title = $%d", n))
		args = append(args, *input.Title)
	}
	if input.Type != nil {
		n++
		updates = append(updates, fmt.Sprintf("type = $%d", n))
		args = append(args, *input.Type)
	}
	if input.Priority != nil {
		n++
		updates = append(updates, fmt.Sprintf("priority = $%d", n))
		args = append(args, *input.Priority)
	}
	if input.Status != nil {
		n++
		updates = append(updates, fmt.Sprintf("status = $%d", n))
		args = append(args, *input.Status)
		if *input.Status == "completed" {
			updates = append(updates, "completed_at = NOW()")
		} else {
			updates = append(updates, "completed_at = NULL")
		}
	}
	if input.AssignedTo != nil {
		n++
		updates = append(updates, fmt.Sprintf("assigned_to = $%d", n))
		args = append(args, *input.AssignedTo)
	}
	if input.DueDate != nil {
		n++
		if *input.DueDate == "" {
			updates = append(updates, "due_date = NULL")
		} else {
			updates = append(updates, fmt.Sprintf("due_date = $%d", n))
			args = append(args, *input.DueDate)
		}
	}

	if len(updates) == 0 {
		return s.GetFollowUp(ctx, tenantID, id)
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE follow_ups SET %s WHERE tenant_id = $1 AND id = $2", strings.Join(updates, ", "))

	_, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update follow-up: %w", err)
	}

	return s.GetFollowUp(ctx, tenantID, id)
}

func (s *Service) DeleteFollowUp(ctx context.Context, tenantID, id string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM follow_ups WHERE tenant_id = $1 AND id = $2", tenantID, id)
	return err
}

func (s *Service) AddNote(ctx context.Context, tenantID, followUpID, authorID, note string) (*Note, error) {
	id := uuid.New().String()
	_, err := s.db.Exec(ctx,
		`INSERT INTO follow_up_notes (id, follow_up_id, author_id, note) VALUES ($1, $2, $3, $4)`,
		id, followUpID, authorID, note)
	if err != nil {
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	return &Note{
		ID:         id,
		FollowUpID: followUpID,
		AuthorID:   &authorID,
		Note:       note,
		CreatedAt:  time.Now(),
	}, nil
}

func (s *Service) ListNotes(ctx context.Context, tenantID, followUpID string) ([]Note, error) {
	query := `
		SELECT n.id, n.follow_up_id, n.author_id, COALESCE(u.name, 'System') as author_name, n.note, n.created_at
		FROM follow_up_notes n
		LEFT JOIN users u ON n.author_id = u.id
		WHERE n.follow_up_id = $1
		ORDER BY n.created_at ASC`

	rows, err := s.db.Query(ctx, query, followUpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.FollowUpID, &n.AuthorID, &n.AuthorName, &n.Note, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Service) GetStats(ctx context.Context, tenantID string) (*Stats, error) {
	var stats Stats
	err := s.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'new') as new_count,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress_count,
			COUNT(*) FILTER (WHERE status = 'waiting') as waiting_count,
			COUNT(*) FILTER (WHERE status = 'completed') as completed_count,
			COUNT(*) FILTER (WHERE status != 'completed' AND due_date < CURRENT_DATE) as overdue_count,
			COUNT(*) FILTER (WHERE status != 'completed' AND due_date = CURRENT_DATE) as due_today_count,
			COUNT(*) FILTER (WHERE status != 'completed' AND due_date >= CURRENT_DATE AND due_date <= CURRENT_DATE + INTERVAL '7 days') as due_this_week_count
		FROM follow_ups WHERE tenant_id = $1
	`, tenantID).Scan(&stats.NewCount, &stats.InProgressCount, &stats.WaitingCount,
		&stats.CompletedCount, &stats.OverdueCount, &stats.DueTodayCount, &stats.DueThisWeekCount)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
