package module

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetTenantModules(ctx context.Context, tenantID string) ([]TenantModule, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, tenant_id, module_name, enabled, enabled_at, created_at, updated_at 
		 FROM tenant_modules WHERE tenant_id = $1 ORDER BY module_name`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer rows.Close()

	var modules []TenantModule
	for rows.Next() {
		var m TenantModule
		err := rows.Scan(&m.ID, &m.TenantID, &m.ModuleName, &m.Enabled, &m.EnabledAt, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}
		modules = append(modules, m)
	}

	return modules, nil
}

func (s *Service) EnableModule(ctx context.Context, tenantID, moduleName string) error {
	// Verify module exists in registry
	if GetModuleByName(moduleName) == nil {
		return fmt.Errorf("module not found: %s", moduleName)
	}

	now := time.Now()
	_, err := s.db.Exec(ctx,
		`UPDATE tenant_modules SET enabled = true, enabled_at = $1 
		 WHERE tenant_id = $2 AND module_name = $3`,
		now, tenantID, moduleName,
	)
	if err != nil {
		return fmt.Errorf("failed to enable module: %w", err)
	}

	return nil
}

func (s *Service) DisableModule(ctx context.Context, tenantID, moduleName string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE tenant_modules SET enabled = false, enabled_at = NULL 
		 WHERE tenant_id = $1 AND module_name = $2`,
		tenantID, moduleName,
	)
	if err != nil {
		return fmt.Errorf("failed to disable module: %w", err)
	}

	return nil
}
