package communication

import (
	"context"
	"fmt"
	"time"

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

// Helper to set tenant context
func (s *Service) setTenantContext(ctx context.Context, tenantID string) error {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// ===== MESSAGE TEMPLATES =====

func (s *Service) ListTemplates(ctx context.Context, tenantID, channel, category string) ([]MessageTemplate, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, name, subject, body, channel, category, variables, created_at, updated_at
		FROM message_templates
		WHERE 1=1`
	
	args := []interface{}{}
	argPos := 1

	if channel != "" {
		query += fmt.Sprintf(` AND channel = $%d`, argPos)
		args = append(args, channel)
		argPos++
	}

	if category != "" {
		query += fmt.Sprintf(` AND category = $%d`, argPos)
		args = append(args, category)
		argPos++
	}

	query += ` ORDER BY category, name`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	templates := []MessageTemplate{}
	for rows.Next() {
		var t MessageTemplate
		err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Subject, &t.Body, &t.Channel, &t.Category, &t.Variables, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, t)
	}

	return templates, nil
}

func (s *Service) CreateTemplate(ctx context.Context, tenantID string, template *MessageTemplate) (*MessageTemplate, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	template.ID = uuid.New().String()
	template.TenantID = tenantID

	query := `
		INSERT INTO message_templates (id, tenant_id, name, subject, body, channel, category, variables)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		template.ID, template.TenantID, template.Name, template.Subject, template.Body,
		template.Channel, template.Category, template.Variables,
	).Scan(&template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, tenantID, templateID string, template *MessageTemplate) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	query := `
		UPDATE message_templates
		SET name = $1, subject = $2, body = $3, channel = $4, category = $5, variables = $6
		WHERE id = $7`

	_, err := s.db.Exec(ctx, query,
		template.Name, template.Subject, template.Body, template.Channel, template.Category, template.Variables, templateID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

func (s *Service) DeleteTemplate(ctx context.Context, tenantID, templateID string) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	_, err := s.db.Exec(ctx, "DELETE FROM message_templates WHERE id = $1", templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// ===== CAMPAIGNS =====

func (s *Service) ListCampaigns(ctx context.Context, tenantID, status string) ([]Campaign, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, name, template_id, channel, subject, body, status,
		       scheduled_at, sent_at, recipient_count, opened_count, clicked_count,
		       target_type, target_id, created_at, updated_at
		FROM campaigns
		WHERE 1=1`

	args := []interface{}{}
	if status != "" {
		query += ` AND status = $1`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	campaigns := []Campaign{}
	for rows.Next() {
		var c Campaign
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.TemplateID, &c.Channel, &c.Subject, &c.Body,
			&c.Status, &c.ScheduledAt, &c.SentAt, &c.RecipientCount, &c.OpenedCount,
			&c.ClickedCount, &c.TargetType, &c.TargetID, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

func (s *Service) CreateCampaign(ctx context.Context, tenantID string, campaign *Campaign) (*Campaign, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	campaign.ID = uuid.New().String()
	campaign.TenantID = tenantID
	campaign.Status = "draft"

	query := `
		INSERT INTO campaigns (id, tenant_id, name, template_id, channel, subject, body, target_type, target_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		campaign.ID, campaign.TenantID, campaign.Name, campaign.TemplateID, campaign.Channel,
		campaign.Subject, campaign.Body, campaign.TargetType, campaign.TargetID,
	).Scan(&campaign.CreatedAt, &campaign.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	return campaign, nil
}

func (s *Service) GetCampaign(ctx context.Context, tenantID, campaignID string) (*Campaign, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, name, template_id, channel, subject, body, status,
		       scheduled_at, sent_at, recipient_count, opened_count, clicked_count,
		       target_type, target_id, created_at, updated_at
		FROM campaigns
		WHERE id = $1`

	var c Campaign
	err := s.db.QueryRow(ctx, query, campaignID).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.TemplateID, &c.Channel, &c.Subject, &c.Body,
		&c.Status, &c.ScheduledAt, &c.SentAt, &c.RecipientCount, &c.OpenedCount,
		&c.ClickedCount, &c.TargetType, &c.TargetID, &c.CreatedAt, &c.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("campaign not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	return &c, nil
}

func (s *Service) UpdateCampaign(ctx context.Context, tenantID, campaignID string, campaign *Campaign) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	// Only allow updating draft campaigns
	var status string
	err := s.db.QueryRow(ctx, "SELECT status FROM campaigns WHERE id = $1", campaignID).Scan(&status)
	if err != nil {
		return fmt.Errorf("failed to get campaign status: %w", err)
	}
	if status != "draft" {
		return fmt.Errorf("can only update draft campaigns")
	}

	query := `
		UPDATE campaigns
		SET name = $1, channel = $2, subject = $3, body = $4, target_type = $5, target_id = $6
		WHERE id = $7`

	_, err = s.db.Exec(ctx, query,
		campaign.Name, campaign.Channel, campaign.Subject, campaign.Body,
		campaign.TargetType, campaign.TargetID, campaignID,
	)

	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	return nil
}

func (s *Service) SendCampaign(ctx context.Context, tenantID, campaignID string, scheduledAt *time.Time) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	// TODO: Actually implement sending logic (email/SMS provider integration)
	// For now, just mark as sent or scheduled

	status := "sent"
	var sentAt *time.Time
	if scheduledAt != nil && scheduledAt.After(time.Now()) {
		status = "scheduled"
	} else {
		now := time.Now()
		sentAt = &now
	}

	query := `
		UPDATE campaigns
		SET status = $1, scheduled_at = $2, sent_at = $3
		WHERE id = $4`

	_, err := s.db.Exec(ctx, query, status, scheduledAt, sentAt, campaignID)
	if err != nil {
		return fmt.Errorf("failed to send campaign: %w", err)
	}

	return nil
}

func (s *Service) GetCampaignRecipients(ctx context.Context, tenantID, campaignID string) ([]CampaignRecipient, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, campaign_id, person_id, status, sent_at, opened_at, clicked_at
		FROM campaign_recipients
		WHERE campaign_id = $1
		ORDER BY sent_at DESC NULLS LAST`

	rows, err := s.db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipients: %w", err)
	}
	defer rows.Close()

	recipients := []CampaignRecipient{}
	for rows.Next() {
		var r CampaignRecipient
		err := rows.Scan(&r.ID, &r.CampaignID, &r.PersonID, &r.Status, &r.SentAt, &r.OpenedAt, &r.ClickedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipient: %w", err)
		}
		recipients = append(recipients, r)
	}

	return recipients, nil
}

// ===== JOURNEYS =====

func (s *Service) ListJourneys(ctx context.Context, tenantID string) ([]Journey, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT j.id, j.tenant_id, j.name, j.description, j.trigger_type, j.trigger_value, j.is_active, j.created_at, j.updated_at,
		       COUNT(DISTINCT je.id) as enrollment_count
		FROM journeys j
		LEFT JOIN journey_enrollments je ON j.id = je.journey_id AND je.status = 'active'
		GROUP BY j.id
		ORDER BY j.created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list journeys: %w", err)
	}
	defer rows.Close()

	journeys := []Journey{}
	for rows.Next() {
		var j Journey
		err := rows.Scan(&j.ID, &j.TenantID, &j.Name, &j.Description, &j.TriggerType, &j.TriggerValue, &j.IsActive, &j.CreatedAt, &j.UpdatedAt, &j.EnrollmentCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journey: %w", err)
		}
		journeys = append(journeys, j)
	}

	return journeys, nil
}

func (s *Service) CreateJourney(ctx context.Context, tenantID string, journey *Journey) (*Journey, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	journey.ID = uuid.New().String()
	journey.TenantID = tenantID

	query := `
		INSERT INTO journeys (id, tenant_id, name, description, trigger_type, trigger_value, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		journey.ID, journey.TenantID, journey.Name, journey.Description,
		journey.TriggerType, journey.TriggerValue, journey.IsActive,
	).Scan(&journey.CreatedAt, &journey.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create journey: %w", err)
	}

	return journey, nil
}

func (s *Service) GetJourney(ctx context.Context, tenantID, journeyID string) (*Journey, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, name, description, trigger_type, trigger_value, is_active, created_at, updated_at
		FROM journeys
		WHERE id = $1`

	var j Journey
	err := s.db.QueryRow(ctx, query, journeyID).Scan(
		&j.ID, &j.TenantID, &j.Name, &j.Description, &j.TriggerType, &j.TriggerValue, &j.IsActive, &j.CreatedAt, &j.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("journey not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get journey: %w", err)
	}

	// Get steps
	steps, err := s.GetJourneySteps(ctx, tenantID, journeyID)
	if err != nil {
		return nil, err
	}
	j.Steps = steps

	return &j, nil
}

func (s *Service) UpdateJourney(ctx context.Context, tenantID, journeyID string, journey *Journey) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	query := `
		UPDATE journeys
		SET name = $1, description = $2, trigger_type = $3, trigger_value = $4, is_active = $5
		WHERE id = $6`

	_, err := s.db.Exec(ctx, query,
		journey.Name, journey.Description, journey.TriggerType, journey.TriggerValue, journey.IsActive, journeyID,
	)

	if err != nil {
		return fmt.Errorf("failed to update journey: %w", err)
	}

	return nil
}

func (s *Service) DeleteJourney(ctx context.Context, tenantID, journeyID string) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	_, err := s.db.Exec(ctx, "DELETE FROM journeys WHERE id = $1", journeyID)
	if err != nil {
		return fmt.Errorf("failed to delete journey: %w", err)
	}

	return nil
}

// ===== JOURNEY STEPS =====

func (s *Service) GetJourneySteps(ctx context.Context, tenantID, journeyID string) ([]JourneyStep, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, journey_id, position, step_type, delay_days, delay_hours, template_id, config, created_at
		FROM journey_steps
		WHERE journey_id = $1
		ORDER BY position`

	rows, err := s.db.Query(ctx, query, journeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get journey steps: %w", err)
	}
	defer rows.Close()

	steps := []JourneyStep{}
	for rows.Next() {
		var s JourneyStep
		err := rows.Scan(&s.ID, &s.JourneyID, &s.Position, &s.StepType, &s.DelayDays, &s.DelayHours, &s.TemplateID, &s.Config, &s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journey step: %w", err)
		}
		steps = append(steps, s)
	}

	return steps, nil
}

func (s *Service) AddJourneyStep(ctx context.Context, tenantID, journeyID string, step *JourneyStep) (*JourneyStep, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	step.ID = uuid.New().String()
	step.JourneyID = journeyID

	query := `
		INSERT INTO journey_steps (id, journey_id, position, step_type, delay_days, delay_hours, template_id, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := s.db.QueryRow(ctx, query,
		step.ID, step.JourneyID, step.Position, step.StepType, step.DelayDays, step.DelayHours, step.TemplateID, step.Config,
	).Scan(&step.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add journey step: %w", err)
	}

	return step, nil
}

func (s *Service) UpdateJourneyStep(ctx context.Context, tenantID, stepID string, step *JourneyStep) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	query := `
		UPDATE journey_steps
		SET position = $1, step_type = $2, delay_days = $3, delay_hours = $4, template_id = $5, config = $6
		WHERE id = $7`

	_, err := s.db.Exec(ctx, query,
		step.Position, step.StepType, step.DelayDays, step.DelayHours, step.TemplateID, step.Config, stepID,
	)

	if err != nil {
		return fmt.Errorf("failed to update journey step: %w", err)
	}

	return nil
}

func (s *Service) DeleteJourneyStep(ctx context.Context, tenantID, stepID string) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	_, err := s.db.Exec(ctx, "DELETE FROM journey_steps WHERE id = $1", stepID)
	if err != nil {
		return fmt.Errorf("failed to delete journey step: %w", err)
	}

	return nil
}

