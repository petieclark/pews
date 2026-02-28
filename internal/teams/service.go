package teams

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db                  *pgxpool.Pool
	notificationService *NotificationService
}

func NewService(db *pgxpool.Pool, jwtSecret string) *Service {
	s := &Service{db: db}
	
	if jwtSecret != "" && len(jwtSecret) > 10 {
		s.notificationService = NewNotificationService(db, jwtSecret)
	}
	
	return s
}

// GetNotificationService returns the notification service (for testing or external use)
func (s *Service) GetNotificationService() *NotificationService {
	return s.notificationService
}

func (s *Service) ListTeams(ctx context.Context, tenantID string) ([]Team, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.tenant_id, t.name, COALESCE(t.description, ''), t.color, t.is_active,
		       t.created_at, t.updated_at,
		       COUNT(DISTINCT tm.id) as member_count,
		       COUNT(DISTINCT tp.id) as position_count
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN team_positions tp ON tp.team_id = t.id
		WHERE t.tenant_id = $1
		GROUP BY t.id
		ORDER BY t.name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []Team{}
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Description, &t.Color, &t.IsActive,
			&t.CreatedAt, &t.UpdatedAt, &t.MemberCount, &t.PositionCount); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (s *Service) CreateTeam(ctx context.Context, tenantID, name, description, color string) (*Team, error) {
	if color == "" {
		color = "#4A8B8C"
	}
	var t Team
	err := s.db.QueryRow(ctx, `
		INSERT INTO teams (tenant_id, name, description, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, name, COALESCE(description, ''), color, is_active, created_at, updated_at`,
		tenantID, name, description, color).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Description, &t.Color, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) GetTeam(ctx context.Context, tenantID, teamID string) (*Team, error) {
	var t Team
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), color, is_active, created_at, updated_at
		FROM teams WHERE id = $1 AND tenant_id = $2`, teamID, tenantID).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Description, &t.Color, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	// Load positions
	posRows, err := s.db.Query(ctx, `
		SELECT id, team_id, name, sort_order, created_at
		FROM team_positions WHERE team_id = $1 ORDER BY sort_order, name`, teamID)
	if err != nil {
		return nil, err
	}
	defer posRows.Close()
	for posRows.Next() {
		var p Position
		if err := posRows.Scan(&p.ID, &p.TeamID, &p.Name, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, err
		}
		t.Positions = append(t.Positions, p)
	}

	// Load members with person names
	memRows, err := s.db.Query(ctx, `
		SELECT tm.id, tm.team_id, tm.person_id, tm.position_id, tm.status, tm.joined_at,
		       p.first_name, p.last_name, COALESCE(p.email, ''),
		       tp.name as position_name
		FROM team_members tm
		JOIN people p ON p.id = tm.person_id
		LEFT JOIN team_positions tp ON tp.id = tm.position_id
		WHERE tm.team_id = $1
		ORDER BY p.last_name, p.first_name`, teamID)
	if err != nil {
		return nil, err
	}
	defer memRows.Close()
	for memRows.Next() {
		var m Member
		if err := memRows.Scan(&m.ID, &m.TeamID, &m.PersonID, &m.PositionID, &m.Status, &m.JoinedAt,
			&m.FirstName, &m.LastName, &m.Email, &m.PositionName); err != nil {
			return nil, err
		}
		t.Members = append(t.Members, m)
	}

	if t.Positions == nil {
		t.Positions = []Position{}
	}
	if t.Members == nil {
		t.Members = []Member{}
	}
	t.MemberCount = len(t.Members)
	return &t, nil
}

func (s *Service) UpdateTeam(ctx context.Context, tenantID, teamID, name, description, color string, isActive bool) (*Team, error) {
	var t Team
	err := s.db.QueryRow(ctx, `
		UPDATE teams SET name = $3, description = $4, color = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, name, COALESCE(description, ''), color, is_active, created_at, updated_at`,
		teamID, tenantID, name, description, color, isActive).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Description, &t.Color, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}
	return &t, nil
}

func (s *Service) DeleteTeam(ctx context.Context, tenantID, teamID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM teams WHERE id = $1 AND tenant_id = $2`, teamID, tenantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}

