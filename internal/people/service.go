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

// People operations

func (s *Service) ListPeople(ctx context.Context, tenantID string, query string, page, limit int) ([]Person, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Build query
	sqlQuery := `
		SELECT id, tenant_id, first_name, last_name, 
		       COALESCE(email, ''), COALESCE(phone, ''), 
		       COALESCE(address_line1, ''), COALESCE(address_line2, ''), 
		       COALESCE(city, ''), COALESCE(state, ''), COALESCE(zip, ''), 
		       birthdate, COALESCE(gender, ''), membership_status, 
		       COALESCE(photo_url, ''), COALESCE(notes, ''), 
		       COALESCE(custom_fields, '{}'), created_at, updated_at
		FROM people
		WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM people WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if query != "" {
		searchFilter := fmt.Sprintf(` AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)`, argPos, argPos, argPos, argPos)
		sqlQuery += searchFilter
		countQuery += searchFilter
		args = append(args, "%"+query+"%")
		argPos++
	}

	// Get total count
	var total int
	err = s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count people: %w", err)
	}

	// Add pagination
	sqlQuery += fmt.Sprintf(` ORDER BY last_name, first_name LIMIT $%d OFFSET $%d`, argPos, argPos+1)
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

	return people, total, nil
}

func (s *Service) GetPersonByID(ctx context.Context, tenantID, personID string) (*Person, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var p Person
	err = s.db.QueryRow(ctx, `
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
	tags, err := s.GetPersonTags(ctx, tenantID, personID)
	if err == nil {
		p.Tags = tags
	}

	// Load household
	household, err := s.GetPersonHousehold(ctx, tenantID, personID)
	if err == nil {
		p.Household = household
	}

	return &p, nil
}

func (s *Service) CreatePerson(ctx context.Context, tenantID string, p *Person) (*Person, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	p.ID = uuid.New().String()
	p.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	result, err := s.db.Exec(ctx, "DELETE FROM people WHERE id = $1", personID)
	if err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("person not found")
	}

	return nil
}

// Tag operations

func (s *Service) GetPersonTags(ctx context.Context, tenantID, personID string) ([]Tag, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO person_tags (person_id, tag_id) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING`, personID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to person: %w", err)
	}

	return nil
}

func (s *Service) RemoveTagFromPerson(ctx context.Context, tenantID, personID, tagID string) error {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	_, err = s.db.Exec(ctx, "DELETE FROM person_tags WHERE person_id = $1 AND tag_id = $2", personID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from person: %w", err)
	}

	return nil
}

func (s *Service) ListTags(ctx context.Context, tenantID string) ([]Tag, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, name, color, created_at
		FROM tags
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
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

func (s *Service) CreateTag(ctx context.Context, tenantID string, tag *Tag) (*Tag, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	tag.ID = uuid.New().String()
	tag.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var h Household
	err = s.db.QueryRow(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	h.ID = uuid.New().String()
	h.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	_, err = s.db.Exec(ctx, `
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
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	_, err = s.db.Exec(ctx, "DELETE FROM household_members WHERE household_id = $1 AND person_id = $2", householdID, personID)
	if err != nil {
		return fmt.Errorf("failed to remove member from household: %w", err)
	}

	return nil
}