// ===== JOURNEY ENROLLMENTS =====

func (s *Service) EnrollInJourney(ctx context.Context, tenantID, journeyID, personID string) (*JourneyEnrollment, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	enrollment := &JourneyEnrollment{
		ID:          uuid.New().String(),
		JourneyID:   journeyID,
		PersonID:    personID,
		CurrentStep: 0,
		Status:      "active",
		EnrolledAt:  time.Now(),
	}

	// Calculate next step time (first step)
	nextStepAt := time.Now()
	enrollment.NextStepAt = &nextStepAt

	query := `
		INSERT INTO journey_enrollments (id, journey_id, person_id, current_step, status, enrolled_at, next_step_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (journey_id, person_id) DO NOTHING
		RETURNING id`

	err := s.db.QueryRow(ctx, query,
		enrollment.ID, enrollment.JourneyID, enrollment.PersonID, enrollment.CurrentStep,
		enrollment.Status, enrollment.EnrolledAt, enrollment.NextStepAt,
	).Scan(&enrollment.ID)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("person already enrolled in this journey")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to enroll in journey: %w", err)
	}

	return enrollment, nil
}

func (s *Service) GetJourneyEnrollments(ctx context.Context, tenantID, journeyID string) ([]JourneyEnrollment, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT je.id, je.journey_id, je.person_id, je.current_step, je.status, je.enrolled_at, je.next_step_at, je.completed_at,
		       p.first_name || ' ' || p.last_name as person_name
		FROM journey_enrollments je
		JOIN people p ON je.person_id = p.id
		WHERE je.journey_id = $1
		ORDER BY je.enrolled_at DESC`

	rows, err := s.db.Query(ctx, query, journeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrollments: %w", err)
	}
	defer rows.Close()

	enrollments := []JourneyEnrollment{}
	for rows.Next() {
		var e JourneyEnrollment
		err := rows.Scan(&e.ID, &e.JourneyID, &e.PersonID, &e.CurrentStep, &e.Status, &e.EnrolledAt, &e.NextStepAt, &e.CompletedAt, &e.PersonName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, nil
}

// ===== CONNECTION CARDS =====

func (s *Service) SubmitConnectionCard(ctx context.Context, tenantID string, card *ConnectionCard) (*ConnectionCard, error) {
	// Note: This is a PUBLIC endpoint, so we manually set tenant context
	// The tenantID will come from the subdomain or API key in the request
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	card.ID = uuid.New().String()
	card.TenantID = tenantID
	card.SubmittedAt = time.Now()

	query := `
		INSERT INTO connection_cards (id, tenant_id, first_name, last_name, email, phone, is_first_visit, how_heard, prayer_request, interested_in)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING submitted_at`

	err := s.db.QueryRow(ctx, query,
		card.ID, card.TenantID, card.FirstName, card.LastName, card.Email, card.Phone,
		card.IsFirstVisit, card.HowHeard, card.PrayerRequest, card.InterestedIn,
	).Scan(&card.SubmittedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to submit connection card: %w", err)
	}

	return card, nil
}

func (s *Service) ListConnectionCards(ctx context.Context, tenantID string, processedOnly bool) ([]ConnectionCard, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, first_name, last_name, email, phone, is_first_visit, how_heard, prayer_request, interested_in, submitted_at, processed, person_id
		FROM connection_cards
		WHERE 1=1`

	args := []interface{}{}
	if processedOnly {
		query += ` AND processed = $1`
		args = append(args, false)
	}

	query += ` ORDER BY submitted_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list connection cards: %w", err)
	}
	defer rows.Close()

	cards := []ConnectionCard{}
	for rows.Next() {
		var c ConnectionCard
		err := rows.Scan(&c.ID, &c.TenantID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.IsFirstVisit, &c.HowHeard, &c.PrayerRequest, &c.InterestedIn, &c.SubmittedAt, &c.Processed, &c.PersonID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection card: %w", err)
		}
		cards = append(cards, c)
	}

	return cards, nil
}

func (s *Service) GetConnectionCard(ctx context.Context, tenantID, cardID string) (*ConnectionCard, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, first_name, last_name, email, phone, is_first_visit, how_heard, prayer_request, interested_in, submitted_at, processed, person_id
		FROM connection_cards
		WHERE id = $1`

	var c ConnectionCard
	err := s.db.QueryRow(ctx, query, cardID).Scan(
		&c.ID, &c.TenantID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
		&c.IsFirstVisit, &c.HowHeard, &c.PrayerRequest, &c.InterestedIn,
		&c.SubmittedAt, &c.Processed, &c.PersonID,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("connection card not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get connection card: %w", err)
	}

	return &c, nil
}

