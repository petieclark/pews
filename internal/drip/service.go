package drip

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func joinStrings(arr []string, sep string) string {
	return strings.Join(arr, sep)
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Campaign CRUD

func (s *Service) ListCampaigns(ctx context.Context, tenantID string) ([]Campaign, error) {
	query := `
		SELECT 
			c.id, c.tenant_id, c.name, c.trigger_event, c.is_active, 
			c.created_at, c.updated_at,
			COUNT(DISTINCT e.id) as enrollment_count
		FROM drip_campaigns c
		LEFT JOIN drip_enrollments e ON e.campaign_id = c.id AND e.status = 'active'
		WHERE c.tenant_id = $1
		GROUP BY c.id
		ORDER BY c.created_at DESC
	`

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	campaigns := []Campaign{}
	for rows.Next() {
		var c Campaign
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.TriggerEvent, &c.IsActive,
			&c.CreatedAt, &c.UpdatedAt, &c.EnrollmentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

func (s *Service) GetCampaign(ctx context.Context, tenantID, campaignID string) (*Campaign, error) {
	query := `
		SELECT 
			c.id, c.tenant_id, c.name, c.trigger_event, c.is_active, 
			c.created_at, c.updated_at,
			COUNT(DISTINCT e.id) as enrollment_count
		FROM drip_campaigns c
		LEFT JOIN drip_enrollments e ON e.campaign_id = c.id AND e.status = 'active'
		WHERE c.tenant_id = $1 AND c.id = $2
		GROUP BY c.id
	`

	var c Campaign
	err := s.db.QueryRow(ctx, query, tenantID, campaignID).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.TriggerEvent, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt, &c.EnrollmentCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Load steps
	steps, err := s.ListSteps(ctx, tenantID, campaignID)
	if err != nil {
		return nil, err
	}
	c.Steps = steps

	return &c, nil
}

func (s *Service) CreateCampaign(ctx context.Context, tenantID string, req CreateCampaignRequest) (*Campaign, error) {
	query := `
		INSERT INTO drip_campaigns (tenant_id, name, trigger_event, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, name, trigger_event, is_active, created_at, updated_at
	`

	var c Campaign
	err := s.db.QueryRow(ctx, query, tenantID, req.Name, req.TriggerEvent, req.IsActive).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.TriggerEvent, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	return &c, nil
}

func (s *Service) UpdateCampaign(ctx context.Context, tenantID, campaignID string, req UpdateCampaignRequest) (*Campaign, error) {
	// Build dynamic update query
	updates := []string{}
	args := []interface{}{tenantID, campaignID}
	argPos := 3

	if req.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argPos))
		args = append(args, req.Name)
		argPos++
	}
	if req.TriggerEvent != "" {
		updates = append(updates, fmt.Sprintf("trigger_event = $%d", argPos))
		args = append(args, req.TriggerEvent)
		argPos++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *req.IsActive)
		argPos++
	}

	if len(updates) == 0 {
		return s.GetCampaign(ctx, tenantID, campaignID)
	}

	query := fmt.Sprintf(`
		UPDATE drip_campaigns
		SET %s, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
		RETURNING id, tenant_id, name, trigger_event, is_active, created_at, updated_at
	`, joinStrings(updates, ", "))

	var c Campaign
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.TriggerEvent, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	return &c, nil
}

func (s *Service) DeleteCampaign(ctx context.Context, tenantID, campaignID string) error {
	query := `DELETE FROM drip_campaigns WHERE tenant_id = $1 AND id = $2`
	result, err := s.db.Exec(ctx, query, tenantID, campaignID)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}

	return nil
}

// Step CRUD

func (s *Service) ListSteps(ctx context.Context, tenantID, campaignID string) ([]Step, error) {
	// Verify campaign belongs to tenant
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM drip_campaigns WHERE id = $1 AND tenant_id = $2)", campaignID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("campaign not found")
	}

	query := `
		SELECT id, campaign_id, step_order, delay_days, action_type, subject, body, template_id, created_at, updated_at
		FROM drip_steps
		WHERE campaign_id = $1
		ORDER BY step_order ASC
	`

	rows, err := s.db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to list steps: %w", err)
	}
	defer rows.Close()

	steps := []Step{}
	for rows.Next() {
		var st Step
		err := rows.Scan(
			&st.ID, &st.CampaignID, &st.StepOrder, &st.DelayDays, &st.ActionType,
			&st.Subject, &st.Body, &st.TemplateID, &st.CreatedAt, &st.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}
		steps = append(steps, st)
	}

	return steps, nil
}

