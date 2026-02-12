package people

import (
	"context"
	"fmt"

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

func (s *Service) GetDB() *pgxpool.Pool {
	return s.db
}

// People operations

func (s *Service) ListPeople(ctx context.Context, tenantID string, query string, status string, sortBy string, page, limit int, tagIDs ...string) ([]Person, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 10000 {
		limit = 25
	}
	offset := (page - 1) * limit

	// Build query
	sqlQuery := `
		SELECT p.id, p.tenant_id, p.first_name, p.last_name, 
		       COALESCE(p.email, ''), COALESCE(p.phone, ''), 
		       COALESCE(p.address_line1, ''), COALESCE(p.address_line2, ''), 
		       COALESCE(p.city, ''), COALESCE(p.state, ''), COALESCE(p.zip, ''), 
		       p.birthdate, COALESCE(p.gender, ''), p.membership_status, 
		       COALESCE(p.photo_url, ''), COALESCE(p.notes, ''), 
		       COALESCE(p.custom_fields, '{}'), p.created_at, p.updated_at
		FROM people p`

	countQuery := `SELECT COUNT(*) FROM people p`
	args := []interface{}{tenantID}
	argPos := 2

	// Join for tag filter
	if len(tagIDs) > 0 && tagIDs[0] != "" {
		tagJoin := fmt.Sprintf(` JOIN person_tags pt ON pt.person_id = p.id AND pt.tag_id = $%d`, argPos)
		sqlQuery += tagJoin
		countQuery += tagJoin
		args = append(args, tagIDs[0])
		argPos++
	}

	sqlQuery += ` WHERE p.tenant_id = $1`
	countQuery += ` WHERE p.tenant_id = $1`

	if query != "" {
		searchFilter := fmt.Sprintf(` AND (p.first_name ILIKE $%d OR p.last_name ILIKE $%d OR p.email ILIKE $%d OR p.phone ILIKE $%d OR (p.first_name || ' ' || p.last_name) ILIKE $%d)`, argPos, argPos, argPos, argPos, argPos)
		sqlQuery += searchFilter
		countQuery += searchFilter
		args = append(args, "%"+query+"%")
		argPos++
	}

	if status != "" && status != "all" {
		statusFilter := fmt.Sprintf(` AND p.membership_status = $%d`, argPos)
		sqlQuery += statusFilter
		countQuery += statusFilter
		args = append(args, status)
		argPos++
	}

	// Get total count
	var total int
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count people: %w", err)
	}

	// Add sorting
	orderBy := " ORDER BY p.last_name, p.first_name"
	switch sortBy {
	case "name_desc":
		orderBy = " ORDER BY p.last_name DESC, p.first_name DESC"
	case "newest":
		orderBy = " ORDER BY p.created_at DESC"
	case "oldest":
		orderBy = " ORDER BY p.created_at ASC"
	}

	sqlQuery += fmt.Sprintf(`%s LIMIT $%d OFFSET $%d`, orderBy, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list people: %w", err)
	}
	defer rows.Close()

	people := []Person{}
	for rows.Next() {
		var p Person
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.AddressLine1, &p.AddressLine2, &p.City, &p.State, &p.Zip,
			&p.Birthdate, &p.Gender, &p.MembershipStatus, &p.PhotoURL, &p.Notes,
			&p.CustomFields, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan person: %w", err)
		}
		people = append(people, p)
	}

	// Batch load tags for all people
	if len(people) > 0 {
		personIDs := make([]string, len(people))
		for i, p := range people {
			personIDs[i] = p.ID
		}
		tagRows, tagErr := s.db.Query(ctx, `
			SELECT pt.person_id, t.id, t.tenant_id, t.name, t.color, t.created_at
			FROM person_tags pt
			JOIN tags t ON t.id = pt.tag_id
			WHERE pt.person_id = ANY($1)
			ORDER BY t.name`, personIDs)
		if tagErr == nil {
			defer tagRows.Close()
			tagMap := make(map[string][]Tag)
			for tagRows.Next() {
				var personID string
				var tag Tag
				if err := tagRows.Scan(&personID, &tag.ID, &tag.TenantID, &tag.Name, &tag.Color, &tag.CreatedAt); err == nil {
					tagMap[personID] = append(tagMap[personID], tag)
				}
			}
			for i := range people {
				if tags, ok := tagMap[people[i].ID]; ok {
					people[i].Tags = tags
				}
			}
		}
	}

	return people, total, nil
}

