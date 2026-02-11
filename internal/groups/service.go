package groups

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petieclark/pews/internal/people"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Group operations

func (s *Service) ListGroups(ctx context.Context, tenantID string, groupType string, active *bool, page, limit int) ([]Group, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	sqlQuery := `
		SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, ''), g.group_type, 
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''), COALESCE(g.meeting_location, ''), 
		       g.is_public, g.max_members, g.is_active, COALESCE(g.photo_url, ''), 
		       g.created_at, g.updated_at,
		       COUNT(DISTINCT gm.id) as member_count
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id
		WHERE 1=1`

	countQuery := `SELECT COUNT(DISTINCT g.id) FROM groups g WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if groupType != "" {
		filter := fmt.Sprintf(` AND g.group_type = $%d`, argPos)
		sqlQuery += filter
		countQuery += filter
		args = append(args, groupType)
		argPos++
	}

	if active != nil {
		filter := fmt.Sprintf(` AND g.is_active = $%d`, argPos)
		sqlQuery += filter
		countQuery += filter
		args = append(args, *active)
		argPos++
	}

	// Get total count
	var total int
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	// Add grouping, ordering, and pagination
	sqlQuery += ` GROUP BY g.id ORDER BY g.name`
	sqlQuery += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		var g Group
		err := rows.Scan(
			&g.ID, &g.TenantID, &g.Name, &g.Description, &g.GroupType,
			&g.MeetingDay, &g.MeetingTime, &g.MeetingLocation,
			&g.IsPublic, &g.MaxMembers, &g.IsActive, &g.PhotoURL,
			&g.CreatedAt, &g.UpdatedAt, &g.MemberCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, g)
	}

	return groups, total, nil
}

func (s *Service) GetGroupByID(ctx context.Context, tenantID, groupID string) (*Group, error) {
	var g Group
	err := s.db.QueryRow(ctx, `
		SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, ''), g.group_type, 
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''), COALESCE(g.meeting_location, ''), 
		       g.is_public, g.max_members, g.is_active, COALESCE(g.photo_url, ''), 
		       g.created_at, g.updated_at,
		       COUNT(DISTINCT gm.id) as member_count
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id
		WHERE g.id = $1 AND g.tenant_id = $2
		GROUP BY g.id`, groupID, tenantID).Scan(
		&g.ID, &g.TenantID, &g.Name, &g.Description, &g.GroupType,
		&g.MeetingDay, &g.MeetingTime, &g.MeetingLocation,
		&g.IsPublic, &g.MaxMembers, &g.IsActive, &g.PhotoURL,
		&g.CreatedAt, &g.UpdatedAt, &g.MemberCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Load members
	members, err2 := s.GetGroupMembers(ctx, tenantID, groupID)
	if err2 == nil {
		g.Members = members
	}

	return &g, nil
}

func (s *Service) CreateGroup(ctx context.Context, tenantID string, g *Group) (*Group, error) {
	g.ID = uuid.New().String()
	g.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO groups (
			id, tenant_id, name, description, group_type, 
			meeting_day, meeting_time, meeting_location, 
			is_public, max_members, is_active, photo_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`,
		g.ID, g.TenantID, g.Name, g.Description, g.GroupType,
		g.MeetingDay, g.MeetingTime, g.MeetingLocation,
		g.IsPublic, g.MaxMembers, g.IsActive, g.PhotoURL,
	).Scan(&g.CreatedAt, &g.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return g, nil
}

