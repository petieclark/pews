package teams

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
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

	var teams []Team
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
		       sta.status, COALESCE(sta.notes, ''),
		       p.first_name, p.last_name, COALESCE(p.email, ''),
		       tp.name as position_name, t.name as team_name, t.color as team_color
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

	var assignments []ServiceTeamAssignment
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

func (s *Service) SaveServiceAssignments(ctx context.Context, tenantID, serviceID string, assignments []ServiceTeamAssignment) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments for this service
	_, err = tx.Exec(ctx, `DELETE FROM service_team_assignments WHERE service_id = $1 AND tenant_id = $2`, serviceID, tenantID)
	if err != nil {
		return err
	}

	for _, a := range assignments {
		_, err = tx.Exec(ctx, `
			INSERT INTO service_team_assignments (tenant_id, service_id, team_id, position_id, person_id, status, notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, serviceID, a.TeamID, a.PositionID, a.PersonID, a.Status, a.Notes)
		if err != nil {
			return fmt.Errorf("failed to insert assignment: %w", err)
		}
	}

	return tx.Commit(ctx)
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
		       sta.status, COALESCE(sta.notes, ''),
		       p.first_name, p.last_name, COALESCE(p.email, ''),
		       tp.name as position_name, t.name as team_name, t.color as team_color,
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

	var assignments []ServiceTeamAssignment
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
