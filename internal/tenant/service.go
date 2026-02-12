package tenant

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func (s *Service) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	tenant := &Tenant{
		ID:   uuid.New().String(),
		Name: name,
		Slug: slugify(name),
		Plan: "free",
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO tenants (id, name, slug, domain, plan) VALUES ($1, $2, $3, $4, $5)`,
		tenant.ID, tenant.Name, tenant.Slug, "", tenant.Plan,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Initialize default modules (all disabled)
	modules := []string{"people", "giving", "services", "groups", "checkins"}
	for _, moduleName := range modules {
		_, err := s.db.Exec(ctx,
			`INSERT INTO tenant_modules (tenant_id, module_name, enabled) VALUES ($1, $2, $3)`,
			tenant.ID, moduleName, false,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize module %s: %w", moduleName, err)
		}
	}

	return tenant, nil
}

func (s *Service) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	tenant := &Tenant{}
	err := s.db.QueryRow(ctx,
		`SELECT id, name, slug, COALESCE(domain, ''), plan, 
		        COALESCE(address_line1, ''), COALESCE(address_line2, ''), 
		        COALESCE(city, ''), COALESCE(state, ''), COALESCE(zip, ''),
		        COALESCE(phone, ''), COALESCE(website, ''), COALESCE(email, ''),
		        COALESCE(ein, ''), COALESCE(logo, ''), COALESCE(about, ''),
		        COALESCE(onboarding_completed, FALSE),
		        created_at, updated_at 
		 FROM tenants WHERE slug = $1`,
		slug,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Plan,
		&tenant.AddressLine1, &tenant.AddressLine2, &tenant.City, &tenant.State, &tenant.Zip,
		&tenant.Phone, &tenant.Website, &tenant.Email, &tenant.EIN, &tenant.Logo, &tenant.About,
		&tenant.OnboardingCompleted,
		&tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	return tenant, nil
}

func (s *Service) GetTenantByID(ctx context.Context, id string) (*Tenant, error) {
	tenant := &Tenant{}
	err := s.db.QueryRow(ctx,
		`SELECT id, name, slug, COALESCE(domain, ''), plan, 
		        COALESCE(address_line1, ''), COALESCE(address_line2, ''), 
		        COALESCE(city, ''), COALESCE(state, ''), COALESCE(zip, ''),
		        COALESCE(phone, ''), COALESCE(website, ''), COALESCE(email, ''),
		        COALESCE(ein, ''), COALESCE(logo, ''), COALESCE(about, ''),
		        COALESCE(onboarding_completed, FALSE),
		        created_at, updated_at 
		 FROM tenants WHERE id = $1`,
		id,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Plan,
		&tenant.AddressLine1, &tenant.AddressLine2, &tenant.City, &tenant.State, &tenant.Zip,
		&tenant.Phone, &tenant.Website, &tenant.Email, &tenant.EIN, &tenant.Logo, &tenant.About,
		&tenant.OnboardingCompleted,
		&tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	return tenant, nil
}

func (s *Service) UpdateTenant(ctx context.Context, id string, name, domain string) (*Tenant, error) {
	slug := slugify(name)

	_, err := s.db.Exec(ctx,
		`UPDATE tenants SET name = $1, slug = $2, domain = $3 WHERE id = $4`,
		name, slug, domain, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return s.GetTenantByID(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) (*Tenant, error) {
	_, err := s.db.Exec(ctx,
		`UPDATE tenants SET 
			name = $1, 
			address_line1 = $2, 
			address_line2 = $3,
			city = $4,
			state = $5,
			zip = $6,
			phone = $7,
			website = $8,
			email = $9,
			ein = $10,
			about = $11
		WHERE id = $12`,
		req.Name, req.AddressLine1, req.AddressLine2, req.City, req.State, req.Zip,
		req.Phone, req.Website, req.Email, req.EIN, req.About, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.GetTenantByID(ctx, id)
}

func (s *Service) SetOnboardingCompleted(ctx context.Context, id string, completed bool) error {
	_, err := s.db.Exec(ctx,
		`UPDATE tenants SET onboarding_completed = $1 WHERE id = $2`,
		completed, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}
	return nil
}

func (s *Service) UpdateLogo(ctx context.Context, id string, logo string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE tenants SET logo = $1 WHERE id = $2`,
		logo, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update logo: %w", err)
	}

	return nil
}
