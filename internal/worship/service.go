package worship

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

// ListPlans returns all service plans for the tenant
func (s *Service) ListPlans(ctx context.Context, tenantID string) ([]ServicePlan, error) {
	query := `
		SELECT id, tenant_id, service_id, created_by, COALESCE(notes, ''), status, created_at, updated_at
		FROM service_plans
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service plans: %w", err)
	}
	defer rows.Close()

	plans := []ServicePlan{}
	for rows.Next() {
		var plan ServicePlan
		err := rows.Scan(
			&plan.ID,
			&plan.TenantID,
			&plan.ServiceID,
			&plan.CreatedBy,
			&plan.Notes,
			&plan.Status,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

// GetPlan returns a service plan with its items
func (s *Service) GetPlan(ctx context.Context, tenantID, planID string) (*ServicePlan, error) {
	query := `
		SELECT id, tenant_id, service_id, created_by, COALESCE(notes, ''), status, created_at, updated_at
		FROM service_plans
		WHERE tenant_id = $1 AND id = $2
	`

	var plan ServicePlan
	err := s.db.QueryRow(ctx, query, tenantID, planID).Scan(
		&plan.ID,
		&plan.TenantID,
		&plan.ServiceID,
		&plan.CreatedBy,
		&plan.Notes,
		&plan.Status,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get service plan: %w", err)
	}

	// Get plan items
	items, err := s.GetPlanItems(ctx, tenantID, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan items: %w", err)
	}
	plan.Items = items

	return &plan, nil
}

// CreatePlan creates a new service plan
func (s *Service) CreatePlan(ctx context.Context, tenantID, serviceID, createdBy, notes string) (*ServicePlan, error) {
	query := `
		INSERT INTO service_plans (tenant_id, service_id, created_by, notes, status)
		VALUES ($1, $2, $3, $4, 'draft')
		RETURNING id, tenant_id, service_id, created_by, notes, status, created_at, updated_at
	`

	var plan ServicePlan
	err := s.db.QueryRow(ctx, query, tenantID, serviceID, createdBy, notes).Scan(
		&plan.ID,
		&plan.TenantID,
		&plan.ServiceID,
		&plan.CreatedBy,
		&plan.Notes,
		&plan.Status,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service plan: %w", err)
	}

	return &plan, nil
}

// UpdatePlan updates a service plan
func (s *Service) UpdatePlan(ctx context.Context, tenantID, planID, notes string) (*ServicePlan, error) {
	query := `
		UPDATE service_plans
		SET notes = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
		RETURNING id, tenant_id, service_id, created_by, notes, status, created_at, updated_at
	`

	var plan ServicePlan
	err := s.db.QueryRow(ctx, query, tenantID, planID, notes).Scan(
		&plan.ID,
		&plan.TenantID,
		&plan.ServiceID,
		&plan.CreatedBy,
		&plan.Notes,
		&plan.Status,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update service plan: %w", err)
	}

	return &plan, nil
}

// PublishPlan changes the plan status to published
func (s *Service) PublishPlan(ctx context.Context, tenantID, planID string) (*ServicePlan, error) {
	query := `
		UPDATE service_plans
		SET status = 'published', updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
		RETURNING id, tenant_id, service_id, created_by, notes, status, created_at, updated_at
	`

	var plan ServicePlan
	err := s.db.QueryRow(ctx, query, tenantID, planID).Scan(
		&plan.ID,
		&plan.TenantID,
		&plan.ServiceID,
		&plan.CreatedBy,
		&plan.Notes,
		&plan.Status,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to publish service plan: %w", err)
	}

	return &plan, nil
}

// GetPlanItems returns all items for a service plan
func (s *Service) GetPlanItems(ctx context.Context, tenantID, planID string) ([]ServicePlanItem, error) {
	query := `
		SELECT 
			spi.id, spi.plan_id, spi.item_order, spi.item_type, COALESCE(spi.title, ''),
			spi.duration_minutes, COALESCE(spi.notes, ''), spi.song_id, spi.assigned_to,
			spi.created_at, spi.updated_at,
			COALESCE(songs.title, '') as song_title,
			COALESCE(users.email, '') as assigned_to_name
		FROM service_plan_items spi
		INNER JOIN service_plans sp ON spi.plan_id = sp.id
		LEFT JOIN songs ON spi.song_id = songs.id
		LEFT JOIN users ON spi.assigned_to = users.id
		WHERE sp.tenant_id = $1 AND spi.plan_id = $2
		ORDER BY spi.item_order ASC
	`

	rows, err := s.db.Query(ctx, query, tenantID, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan items: %w", err)
	}
	defer rows.Close()

	items := []ServicePlanItem{}
	for rows.Next() {
		var item ServicePlanItem
		err := rows.Scan(
			&item.ID,
			&item.PlanID,
			&item.ItemOrder,
			&item.ItemType,
			&item.Title,
			&item.DurationMinutes,
			&item.Notes,
			&item.SongID,
			&item.AssignedTo,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.SongTitle,
			&item.AssignedToName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// AddItem adds a new item to a service plan
func (s *Service) AddItem(ctx context.Context, tenantID, planID string, itemOrder int, itemType, title string, durationMinutes *int, notes string, songID, assignedTo *string) (*ServicePlanItem, error) {
	// Verify the plan belongs to the tenant
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM service_plans WHERE tenant_id = $1 AND id = $2)", tenantID, planID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("service plan not found")
	}

	query := `
		INSERT INTO service_plan_items (plan_id, item_order, item_type, title, duration_minutes, notes, song_id, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, plan_id, item_order, item_type, title, duration_minutes, notes, song_id, assigned_to, created_at, updated_at
	`

	var item ServicePlanItem
	err = s.db.QueryRow(ctx, query, planID, itemOrder, itemType, title, durationMinutes, notes, songID, assignedTo).Scan(
		&item.ID,
		&item.PlanID,
		&item.ItemOrder,
		&item.ItemType,
		&item.Title,
		&item.DurationMinutes,
		&item.Notes,
		&item.SongID,
		&item.AssignedTo,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan item: %w", err)
	}

	return &item, nil
}

// UpdateItem updates an existing item
func (s *Service) UpdateItem(ctx context.Context, tenantID, planID, itemID string, itemOrder int, itemType, title string, durationMinutes *int, notes string, songID, assignedTo *string) (*ServicePlanItem, error) {
	query := `
		UPDATE service_plan_items
		SET item_order = $4, item_type = $5, title = $6, duration_minutes = $7, notes = $8, song_id = $9, assigned_to = $10, updated_at = NOW()
		WHERE id = $3 AND plan_id = $2 AND EXISTS(SELECT 1 FROM service_plans WHERE id = $2 AND tenant_id = $1)
		RETURNING id, plan_id, item_order, item_type, title, duration_minutes, notes, song_id, assigned_to, created_at, updated_at
	`

	var item ServicePlanItem
	err := s.db.QueryRow(ctx, query, tenantID, planID, itemID, itemOrder, itemType, title, durationMinutes, notes, songID, assignedTo).Scan(
		&item.ID,
		&item.PlanID,
		&item.ItemOrder,
		&item.ItemType,
		&item.Title,
		&item.DurationMinutes,
		&item.Notes,
		&item.SongID,
		&item.AssignedTo,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update plan item: %w", err)
	}

	return &item, nil
}

// DeleteItem deletes an item from a service plan
func (s *Service) DeleteItem(ctx context.Context, tenantID, planID, itemID string) error {
	query := `
		DELETE FROM service_plan_items
		WHERE id = $3 AND plan_id = $2 AND EXISTS(SELECT 1 FROM service_plans WHERE id = $2 AND tenant_id = $1)
	`

	result, err := s.db.Exec(ctx, query, tenantID, planID, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete plan item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("plan item not found")
	}

	return nil
}