func (s *Service) GetPersonByID(ctx context.Context, tenantID, personID string) (*Person, error) {
	var p Person
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, first_name, last_name, 
		       COALESCE(email, ''), COALESCE(phone, ''), 
		       COALESCE(address_line1, ''), COALESCE(address_line2, ''), 
		       COALESCE(city, ''), COALESCE(state, ''), COALESCE(zip, ''), 
		       birthdate, COALESCE(gender, ''), membership_status, 
		       COALESCE(photo_url, ''), COALESCE(notes, ''), 
		       COALESCE(custom_fields, '{}'), created_at, updated_at
		FROM people WHERE id = $1`, personID).Scan(
		&p.ID, &p.TenantID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
		&p.AddressLine1, &p.AddressLine2, &p.City, &p.State, &p.Zip,
		&p.Birthdate, &p.Gender, &p.MembershipStatus, &p.PhotoURL, &p.Notes,
		&p.CustomFields, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	// Load tags
	tags, err2 := s.GetPersonTags(ctx, tenantID, personID)
	if err2 == nil {
		p.Tags = tags
	}

	// Load household
	household, err2 := s.GetPersonHousehold(ctx, tenantID, personID)
	if err2 == nil {
		p.Household = household
	}

	return &p, nil
}

func (s *Service) CreatePerson(ctx context.Context, tenantID string, p *Person) (*Person, error) {
	p.ID = uuid.New().String()
	p.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO people (
			id, tenant_id, first_name, last_name, email, phone, 
			address_line1, address_line2, city, state, zip, 
			birthdate, gender, membership_status, photo_url, notes, custom_fields
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING created_at, updated_at`,
		p.ID, p.TenantID, p.FirstName, p.LastName, p.Email, p.Phone,
		p.AddressLine1, p.AddressLine2, p.City, p.State, p.Zip,
		p.Birthdate, p.Gender, p.MembershipStatus, p.PhotoURL, p.Notes, p.CustomFields,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	return p, nil
}

func (s *Service) UpdatePerson(ctx context.Context, tenantID, personID string, p *Person) (*Person, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE people SET 
			first_name = $1, last_name = $2, email = $3, phone = $4,
			address_line1 = $5, address_line2 = $6, city = $7, state = $8, zip = $9,
			birthdate = $10, gender = $11, membership_status = $12, photo_url = $13, 
			notes = $14, custom_fields = $15
		WHERE id = $16
		RETURNING created_at, updated_at`,
		p.FirstName, p.LastName, p.Email, p.Phone,
		p.AddressLine1, p.AddressLine2, p.City, p.State, p.Zip,
		p.Birthdate, p.Gender, p.MembershipStatus, p.PhotoURL, p.Notes, p.CustomFields,
		personID,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to update person: %w", err)
	}

	p.ID = personID
	p.TenantID = tenantID

	return p, nil
}

func (s *Service) DeletePerson(ctx context.Context, tenantID, personID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM people WHERE id = $1", personID)
	if err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("person not found")
	}

	return nil
}

func (s *Service) BulkUpdateStatus(ctx context.Context, tenantID string, personIDs []string, status string) (int64, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE people SET membership_status = $1 
		WHERE tenant_id = $2 AND id = ANY($3)`,
		status, tenantID, personIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk update: %w", err)
	}
	return result.RowsAffected(), nil
}