func (s *Service) CreateStep(ctx context.Context, tenantID, campaignID string, req CreateStepRequest) (*Step, error) {
	// Verify campaign belongs to tenant
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM drip_campaigns WHERE id = $1 AND tenant_id = $2)", campaignID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("campaign not found")
	}

	query := `
		INSERT INTO drip_steps (campaign_id, step_order, delay_days, action_type, subject, body, template_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, campaign_id, step_order, delay_days, action_type, subject, body, template_id, created_at, updated_at
	`

	var st Step
	err = s.db.QueryRow(ctx, query,
		campaignID, req.StepOrder, req.DelayDays, req.ActionType, req.Subject, req.Body, req.TemplateID,
	).Scan(
		&st.ID, &st.CampaignID, &st.StepOrder, &st.DelayDays, &st.ActionType,
		&st.Subject, &st.Body, &st.TemplateID, &st.CreatedAt, &st.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create step: %w", err)
	}

	return &st, nil
}

func (s *Service) UpdateStep(ctx context.Context, tenantID, campaignID, stepID string, req UpdateStepRequest) (*Step, error) {
	// Verify step belongs to campaign and tenant
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM drip_steps s
			JOIN drip_campaigns c ON c.id = s.campaign_id
			WHERE s.id = $1 AND s.campaign_id = $2 AND c.tenant_id = $3
		)
	`, stepID, campaignID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("step not found")
	}

	// Build dynamic update
	updates := []string{}
	args := []interface{}{stepID}
	argPos := 2

	if req.StepOrder != nil {
		updates = append(updates, fmt.Sprintf("step_order = $%d", argPos))
		args = append(args, *req.StepOrder)
		argPos++
	}
	if req.DelayDays != nil {
		updates = append(updates, fmt.Sprintf("delay_days = $%d", argPos))
		args = append(args, *req.DelayDays)
		argPos++
	}
	if req.ActionType != "" {
		updates = append(updates, fmt.Sprintf("action_type = $%d", argPos))
		args = append(args, req.ActionType)
		argPos++
	}
	if req.Subject != "" {
		updates = append(updates, fmt.Sprintf("subject = $%d", argPos))
		args = append(args, req.Subject)
		argPos++
	}
	if req.Body != "" {
		updates = append(updates, fmt.Sprintf("body = $%d", argPos))
		args = append(args, req.Body)
		argPos++
	}
	if req.TemplateID != nil {
		updates = append(updates, fmt.Sprintf("template_id = $%d", argPos))
		args = append(args, *req.TemplateID)
		argPos++
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE drip_steps
		SET %s, updated_at = NOW()
		WHERE id = $1
		RETURNING id, campaign_id, step_order, delay_days, action_type, subject, body, template_id, created_at, updated_at
	`, joinStrings(updates, ", "))

	var st Step
	err = s.db.QueryRow(ctx, query, args...).Scan(
		&st.ID, &st.CampaignID, &st.StepOrder, &st.DelayDays, &st.ActionType,
		&st.Subject, &st.Body, &st.TemplateID, &st.CreatedAt, &st.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update step: %w", err)
	}

	return &st, nil
}

