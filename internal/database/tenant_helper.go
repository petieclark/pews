package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SetTenantContext manually sets the tenant context for the current connection.
// This should ONLY be used for public endpoints that don't go through auth middleware.
// For authenticated routes, the middleware handles this automatically.
func SetTenantContext(ctx context.Context, pool *pgxpool.Pool, tenantID string) error {
	_, err := pool.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}