// Tag operations

func (s *Service) GetPersonTags(ctx context.Context, tenantID, personID string) ([]Tag, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.tenant_id, t.name, t.color, t.created_at
		FROM tags t
		JOIN person_tags pt ON pt.tag_id = t.id
		WHERE pt.person_id = $1
		ORDER BY t.name`, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person tags: %w", err)
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.TenantID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *Service) AddTagToPerson(ctx context.Context, tenantID, personID, tagID string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO person_tags (person_id, tag_id) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING`, personID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to person: %w", err)
	}

	return nil
}

func (s *Service) RemoveTagFromPerson(ctx context.Context, tenantID, personID, tagID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM person_tags WHERE person_id = $1 AND tag_id = $2", personID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from person: %w", err)
	}

	return nil
}

func (s *Service) ListTags(ctx context.Context, tenantID string) ([]Tag, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.tenant_id, t.name, t.color, t.created_at, COUNT(pt.person_id)
		FROM tags t
		LEFT JOIN person_tags pt ON pt.tag_id = t.id
		WHERE t.tenant_id = $1
		GROUP BY t.id
		ORDER BY t.name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.TenantID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.PersonCount); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *Service) DeleteTag(ctx context.Context, tenantID, tagID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM tags WHERE id = $1 AND tenant_id = $2", tagID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func (s *Service) AddTagsToPerson(ctx context.Context, tenantID, personID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		_, err := s.db.Exec(ctx, `
			INSERT INTO person_tags (person_id, tag_id)
			SELECT $1, $2 WHERE EXISTS (SELECT 1 FROM tags WHERE id = $2 AND tenant_id = $3)
			ON CONFLICT DO NOTHING`, personID, tagID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to add tag %s: %w", tagID, err)
		}
	}
	return nil
}

func (s *Service) BulkAddTag(ctx context.Context, tenantID string, personIDs []string, tagID string) (int64, error) {
	var count int64
	for _, pid := range personIDs {
		result, err := s.db.Exec(ctx, `
			INSERT INTO person_tags (person_id, tag_id)
			SELECT $1, $2 WHERE EXISTS (SELECT 1 FROM tags WHERE id = $2 AND tenant_id = $3)
			ON CONFLICT DO NOTHING`, pid, tagID, tenantID)
		if err != nil {
			return count, fmt.Errorf("failed to bulk add tag: %w", err)
		}
		count += result.RowsAffected()
	}
	return count, nil
}

func (s *Service) BulkRemoveTag(ctx context.Context, tenantID string, personIDs []string, tagID string) (int64, error) {
	result, err := s.db.Exec(ctx, `
		DELETE FROM person_tags WHERE tag_id = $1 AND person_id = ANY($2)`, tagID, personIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk remove tag: %w", err)
	}
	return result.RowsAffected(), nil
}

// EnsureTagForTenant creates a tag if it doesn't exist, returns the tag ID.
func (s *Service) EnsureTagForTenant(ctx context.Context, tenantID, name, color string) (string, error) {
	var tagID string
	err := s.db.QueryRow(ctx, `
		INSERT INTO tags (id, tenant_id, name, color)
		VALUES (gen_random_uuid(), $1, $2, $3)
		ON CONFLICT (tenant_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, tenantID, name, color).Scan(&tagID)
	if err != nil {
		return "", fmt.Errorf("failed to ensure tag: %w", err)
	}
	return tagID, nil
}

// AutoTagPerson ensures a tag exists and adds it to a person.
func (s *Service) AutoTagPerson(ctx context.Context, tenantID, personID, tagName string) error {
	tagID, err := s.EnsureTagForTenant(ctx, tenantID, tagName, "#4A8B8C")
	if err != nil {
		return err
	}
	return s.AddTagToPerson(ctx, tenantID, personID, tagID)
}