func (s *Service) DeleteStep(ctx context.Context, tenantID, campaignID, stepID string) error {
	// Verify step belongs to campaign and tenant
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM drip_steps s
			JOIN drip_campaigns c ON c.id = s.campaign_id
			WHERE s.id = $1 AND s.campaign_id = $2 AND c.tenant_id = $3
		)
	`, stepID, campaignID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("step not found")
	}

	query := `DELETE FROM drip_steps WHERE id = $1`
	_, err = s.db.Exec(ctx, query, stepID)
	if err != nil {
		return fmt.Errorf("failed to delete step: %w", err)
	}

	return nil
}

// Enrollment

func (s *Service) EnrollPerson(ctx context.Context, tenantID, campaignID, personID string) (*Enrollment, error) {
	// Verify campaign is active and belongs to tenant
	var isActive bool
	err := s.db.QueryRow(ctx, "SELECT is_active FROM drip_campaigns WHERE id = $1 AND tenant_id = $2", campaignID, tenantID).Scan(&isActive)
	if err != nil {
		return nil, fmt.Errorf("campaign not found")
	}
	if !isActive {
		return nil, fmt.Errorf("campaign is not active")
	}

	// Check if already enrolled
	var existingID string
	err = s.db.QueryRow(ctx, "SELECT id FROM drip_enrollments WHERE campaign_id = $1 AND person_id = $2", campaignID, personID).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("person already enrolled in this campaign")
	}

	// Create enrollment
	query := `
		INSERT INTO drip_enrollments (campaign_id, person_id, status, enrolled_at)
		VALUES ($1, $2, 'active', NOW())
		RETURNING id, campaign_id, person_id, status, enrolled_at, completed_at
	`

	var e Enrollment
	err = s.db.QueryRow(ctx, query, campaignID, personID).Scan(
		&e.ID, &e.CampaignID, &e.PersonID, &e.Status, &e.EnrolledAt, &e.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll person: %w", err)
	}

	// Schedule all steps
	err = s.scheduleStepsForEnrollment(ctx, e.ID, campaignID, e.EnrolledAt)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule steps: %w", err)
	}

	return &e, nil
}

func (s *Service) scheduleStepsForEnrollment(ctx context.Context, enrollmentID, campaignID string, enrolledAt time.Time) error {
	// Get all steps for campaign
	steps, err := s.db.Query(ctx, `
		SELECT id, delay_days FROM drip_steps WHERE campaign_id = $1 ORDER BY step_order ASC
	`, campaignID)
	if err != nil {
		return err
	}
	defer steps.Close()

	for steps.Next() {
		var stepID string
		var delayDays int
		if err := steps.Scan(&stepID, &delayDays); err != nil {
			return err
		}

		scheduledAt := enrolledAt.AddDate(0, 0, delayDays)

		_, err := s.db.Exec(ctx, `
			INSERT INTO drip_step_executions (enrollment_id, step_id, status, scheduled_at)
			VALUES ($1, $2, 'pending', $3)
		`, enrollmentID, stepID, scheduledAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ListEnrollments(ctx context.Context, tenantID, campaignID string) ([]Enrollment, error) {
	// Verify campaign belongs to tenant
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM drip_campaigns WHERE id = $1 AND tenant_id = $2)", campaignID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("campaign not found")
	}

	query := `
		SELECT 
			e.id, e.campaign_id, e.person_id, e.status, e.enrolled_at, e.completed_at,
			p.first_name || ' ' || p.last_name as person_name,
			p.email as person_email
		FROM drip_enrollments e
		JOIN people p ON p.id = e.person_id
		WHERE e.campaign_id = $1
		ORDER BY e.enrolled_at DESC
	`

	rows, err := s.db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	enrollments := []Enrollment{}
	for rows.Next() {
		var e Enrollment
		err := rows.Scan(
			&e.ID, &e.CampaignID, &e.PersonID, &e.Status, &e.EnrolledAt, &e.CompletedAt,
			&e.PersonName, &e.PersonEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, nil
}

// ProcessPendingSteps processes all pending drip steps that are due
func (s *Service) ProcessPendingSteps(ctx context.Context) error {
	// Get all pending steps that are due
	query := `
		SELECT 
			ex.id, ex.enrollment_id, ex.step_id, ex.scheduled_at,
			st.action_type, st.subject, st.body,
			e.person_id,
			c.tenant_id
		FROM drip_step_executions ex
		JOIN drip_enrollments e ON e.id = ex.enrollment_id
		JOIN drip_steps st ON st.id = ex.step_id
		JOIN drip_campaigns c ON c.id = e.campaign_id
		WHERE ex.status = 'pending'
		AND ex.scheduled_at <= NOW()
		AND e.status = 'active'
		AND c.is_active = true
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query pending steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var execID, enrollmentID, stepID, actionType, subject, body, personID, tenantID string
		var scheduledAt time.Time
		err := rows.Scan(&execID, &enrollmentID, &stepID, &scheduledAt, &actionType, &subject, &body, &personID, &tenantID)
		if err != nil {
			fmt.Printf("Failed to scan step execution: %v\n", err)
			continue
		}

		// Execute the step (this would integrate with email/SMS service)
		err = s.executeStep(ctx, execID, actionType, subject, body, personID, tenantID)
		if err != nil {
			// Mark as failed
			_, _ = s.db.Exec(ctx, `
				UPDATE drip_step_executions 
				SET status = 'failed', error_message = $1, executed_at = NOW()
				WHERE id = $2
			`, err.Error(), execID)
			fmt.Printf("Failed to execute step %s: %v\n", execID, err)
			continue
		}

		// Mark as sent
		_, err = s.db.Exec(ctx, `
			UPDATE drip_step_executions 
			SET status = 'sent', executed_at = NOW()
			WHERE id = $1
		`, execID)
		if err != nil {
			fmt.Printf("Failed to update step execution status: %v\n", err)
		}

		// Check if all steps are complete for this enrollment
		s.checkEnrollmentCompletion(ctx, enrollmentID)
	}

	return nil
}

func (s *Service) executeStep(ctx context.Context, execID, actionType, subject, body, personID, tenantID string) error {
	// This is a placeholder - in a real implementation, this would:
	// - For email: send via SMTP/SendGrid/etc
	// - For SMS: send via Twilio/etc
	// - For follow_up: create a task/reminder

	fmt.Printf("Executing step %s: %s to person %s\n", execID, actionType, personID)

	// For now, just log it
	switch actionType {
	case "email":
		fmt.Printf("  Email: %s - %s\n", subject, body)
	case "sms":
		fmt.Printf("  SMS: %s\n", body)
	case "follow_up":
		fmt.Printf("  Follow-up reminder: %s\n", body)
	}

	return nil
}

func (s *Service) checkEnrollmentCompletion(ctx context.Context, enrollmentID string) {
	// Check if all steps are complete
	var pendingCount int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM drip_step_executions
		WHERE enrollment_id = $1 AND status = 'pending'
	`, enrollmentID).Scan(&pendingCount)
	if err != nil {
		return
	}

	if pendingCount == 0 {
		// Mark enrollment as completed
		_, _ = s.db.Exec(ctx, `
			UPDATE drip_enrollments
			SET status = 'completed', completed_at = NOW()
			WHERE id = $1
		`, enrollmentID)
	}
}
