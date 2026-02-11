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
		       t.created_at, t.updated_at, COUNT(tm.id) as member_count
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
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
			&t.CreatedAt, &t.UpdatedAt, &t.MemberCount); err != nil {
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
	return &m, nil
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