func (s *Service) CreateTag(ctx context.Context, tenantID string, tag *Tag) (*Tag, error) {
	tag.ID = uuid.New().String()
	tag.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO tags (id, tenant_id, name, color) 
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		tag.ID, tag.TenantID, tag.Name, tag.Color,
	).Scan(&tag.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// Household operations

func (s *Service) GetPersonHousehold(ctx context.Context, tenantID, personID string) (*Household, error) {
	var h Household
	err := s.db.QueryRow(ctx, `
		SELECT h.id, h.tenant_id, h.name, COALESCE(h.primary_contact_id, ''), 
		       COALESCE(h.address_line1, ''), COALESCE(h.address_line2, ''), 
		       COALESCE(h.city, ''), COALESCE(h.state, ''), COALESCE(h.zip, ''),
		       h.created_at, h.updated_at
		FROM households h
		JOIN household_members hm ON hm.household_id = h.id
		WHERE hm.person_id = $1
		LIMIT 1`, personID).Scan(
		&h.ID, &h.TenantID, &h.Name, &h.PrimaryContactID,
		&h.AddressLine1, &h.AddressLine2, &h.City, &h.State, &h.Zip,
		&h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get household: %w", err)
	}

	return &h, nil
}

func (s *Service) ListHouseholds(ctx context.Context, tenantID string) ([]Household, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, name, COALESCE(primary_contact_id, ''), 
		       COALESCE(address_line1, ''), COALESCE(address_line2, ''), 
		       COALESCE(city, ''), COALESCE(state, ''), COALESCE(zip, ''),
		       created_at, updated_at
		FROM households
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list households: %w", err)
	}
	defer rows.Close()

	households := []Household{}
	for rows.Next() {
		var h Household
		if err := rows.Scan(
			&h.ID, &h.TenantID, &h.Name, &h.PrimaryContactID,
			&h.AddressLine1, &h.AddressLine2, &h.City, &h.State, &h.Zip,
			&h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan household: %w", err)
		}
		households = append(households, h)
	}

	return households, nil
}

func (s *Service) CreateHousehold(ctx context.Context, tenantID string, h *Household) (*Household, error) {
	h.ID = uuid.New().String()
	h.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO households (
			id, tenant_id, name, primary_contact_id,
			address_line1, address_line2, city, state, zip
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`,
		h.ID, h.TenantID, h.Name, h.PrimaryContactID,
		h.AddressLine1, h.AddressLine2, h.City, h.State, h.Zip,
	).Scan(&h.CreatedAt, &h.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create household: %w", err)
	}

	return h, nil
}

func (s *Service) UpdateHousehold(ctx context.Context, tenantID, householdID string, h *Household) (*Household, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE households SET 
			name = $1, primary_contact_id = $2,
			address_line1 = $3, address_line2 = $4, city = $5, state = $6, zip = $7
		WHERE id = $8
		RETURNING created_at, updated_at`,
		h.Name, h.PrimaryContactID,
		h.AddressLine1, h.AddressLine2, h.City, h.State, h.Zip,
		householdID,
	).Scan(&h.CreatedAt, &h.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("household not found")
		}
		return nil, fmt.Errorf("failed to update household: %w", err)
	}

	h.ID = householdID
	h.TenantID = tenantID

	return h, nil
}

func (s *Service) AddMemberToHousehold(ctx context.Context, tenantID, householdID, personID, role string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO household_members (household_id, person_id, role) 
		VALUES ($1, $2, $3)
		ON CONFLICT (household_id, person_id) DO UPDATE SET role = $3`,
		householdID, personID, role)
	if err != nil {
		return fmt.Errorf("failed to add member to household: %w", err)
	}

	return nil
}

func (s *Service) RemoveMemberFromHousehold(ctx context.Context, tenantID, householdID, personID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM household_members WHERE household_id = $1 AND person_id = $2", householdID, personID)
	if err != nil {
		return fmt.Errorf("failed to remove member from household: %w", err)
	}

	return nil
}
