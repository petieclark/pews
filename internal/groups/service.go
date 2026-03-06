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
		WHERE g.tenant_id = $1`

	countQuery := `SELECT COUNT(DISTINCT g.id) FROM groups g WHERE g.tenant_id = $1`
	args := []interface{}{tenantID}
	argPos := 2

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

	// Auto-tag person with group name
	var groupName string
	_ = s.db.QueryRow(ctx, "SELECT name FROM groups WHERE id = $1", groupID).Scan(&groupName)
	if groupName != "" {
		s.autoTagPerson(ctx, tenantID, personID, groupName)
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

func (s *Service) GetPersonGroups(ctx context.Context, tenantID, personID string) (interface{}, error) {
	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, ''), g.group_type, 
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''), COALESCE(g.meeting_location, ''), 
		       g.is_public, g.max_members, g.is_active, COALESCE(g.photo_url, ''), 
		       g.created_at, g.updated_at, gm.role
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.person_id = $1 AND g.tenant_id = $2
		ORDER BY g.name`, personID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person groups: %w", err)
	}
	defer rows.Close()

	groups := []map[string]interface{}{}
	for rows.Next() {
		var g Group
		var role string
		err := rows.Scan(
			&g.ID, &g.TenantID, &g.Name, &g.Description, &g.GroupType,
			&g.MeetingDay, &g.MeetingTime, &g.MeetingLocation,
			&g.IsPublic, &g.MaxMembers, &g.IsActive, &g.PhotoURL,
			&g.CreatedAt, &g.UpdatedAt, &role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		m := map[string]interface{}{
			"id": g.ID, "tenant_id": g.TenantID, "name": g.Name,
			"description": g.Description, "group_type": g.GroupType,
			"meeting_day": g.MeetingDay, "meeting_time": g.MeetingTime,
			"meeting_location": g.MeetingLocation, "is_public": g.IsPublic,
			"max_members": g.MaxMembers, "is_active": g.IsActive,
			"photo_url": g.PhotoURL, "created_at": g.CreatedAt,
			"updated_at": g.UpdatedAt, "role": role,
		}
		groups = append(groups, m)
	}

	return groups, nil
}