func (s *Service) ProcessConnectionCard(ctx context.Context, tenantID, cardID, personID string) error {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return err
	}

	query := `
		UPDATE connection_cards
		SET processed = TRUE, person_id = $1
		WHERE id = $2`

	_, err := s.db.Exec(ctx, query, personID, cardID)
	if err != nil {
		return fmt.Errorf("failed to process connection card: %w", err)
	}

	return nil
}

// ===== STATS =====

func (s *Service) GetStats(ctx context.Context, tenantID string) (*CommunicationStats, error) {
	if err := s.setTenantContext(ctx, tenantID); err != nil {
		return nil, err
	}

	stats := &CommunicationStats{}

	// Emails sent this month
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM campaign_recipients cr
		JOIN campaigns c ON cr.campaign_id = c.id
		WHERE c.channel = 'email'
		AND cr.status IN ('sent', 'delivered', 'opened', 'clicked')
		AND DATE_TRUNC('month', cr.sent_at) = DATE_TRUNC('month', CURRENT_DATE)
	`).Scan(&stats.EmailsSentThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get email count: %w", err)
	}

	// SMS sent this month
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM campaign_recipients cr
		JOIN campaigns c ON cr.campaign_id = c.id
		WHERE c.channel = 'sms'
		AND cr.status IN ('sent', 'delivered')
		AND DATE_TRUNC('month', cr.sent_at) = DATE_TRUNC('month', CURRENT_DATE)
	`).Scan(&stats.SMSSentThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMS count: %w", err)
	}

	// Open rate (all time)
	var totalSent, totalOpened int
	err = s.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE cr.status IN ('sent', 'delivered', 'opened', 'clicked')),
			COUNT(*) FILTER (WHERE cr.status IN ('opened', 'clicked'))
		FROM campaign_recipients cr
		JOIN campaigns c ON cr.campaign_id = c.id
		WHERE c.channel = 'email'
	`).Scan(&totalSent, &totalOpened)
	if err != nil {
		return nil, fmt.Errorf("failed to get open rate: %w", err)
	}

	if totalSent > 0 {
		stats.OpenRate = float64(totalOpened) / float64(totalSent) * 100
	}

	// Active journeys
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM journeys
		WHERE is_active = TRUE
	`).Scan(&stats.ActiveJourneys)
	if err != nil {
		return nil, fmt.Errorf("failed to get active journeys: %w", err)
	}

	// Unprocessed cards
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM connection_cards
		WHERE processed = FALSE
	`).Scan(&stats.UnprocessedCards)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed cards: %w", err)
	}

	return stats, nil
}