func (s *Service) AddPosition(ctx context.Context, teamID, name string, sortOrder int) (*Position, error) {
	var p Position
	err := s.db.QueryRow(ctx, `
		INSERT INTO team_positions (team_id, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, team_id, name, sort_order, created_at`,
		teamID, name, sortOrder).Scan(&p.ID, &p.TeamID, &p.Name, &p.SortOrder, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) DeletePosition(ctx context.Context, teamID, positionID string) error {
	// Nullify position_id on members first
	_, _ = s.db.Exec(ctx, `UPDATE team_members SET position_id = NULL WHERE position_id = $1`, positionID)
	ct, err := s.db.Exec(ctx, `DELETE FROM team_positions WHERE id = $1 AND team_id = $2`, positionID, teamID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("position not found")
	}
	return nil
}

func (s *Service) AddMember(ctx context.Context, teamID, personID string, positionID *string) (*Member, error) {
	var m Member
	err := s.db.QueryRow(ctx, `
		INSERT INTO team_members (team_id, person_id, position_id)
		VALUES ($1, $2, $3)
		RETURNING id, team_id, person_id, position_id, status, joined_at`,
		teamID, personID, positionID).Scan(&m.ID, &m.TeamID, &m.PersonID, &m.PositionID, &m.Status, &m.JoinedAt)
	if err != nil {
		return nil, err
	}
	// Fetch person name
	_ = s.db.QueryRow(ctx, `SELECT first_name, last_name, COALESCE(email, '') FROM people WHERE id = $1`, personID).
		Scan(&m.FirstName, &m.LastName, &m.Email)

	// Auto-tag person with team name
	var teamName, tenantID string
	_ = s.db.QueryRow(ctx, "SELECT name, tenant_id FROM volunteer_teams WHERE id = $1", teamID).Scan(&teamName, &tenantID)
	if teamName != "" && tenantID != "" {
		s.autoTagPerson(ctx, tenantID, personID, teamName)
	}

	return &m, nil
}

func (s *Service) autoTagPerson(ctx context.Context, tenantID, personID, tagName string) {
	var tagID string
	_ = s.db.QueryRow(ctx, `
		INSERT INTO tags (id, tenant_id, name, color)
		VALUES (gen_random_uuid(), $1, $2, '#4A8B8C')
		ON CONFLICT (tenant_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, tenantID, tagName).Scan(&tagID)
	if tagID != "" {
		_, _ = s.db.Exec(ctx, `INSERT INTO person_tags (person_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, personID, tagID)
	}
}

func (s *Service) UpdateMember(ctx context.Context, memberID string, positionID *string) error {
	_, err := s.db.Exec(ctx, `UPDATE team_members SET position_id = $2 WHERE id = $1`, memberID, positionID)
	return err
}

func (s *Service) DeleteMember(ctx context.Context, teamID, memberID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM team_members WHERE id = $1 AND team_id = $2`, memberID, teamID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// UpdateMemberStatus updates member status (active/inactive/on-break)
func (s *Service) UpdateMemberStatus(ctx context.Context, memberID, status string) error {
	_, err := s.db.Exec(ctx, `UPDATE team_members SET status = $2 WHERE id = $1`, memberID, status)
	return err
}

// UpdatePosition updates a position name and sort_order
func (s *Service) UpdatePosition(ctx context.Context, teamID, positionID, name string, sortOrder int) (*Position, error) {
	var p Position
	err := s.db.QueryRow(ctx, `
		UPDATE team_positions SET name = $3, sort_order = $4
		WHERE id = $1 AND team_id = $2
		RETURNING id, team_id, name, sort_order, created_at`,
		positionID, teamID, name, sortOrder).Scan(&p.ID, &p.TeamID, &p.Name, &p.SortOrder, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ---- Service Team Assignments ----

func (s *Service) GetServiceAssignments(ctx context.Context, tenantID, serviceID string) ([]ServiceTeamAssignment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT sta.id, sta.tenant_id, sta.service_id, sta.team_id, sta.position_id, sta.person_id,
		       COALESCE(sta.status, 'pending'), COALESCE(sta.notes, ''),
		       COALESCE(p.first_name, ''), COALESCE(p.last_name, ''), COALESCE(p.email, ''),
		       tp.name as position_name, t.name as team_name, COALESCE(t.color, '#4A8B8C') as team_color
		FROM service_team_assignments sta
		JOIN people p ON p.id = sta.person_id
		JOIN teams t ON t.id = sta.team_id
		LEFT JOIN team_positions tp ON tp.id = sta.position_id
		WHERE sta.service_id = $1 AND sta.tenant_id = $2
		ORDER BY t.name, tp.sort_order, p.last_name`, serviceID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []ServiceTeamAssignment{}
	for rows.Next() {
		var a ServiceTeamAssignment
		if err := rows.Scan(&a.ID, &a.TenantID, &a.ServiceID, &a.TeamID, &a.PositionID, &a.PersonID,
			&a.Status, &a.Notes,
			&a.PersonFirstName, &a.PersonLastName, &a.PersonEmail,
			&a.PositionName, &a.TeamName, &a.TeamColor); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (s *Service) SaveServiceAssignments(ctx context.Context, tenantID, serviceID string, assignments []ServiceTeamAssignment) ([]ServiceTeamAssignment, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments for this service
	_, err = tx.Exec(ctx, `DELETE FROM service_team_assignments WHERE service_id = $1 AND tenant_id = $2`, serviceID, tenantID)
	if err != nil {
		return nil, err
	}

	var createdAssignments []ServiceTeamAssignment
	
	for _, a := range assignments {
		var created ServiceTeamAssignment
		err = tx.QueryRow(ctx, `
			INSERT INTO service_team_assignments (tenant_id, service_id, team_id, position_id, person_id, status, notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, tenant_id, service_id, team_id, position_id, person_id, status, notes`,
			tenantID, serviceID, a.TeamID, a.PositionID, a.PersonID, a.Status, a.Notes).Scan(
			&created.ID, &created.TenantID, &created.ServiceID, &created.TeamID, 
			&created.PositionID, &created.PersonID, &created.Status, &created.Notes)
		
		if err != nil {
			return nil, fmt.Errorf("failed to insert assignment: %w", err)
		}
		createdAssignments = append(createdAssignments, created)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Send notifications for new assignments (fire and forget via goroutine)
	if s.notificationService != nil && len(createdAssignments) > 0 {
		go func(assignments []ServiceTeamAssignment) {
			for _, a := range assignments {
				notifErr := s.notificationService.SendAssignmentNotification(ctx, a.ID, serviceID)
				if notifErr != nil {
					// Log error (in production use proper logger)
					fmt.Printf("[NOTIFICATION] Failed to send assignment notification for ID %s: %v\n", a.ID, notifErr)
				} else {
					// Mark as notified
					s.notificationService.UpdateNotificationSent(ctx, a.ID)
				}
			}
		}(createdAssignments)
	}

	return createdAssignments, nil
}

func (s *Service) CopyServiceAssignments(ctx context.Context, tenantID, targetServiceID, sourceServiceID string) ([]ServiceTeamAssignment, error) {
	// Copy assignments from source to target
	_, err := s.db.Exec(ctx, `
		INSERT INTO service_team_assignments (tenant_id, service_id, team_id, position_id, person_id, status, notes)
		SELECT tenant_id, $3, team_id, position_id, person_id, 'pending', notes
		FROM service_team_assignments
		WHERE service_id = $1 AND tenant_id = $2
		ON CONFLICT (service_id, position_id, person_id) DO NOTHING`,
		sourceServiceID, tenantID, targetServiceID)
	if err != nil {
		return nil, err
	}

	return s.GetServiceAssignments(ctx, tenantID, targetServiceID)
}

func (s *Service) UpdateAssignmentStatus(ctx context.Context, tenantID, assignmentID, status string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE service_team_assignments SET status = $3, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`, assignmentID, tenantID, status)
	return err
}

func (s *Service) GetPersonSchedule(ctx context.Context, tenantID, personID string) ([]ServiceTeamAssignment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT sta.id, sta.tenant_id, sta.service_id, sta.team_id, sta.position_id, sta.person_id,
		       COALESCE(sta.status, 'pending'), COALESCE(sta.notes, ''),
		       COALESCE(p.first_name, ''), COALESCE(p.last_name, ''), COALESCE(p.email, ''),
		       tp.name as position_name, t.name as team_name, COALESCE(t.color, '#4A8B8C') as team_color,
		       s.service_date::text, COALESCE(s.service_time, ''),
		       COALESCE(st.name, '')
		FROM service_team_assignments sta
		JOIN people p ON p.id = sta.person_id
		JOIN teams t ON t.id = sta.team_id
		JOIN services s ON s.id = sta.service_id
		LEFT JOIN service_types st ON st.id = s.service_type_id
		LEFT JOIN team_positions tp ON tp.id = sta.position_id
		WHERE sta.person_id = $1 AND sta.tenant_id = $2
		  AND s.service_date >= CURRENT_DATE
		ORDER BY s.service_date, s.service_time`, personID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []ServiceTeamAssignment{}
	for rows.Next() {
		var a ServiceTeamAssignment
		if err := rows.Scan(&a.ID, &a.TenantID, &a.ServiceID, &a.TeamID, &a.PositionID, &a.PersonID,
			&a.Status, &a.Notes,
			&a.PersonFirstName, &a.PersonLastName, &a.PersonEmail,
			&a.PositionName, &a.TeamName, &a.TeamColor,
			&a.ServiceDate, &a.ServiceTime, &a.ServiceTypeName); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

// ---- Volunteer Blockouts ----

func (s *Service) GetPersonBlockouts(ctx context.Context, tenantID, personID string) ([]VolunteerBlockout, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, person_id, start_date::text, end_date::text, reason, is_recurring, day_of_week, created_at, updated_at
		FROM volunteer_blockouts
		WHERE person_id = $1 AND tenant_id = $2
		ORDER BY start_date`, personID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blockouts := []VolunteerBlockout{}
	for rows.Next() {
		var b VolunteerBlockout
		var dayOfWeek *int
		if err := rows.Scan(&b.ID, &b.TenantID, &b.PersonID, &b.StartDate, &b.EndDate, &b.Reason, &b.IsRecurring, &dayOfWeek, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		if dayOfWeek != nil {
			dow := int(*dayOfWeek)
			b.DayOfWeek = &dow
		}
		blockouts = append(blockouts, b)
	}
	return blockouts, nil
}

func (s *Service) CreateBlockout(ctx context.Context, tenantID, personID string, startDate, endDate, reason string, isRecurring bool, dayOfWeek *int) (*VolunteerBlockout, error) {
	var b VolunteerBlockout
	err := s.db.QueryRow(ctx, `
		INSERT INTO volunteer_blockouts (tenant_id, person_id, start_date, end_date, reason, is_recurring, day_of_week)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, person_id, start_date::text, end_date::text, reason, is_recurring, day_of_week, created_at, updated_at`,
		tenantID, personID, startDate, endDate, reason, isRecurring, dayOfWeek).Scan(
		&b.ID, &b.TenantID, &b.PersonID, &b.StartDate, &b.EndDate, &b.Reason, &b.IsRecurring, &b.DayOfWeek, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) UpdateBlockout(ctx context.Context, tenantID, blockoutID string, startDate, endDate, reason string, isRecurring bool, dayOfWeek *int) (*VolunteerBlockout, error) {
	var b VolunteerBlockout
	err := s.db.QueryRow(ctx, `
		UPDATE volunteer_blockouts SET start_date = $3, end_date = $4, reason = $5, is_recurring = $6, day_of_week = $7, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, person_id, start_date::text, end_date::text, reason, is_recurring, day_of_week, created_at, updated_at`,
		blockoutID, tenantID, startDate, endDate, reason, isRecurring, dayOfWeek).Scan(
		&b.ID, &b.TenantID, &b.PersonID, &b.StartDate, &b.EndDate, &b.Reason, &b.IsRecurring, &b.DayOfWeek, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update blockout: %w", err)
	}
	return &b, nil
}

func (s *Service) DeleteBlockout(ctx context.Context, tenantID, blockoutID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM volunteer_blockouts WHERE id = $1 AND tenant_id = $2`, blockoutID, tenantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("blockout not found")
	}
	return nil
}

// IsVolunteerBlocked checks if a volunteer is blocked on a specific date
// Returns true if blocked, along with the matching blockout info
func (s *Service) IsVolunteerBlocked(ctx context.Context, tenantID, personID string, checkDate time.Time) (*BlockoutMatch, error) {
	// Check non-recurring blockouts where date falls within range
	var b VolunteerBlockout
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, person_id, start_date::text, end_date::text, reason, is_recurring, day_of_week, created_at, updated_at
		FROM volunteer_blockouts
		WHERE person_id = $1 AND tenant_id = $2
		  AND is_recurring = false
		  AND $3 >= start_date AND $3 <= end_date
		LIMIT 1`, personID, tenantID, checkDate).Scan(
		&b.ID, &b.TenantID, &b.PersonID, &b.StartDate, &b.EndDate, &b.Reason, &b.IsRecurring, &b.DayOfWeek, &b.CreatedAt, &b.UpdatedAt)

	if err == nil {
		// Found a matching non-recurring blockout
		return &BlockoutMatch{Blockout: b, IsRecurring: false}, nil
	} else if err != pgx.ErrNoRows {
		// Some other error occurred
		return nil, err
	}

	// Check recurring blockouts by day_of_week
	var b2 VolunteerBlockout
	err = s.db.QueryRow(ctx, `
		SELECT id, tenant_id, person_id, start_date::text, end_date::text, reason, is_recurring, day_of_week, created_at, updated_at
		FROM volunteer_blockouts
		WHERE person_id = $1 AND tenant_id = $2
		  AND is_recurring = true
		  AND day_of_week IS NOT NULL
		  AND EXTRACT(DOW FROM $3) = day_of_week
		LIMIT 1`, personID, tenantID, checkDate).Scan(
		&b2.ID, &b2.TenantID, &b2.PersonID, &b2.StartDate, &b2.EndDate, &b2.Reason, &b2.IsRecurring, &b2.DayOfWeek, &b2.CreatedAt, &b2.UpdatedAt)

	if err == nil {
		return &BlockoutMatch{Blockout: b2, IsRecurring: true}, nil
	} else if err != pgx.ErrNoRows {
		return nil, err
	}

	// No blockout found
	return nil, nil
}

// CheckAssignmentsForConflicts checks all assignments for blockout conflicts
// Returns a map of person_id -> matching blockout for those who are blocked
func (s *Service) CheckAssignmentsForConflicts(ctx context.Context, tenantID string, serviceDate time.Time, personIDs []string) (map[string]*BlockoutMatch, error) {
	conflicts := make(map[string]*BlockoutMatch)

	for _, personID := range personIDs {
		match, err := s.IsVolunteerBlocked(ctx, tenantID, personID, serviceDate)
		if err != nil {
			return nil, fmt.Errorf("conflict check failed for person %s: %w", personID, err)
		}
		if match != nil {
			conflicts[personID] = match
		}
	}

	return conflicts, nil
}