func (s *Service) UpdateGroup(ctx context.Context, tenantID, groupID string, g *Group) (*Group, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE groups SET 
			name = $1, description = $2, group_type = $3, 
			meeting_day = $4, meeting_time = $5, meeting_location = $6, 
			is_public = $7, max_members = $8, is_active = $9, photo_url = $10
		WHERE id = $11 AND tenant_id = $12
		RETURNING created_at, updated_at`,
		g.Name, g.Description, g.GroupType,
		g.MeetingDay, g.MeetingTime, g.MeetingLocation,
		g.IsPublic, g.MaxMembers, g.IsActive, g.PhotoURL,
		groupID, tenantID,
	).Scan(&g.CreatedAt, &g.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	g.ID = groupID
	g.TenantID = tenantID

	return g, nil
}

func (s *Service) DeleteGroup(ctx context.Context, tenantID, groupID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM groups WHERE id = $1 AND tenant_id = $2", groupID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// Member operations

func (s *Service) GetGroupMembers(ctx context.Context, tenantID, groupID string) ([]Member, error) {
	rows, err := s.db.Query(ctx, `
		SELECT gm.id, gm.group_id, gm.person_id, gm.role, gm.joined_at,
		       p.id, p.tenant_id, p.first_name, p.last_name, COALESCE(p.email, ''), COALESCE(p.phone, ''),
		       COALESCE(p.address_line1, ''), COALESCE(p.address_line2, ''), COALESCE(p.city, ''), COALESCE(p.state, ''), COALESCE(p.zip, ''),
		       p.birthdate, COALESCE(p.gender, ''), p.membership_status, COALESCE(p.photo_url, ''), COALESCE(p.notes, ''),
		       COALESCE(p.custom_fields, '{}'), p.created_at, p.updated_at
		FROM group_members gm
		JOIN people p ON p.id = gm.person_id
		JOIN groups g ON g.id = gm.group_id
		WHERE gm.group_id = $1 AND g.tenant_id = $2
		ORDER BY gm.role DESC, p.last_name, p.first_name`, groupID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	members := []Member{}
	for rows.Next() {
		var m Member
		p := &people.Person{}
		err := rows.Scan(
			&m.ID, &m.GroupID, &m.PersonID, &m.Role, &m.JoinedAt,
			&p.ID, &p.TenantID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.AddressLine1, &p.AddressLine2, &p.City, &p.State, &p.Zip,
			&p.Birthdate, &p.Gender, &p.MembershipStatus, &p.PhotoURL, &p.Notes,
			&p.CustomFields, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		m.Person = p
		members = append(members, m)
	}

	return members, nil
}

func (s *Service) AddMemberToGroup(ctx context.Context, tenantID, groupID, personID, role string) (*Member, error) {
	// Verify group belongs to tenant
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1 AND tenant_id = $2)", groupID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("group not found or access denied")
	}

	// Verify person belongs to tenant
	err = s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM people WHERE id = $1 AND tenant_id = $2)", personID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("person not found or access denied")
	}

	memberID := uuid.New().String()
	var m Member
	err = s.db.QueryRow(ctx, `
		INSERT INTO group_members (id, group_id, person_id, role) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, group_id, person_id, role, joined_at`,
		memberID, groupID, personID, role,
	).Scan(&m.ID, &m.GroupID, &m.PersonID, &m.Role, &m.JoinedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add member to group: %w", err)
	}

	return &m, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, tenantID, memberID, role string) error {
	// Verify member belongs to a group owned by tenant
	result, err := s.db.Exec(ctx, `
		UPDATE group_members gm
		SET role = $1 
		FROM groups g
		WHERE gm.id = $2 AND gm.group_id = g.id AND g.tenant_id = $3`,
		role, memberID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (s *Service) RemoveMemberFromGroup(ctx context.Context, tenantID, memberID string) error {
	// Verify member belongs to a group owned by tenant
	result, err := s.db.Exec(ctx, `
		DELETE FROM group_members gm
		USING groups g
		WHERE gm.id = $1 AND gm.group_id = g.id AND g.tenant_id = $2`, memberID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to remove member from group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (s *Service) GetPersonGroups(ctx context.Context, tenantID, personID string) ([]Group, error) {
	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, ''), g.group_type, 
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''), COALESCE(g.meeting_location, ''), 
		       g.is_public, g.max_members, g.is_active, COALESCE(g.photo_url, ''), 
		       g.created_at, g.updated_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.person_id = $1 AND g.tenant_id = $2
		ORDER BY g.name`, personID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person groups: %w", err)
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		var g Group
		err := rows.Scan(
			&g.ID, &g.TenantID, &g.Name, &g.Description, &g.GroupType,
			&g.MeetingDay, &g.MeetingTime, &g.MeetingLocation,
			&g.IsPublic, &g.MaxMembers, &g.IsActive, &g.PhotoURL,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, g)
	}

	return groups, nil
}
