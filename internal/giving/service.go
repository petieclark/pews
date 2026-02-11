package giving

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

// GetDB returns the database pool for direct queries
func (s *Service) GetDB() *pgxpool.Pool {
	return s.db
}

// Funds

func (s *Service) ListFunds(ctx context.Context, tenantID string) ([]Fund, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at 
		 FROM funds WHERE tenant_id = $1 ORDER BY is_default DESC, name ASC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var funds []Fund
	for rows.Next() {
		var f Fund
		if err := rows.Scan(&f.ID, &f.TenantID, &f.Name, &f.Description, &f.IsDefault, &f.IsActive, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		funds = append(funds, f)
	}

	return funds, rows.Err()
}

func (s *Service) CreateFund(ctx context.Context, tenantID, name, description string, isDefault bool) (*Fund, error) {
	fund := &Fund{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		IsDefault:   isDefault,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// If this is default, unset any other default
	if isDefault {
		_, err := s.db.Exec(ctx, `UPDATE funds SET is_default = FALSE WHERE tenant_id = $1`, tenantID)
		if err != nil {
			return nil, err
		}
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO funds (id, tenant_id, name, description, is_default, is_active) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		fund.ID, fund.TenantID, fund.Name, fund.Description, fund.IsDefault, fund.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fund: %w", err)
	}

	return fund, nil
}

func (s *Service) UpdateFund(ctx context.Context, tenantID, fundID, name, description string, isDefault, isActive bool) (*Fund, error) {
	// If setting as default, unset others
	if isDefault {
		_, err := s.db.Exec(ctx, `UPDATE funds SET is_default = FALSE WHERE tenant_id = $1 AND id != $2`, tenantID, fundID)
		if err != nil {
			return nil, err
		}
	}

	_, err := s.db.Exec(ctx,
		`UPDATE funds SET name = $1, description = $2, is_default = $3, is_active = $4, updated_at = NOW()
		 WHERE id = $5 AND tenant_id = $6`,
		name, description, isDefault, isActive, fundID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update fund: %w", err)
	}

	return s.GetFund(ctx, tenantID, fundID)
}

func (s *Service) GetFund(ctx context.Context, tenantID, fundID string) (*Fund, error) {
	var f Fund
	err := s.db.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at 
		 FROM funds WHERE id = $1 AND tenant_id = $2`,
		fundID, tenantID,
	).Scan(&f.ID, &f.TenantID, &f.Name, &f.Description, &f.IsDefault, &f.IsActive, &f.CreatedAt, &f.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("fund not found")
	}
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// Donations

func (s *Service) ListDonations(ctx context.Context, tenantID, personID, fundID, fromDate, toDate string, limit, offset int) ([]Donation, int, error) {
	query := `
		SELECT d.id, d.tenant_id, d.person_id, d.fund_id, d.amount_cents, d.currency,
		       d.payment_method, d.stripe_payment_intent_id, d.stripe_charge_id, d.status,
		       d.is_recurring, d.recurring_frequency, d.stripe_subscription_id, d.memo,
		       d.donated_at, d.created_at, d.updated_at,
		       COALESCE(p.first_name || ' ' || p.last_name, 'Anonymous') as person_name,
		       f.name as fund_name
		FROM donations d
		LEFT JOIN people p ON d.person_id = p.id
		LEFT JOIN funds f ON d.fund_id = f.id
		WHERE d.tenant_id = $1
	`
	args := []interface{}{tenantID}
	argNum := 2

	if personID != "" {
		query += fmt.Sprintf(" AND d.person_id = $%d", argNum)
		args = append(args, personID)
		argNum++
	}

	if fundID != "" {
		query += fmt.Sprintf(" AND d.fund_id = $%d", argNum)
		args = append(args, fundID)
		argNum++
	}

	if fromDate != "" {
		query += fmt.Sprintf(" AND d.donated_at >= $%d", argNum)
		args = append(args, fromDate)
		argNum++
	}

	if toDate != "" {
		query += fmt.Sprintf(" AND d.donated_at <= $%d", argNum)
		args = append(args, toDate)
		argNum++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS count_query"
	var total int
	if err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += " ORDER BY d.donated_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var donations []Donation
	for rows.Next() {
		var d Donation
		var personName string
		var fundName string

		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.PersonID, &d.FundID, &d.AmountCents, &d.Currency,
			&d.PaymentMethod, &d.StripePaymentIntentID, &d.StripeChargeID, &d.Status,
			&d.IsRecurring, &d.RecurringFrequency, &d.StripeSubscriptionID, &d.Memo,
			&d.DonatedAt, &d.CreatedAt, &d.UpdatedAt, &personName, &fundName,
		); err != nil {
			return nil, 0, err
		}

		d.PersonName = &personName
		d.FundName = fundName
		d.AmountDisplay = formatCents(d.AmountCents)
		donations = append(donations, d)
	}

	return donations, total, rows.Err()
}

func (s *Service) GetDonation(ctx context.Context, tenantID, donationID string) (*Donation, error) {
	var d Donation
	var personName string
	var fundName string

	err := s.db.QueryRow(ctx,
		`SELECT d.id, d.tenant_id, d.person_id, d.fund_id, d.amount_cents, d.currency,
		        d.payment_method, d.stripe_payment_intent_id, d.stripe_charge_id, d.status,
		        d.is_recurring, d.recurring_frequency, d.stripe_subscription_id, d.memo,
		        d.donated_at, d.created_at, d.updated_at,
		        COALESCE(p.first_name || ' ' || p.last_name, 'Anonymous') as person_name,
		        f.name as fund_name
		 FROM donations d
		 LEFT JOIN people p ON d.person_id = p.id
		 LEFT JOIN funds f ON d.fund_id = f.id
		 WHERE d.id = $1 AND d.tenant_id = $2`,
		donationID, tenantID,
	).Scan(
		&d.ID, &d.TenantID, &d.PersonID, &d.FundID, &d.AmountCents, &d.Currency,
		&d.PaymentMethod, &d.StripePaymentIntentID, &d.StripeChargeID, &d.Status,
		&d.IsRecurring, &d.RecurringFrequency, &d.StripeSubscriptionID, &d.Memo,
		&d.DonatedAt, &d.CreatedAt, &d.UpdatedAt, &personName, &fundName,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("donation not found")
	}
	if err != nil {
		return nil, err
	}

	d.PersonName = &personName
	d.FundName = fundName
	d.AmountDisplay = formatCents(d.AmountCents)

	return &d, nil
}

func (s *Service) CreateDonation(ctx context.Context, tenantID string, personID *string, fundID string, amountCents int, paymentMethod, memo string, donatedAt time.Time) (*Donation, error) {
	donation := &Donation{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		PersonID:      personID,
		FundID:        fundID,
		AmountCents:   amountCents,
		Currency:      "USD",
		Status:        "completed",
		IsRecurring:   false,
		DonatedAt:     donatedAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if paymentMethod != "" {
		donation.PaymentMethod = &paymentMethod
	}
	if memo != "" {
		donation.Memo = &memo
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO donations (id, tenant_id, person_id, fund_id, amount_cents, currency, 
		                        payment_method, status, memo, donated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		donation.ID, donation.TenantID, donation.PersonID, donation.FundID, donation.AmountCents,
		donation.Currency, donation.PaymentMethod, donation.Status, donation.Memo, donation.DonatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create donation: %w", err)
	}

	return s.GetDonation(ctx, tenantID, donation.ID)
}

func (s *Service) GetPersonGivingHistory(ctx context.Context, tenantID, personID string) ([]Donation, int, error) {
	donations, _, err := s.ListDonations(ctx, tenantID, personID, "", "", "", 100, 0)
	if err != nil {
		return nil, 0, err
	}

	// Calculate total
	totalCents := 0
	for _, d := range donations {
		if d.Status == "completed" {
			totalCents += d.AmountCents
		}
	}

	return donations, totalCents, nil
}

// Giving Stats

func (s *Service) GetGivingStats(ctx context.Context, tenantID string) (*GivingStats, error) {
	stats := &GivingStats{}
	now := time.Now()

	// Total this month
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0), COUNT(*) 
		 FROM donations 
		 WHERE tenant_id = $1 AND status = 'completed' AND donated_at >= $2`,
		tenantID, monthStart,
	).Scan(&stats.TotalThisMonth, &stats.DonationCount)
	if err != nil {
		return nil, err
	}

	// Total this year
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	err = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0) FROM donations 
		 WHERE tenant_id = $1 AND status = 'completed' AND donated_at >= $2`,
		tenantID, yearStart,
	).Scan(&stats.TotalThisYear)
	if err != nil {
		return nil, err
	}

	// Total all time
	var count int
	err = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0), COUNT(*) FROM donations 
		 WHERE tenant_id = $1 AND status = 'completed'`,
		tenantID,
	).Scan(&stats.TotalAllTime, &count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		stats.AverageDonation = stats.TotalAllTime / count
	}

	// Fund breakdown
	rows, err := s.db.Query(ctx,
		`SELECT f.id, f.name, COALESCE(SUM(d.amount_cents), 0) as total, COUNT(DISTINCT d.person_id) as donors
		 FROM funds f
		 LEFT JOIN donations d ON f.id = d.fund_id AND d.status = 'completed' AND d.tenant_id = $1
		 WHERE f.tenant_id = $1 AND f.is_active = TRUE
		 GROUP BY f.id, f.name
		 ORDER BY total DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fs FundSummary
		if err := rows.Scan(&fs.FundID, &fs.FundName, &fs.TotalCents, &fs.DonorCount); err != nil {
			return nil, err
		}
		if stats.TotalAllTime > 0 {
			fs.Percentage = float64(fs.TotalCents) / float64(stats.TotalAllTime) * 100
		}
		stats.FundBreakdown = append(stats.FundBreakdown, fs)
	}

	// Monthly trend (last 12 months)
	monthlyRows, err := s.db.Query(ctx,
		`SELECT 
		   TO_CHAR(donated_at, 'YYYY-MM') as month,
		   COALESCE(SUM(amount_cents), 0) as total,
		   COUNT(*) as count
		 FROM donations
		 WHERE tenant_id = $1 AND status = 'completed' 
		   AND donated_at >= $2
		 GROUP BY TO_CHAR(donated_at, 'YYYY-MM')
		 ORDER BY month DESC`,
		tenantID, now.AddDate(0, -12, 0),
	)
	if err != nil {
		return nil, err
	}
	defer monthlyRows.Close()

	for monthlyRows.Next() {
		var mt MonthlyTotal
		if err := monthlyRows.Scan(&mt.Month, &mt.TotalCents, &mt.Count); err != nil {
			return nil, err
		}
		stats.MonthlyTrend = append(stats.MonthlyTrend, mt)
	}

	return stats, nil
}

// Giving Statements

func (s *Service) GenerateGivingStatement(ctx context.Context, tenantID, personID string, year int) (*GivingStatement, error) {
	// Calculate total for the year
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	var totalCents int
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0) FROM donations
		 WHERE tenant_id = $1 AND person_id = $2 AND status = 'completed'
		   AND donated_at >= $3 AND donated_at < $4`,
		tenantID, personID, yearStart, yearEnd,
	).Scan(&totalCents)
	if err != nil {
		return nil, err
	}

	// Create or update statement
	stmt := &GivingStatement{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		PersonID:    personID,
		Year:        year,
		TotalCents:  totalCents,
		GeneratedAt: time.Now(),
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO giving_statements (id, tenant_id, person_id, year, total_cents)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (tenant_id, person_id, year) 
		 DO UPDATE SET total_cents = $5, generated_at = NOW()`,
		stmt.ID, stmt.TenantID, stmt.PersonID, stmt.Year, stmt.TotalCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create giving statement: %w", err)
	}

	return stmt, nil
}

func (s *Service) ListRecurringDonations(ctx context.Context, tenantID string) ([]Donation, error) {
	rows, err := s.db.Query(ctx,
		`SELECT d.id, d.tenant_id, d.person_id, d.fund_id, d.amount_cents, d.currency,
		        d.payment_method, d.stripe_subscription_id, d.recurring_frequency, d.status,
		        d.created_at, d.updated_at,
		        COALESCE(p.first_name || ' ' || p.last_name, 'Anonymous') as person_name,
		        f.name as fund_name
		 FROM donations d
		 LEFT JOIN people p ON d.person_id = p.id
		 LEFT JOIN funds f ON d.fund_id = f.id
		 WHERE d.tenant_id = $1 AND d.is_recurring = TRUE
		 ORDER BY d.created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donations []Donation
	for rows.Next() {
		var d Donation
		var personName, fundName string

		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.PersonID, &d.FundID, &d.AmountCents, &d.Currency,
			&d.PaymentMethod, &d.StripeSubscriptionID, &d.RecurringFrequency, &d.Status,
			&d.CreatedAt, &d.UpdatedAt, &personName, &fundName,
		); err != nil {
			return nil, err
		}

		d.PersonName = &personName
		d.FundName = fundName
		d.IsRecurring = true
		d.AmountDisplay = formatCents(d.AmountCents)
		donations = append(donations, d)
	}

	return donations, rows.Err()
}

// Kiosk Config

func (s *Service) GetKioskConfig(ctx context.Context, tenantID string) (*KioskConfig, error) {
	var config KioskConfig
	err := s.db.QueryRow(ctx,
		`SELECT kiosk_config FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&config)
	
	if err != nil {
		return nil, err
	}
	
	return &config, nil
}

func (s *Service) UpdateKioskConfig(ctx context.Context, tenantID string, config *KioskConfig) error {
	_, err := s.db.Exec(ctx,
		`UPDATE tenants SET kiosk_config = $1, updated_at = NOW() WHERE id = $2`,
		config,
		tenantID,
	)
	return err
}

// Helper functions

func formatCents(cents int) string {
	dollars := float64(cents) / 100.0
	return fmt.Sprintf("$%.2f", dollars)
}