func (s *Service) ListPublicGroups(ctx context.Context, tenantID string, category string, meetingDay string, search string) ([]Group, error) {
	query := `
		SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, ''), g.group_type,
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''), COALESCE(g.meeting_location, ''),
		       g.is_public, g.max_members, g.is_active, COALESCE(g.photo_url, ''),
		       g.created_at, g.updated_at,
		       COUNT(DISTINCT gm.id) as member_count
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id
		WHERE g.tenant_id = $1 AND g.is_public = TRUE AND g.is_active = TRUE`
	args := []interface{}{tenantID}
	argPos := 2

	if category != "" {
		query += fmt.Sprintf(` AND g.group_type = $%d`, argPos)
		args = append(args, category)
		argPos++
	}
	if meetingDay != "" {
		query += fmt.Sprintf(` AND g.meeting_day = $%d`, argPos)
		args = append(args, meetingDay)
		argPos++
	}
	if search != "" {
		query += fmt.Sprintf(` AND (g.name ILIKE $%d OR g.description ILIKE $%d)`, argPos, argPos)
		args = append(args, "%"+search+"%")
		argPos++
	}

	query += ` GROUP BY g.id ORDER BY g.name`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list public groups: %w", err)
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
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (s *Service) GetPublicGroup(ctx context.Context, tenantID, groupID string) (*Group, error) {
	var g Group
	err := s.db.QueryRow(ctx, `
		SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, ''), g.group_type,
		       COALESCE(g.meeting_day, ''), COALESCE(g.meeting_time, ''), COALESCE(g.meeting_location, ''),
		       g.is_public, g.max_members, g.is_active, COALESCE(g.photo_url, ''),
		       g.created_at, g.updated_at,
		       COUNT(DISTINCT gm.id) as member_count
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id
		WHERE g.id = $1 AND g.tenant_id = $2 AND g.is_public = TRUE AND g.is_active = TRUE
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
		return nil, fmt.Errorf("failed to get public group: %w", err)
	}
	return &g, nil
}

func (s *Service) CreateJoinRequest(ctx context.Context, tenantID string, req *JoinRequest) (*JoinRequest, error) {
	req.ID = uuid.New().String()
	req.TenantID = tenantID
	req.Status = "pending"

	err := s.db.QueryRow(ctx, `
		INSERT INTO group_join_requests (id, tenant_id, group_id, name, email, phone, message, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`,
		req.ID, req.TenantID, req.GroupID, req.Name, req.Email, req.Phone, req.Message, req.Status,
	).Scan(&req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create join request: %w", err)
	}

	return req, nil
}

func (s *Service) ListJoinRequests(ctx context.Context, tenantID string, groupID string, status string) ([]JoinRequest, error) {
	query := `
		SELECT jr.id, jr.tenant_id, jr.group_id, jr.name, jr.email, COALESCE(jr.phone, ''),
		       COALESCE(jr.message, ''), jr.status, jr.created_at, jr.updated_at,
		       COALESCE(g.name, '') as group_name
		FROM group_join_requests jr
		JOIN groups g ON g.id = jr.group_id
		WHERE jr.tenant_id = $1`
	args := []interface{}{tenantID}
	argPos := 2

	if groupID != "" {
		query += fmt.Sprintf(` AND jr.group_id = $%d`, argPos)
		args = append(args, groupID)
		argPos++
	}
	if status != "" {
		query += fmt.Sprintf(` AND jr.status = $%d`, argPos)
		args = append(args, status)
		argPos++
	}

	query += ` ORDER BY jr.created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list join requests: %w", err)
	}
	defer rows.Close()

	requests := []JoinRequest{}
	for rows.Next() {
		var jr JoinRequest
		err := rows.Scan(
			&jr.ID, &jr.TenantID, &jr.GroupID, &jr.Name, &jr.Email, &jr.Phone,
			&jr.Message, &jr.Status, &jr.CreatedAt, &jr.UpdatedAt, &jr.GroupName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan join request: %w", err)
		}
		requests = append(requests, jr)
	}

	return requests, nil
}

func (s *Service) GetJoinRequest(ctx context.Context, tenantID, requestID string) (*JoinRequest, error) {
	var jr JoinRequest
	err := s.db.QueryRow(ctx, `
		SELECT jr.id, jr.tenant_id, jr.group_id, jr.name, jr.email, COALESCE(jr.phone, ''),
		       COALESCE(jr.message, ''), jr.status, jr.created_at, jr.updated_at,
		       COALESCE(g.name, '') as group_name
		FROM group_join_requests jr
		JOIN groups g ON g.id = jr.group_id
		WHERE jr.id = $1 AND jr.tenant_id = $2`, requestID, tenantID).Scan(
		&jr.ID, &jr.TenantID, &jr.GroupID, &jr.Name, &jr.Email, &jr.Phone,
		&jr.Message, &jr.Status, &jr.CreatedAt, &jr.UpdatedAt, &jr.GroupName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("join request not found")
		}
		return nil, fmt.Errorf("failed to get join request: %w", err)
	}
	return &jr, nil
}

func (s *Service) UpdateJoinRequestStatus(ctx context.Context, tenantID, requestID, status string) error {
	result, err := s.db.Exec(ctx,
		`UPDATE group_join_requests SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`,
		status, requestID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update join request status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("join request not found")
	}
	return nil
}

func (s *Service) ApproveJoinRequest(ctx context.Context, tenantID, requestID string) error {
	// Get the join request
	jr, err := s.GetJoinRequest(ctx, tenantID, requestID)
	if err != nil {
		return err
	}

	if jr.Status != "pending" {
		return fmt.Errorf("join request is not pending")
	}

	// Update status to approved
	if err := s.UpdateJoinRequestStatus(ctx, tenantID, requestID, "approved"); err != nil {
		return err
	}

	return nil
}
