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
		`SELECT id, name, slug, COALESCE(domain, ''), plan, created_at, updated_at FROM tenants WHERE slug = $1`,
		slug,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Plan, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	return tenant, nil
}

func (s *Service) GetTenantByID(ctx context.Context, id string) (*Tenant, error) {
	tenant := &Tenant{}
	err := s.db.QueryRow(ctx,
		`SELECT id, name, slug, COALESCE(domain, ''), plan, created_at, updated_at FROM tenants WHERE id = $1`,
		id,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Plan, &tenant.CreatedAt, &tenant.UpdatedAt)

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
