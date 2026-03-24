package giving

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/accountlink"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

type StripeService struct {
	db          *pgxpool.Pool
	stripeKey   string
	frontendURL string
}

func NewStripeService(db *pgxpool.Pool, stripeKey, frontendURL string) *StripeService {
	stripe.Key = stripeKey
	return &StripeService{
		db:          db,
		stripeKey:   stripeKey,
		frontendURL: frontendURL,
	}
}

// --- Stripe Connect ---

func (s *StripeService) GetTenantName(ctx context.Context, tenantID string) (string, error) {
	var name string
	err := s.db.QueryRow(ctx, `SELECT name FROM tenants WHERE id = $1`, tenantID).Scan(&name)
	return name, err
}

func (s *StripeService) CreateConnectOnboardingLink(ctx context.Context, tenantID, tenantName, tenantEmail string) (string, error) {
	var stripeAccountID string
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(stripe_account_id, '') FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID)
	if err != nil {
		return "", err
	}

	if stripeAccountID == "" {
		params := &stripe.AccountParams{
			Type:    stripe.String(string(stripe.AccountTypeExpress)),
			Country: stripe.String("US"),
			Email:   stripe.String(tenantEmail),
			Capabilities: &stripe.AccountCapabilitiesParams{
				CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
					Requested: stripe.Bool(true),
				},
				Transfers: &stripe.AccountCapabilitiesTransfersParams{
					Requested: stripe.Bool(true),
				},
			},
			BusinessType: stripe.String(string(stripe.AccountBusinessTypeNonProfit)),
			BusinessProfile: &stripe.AccountBusinessProfileParams{
				Name: stripe.String(tenantName),
			},
			Metadata: map[string]string{
				"tenant_id": tenantID,
			},
		}

		acc, err := account.New(params)
		if err != nil {
			return "", fmt.Errorf("failed to create Stripe account: %w", err)
		}

		_, err = s.db.Exec(ctx,
			`UPDATE tenants SET stripe_account_id = $1 WHERE id = $2`,
			acc.ID, tenantID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to save Stripe account ID: %w", err)
		}

		stripeAccountID = acc.ID
	}

	returnURL := s.frontendURL + "/dashboard/giving/settings?setup=complete"
	refreshURL := s.frontendURL + "/dashboard/giving/settings?setup=refresh"

	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(stripeAccountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(linkParams)
	if err != nil {
		return "", fmt.Errorf("failed to create account link: %w", err)
	}

	return link.URL, nil
}

func (s *StripeService) RefreshOnboardingLink(ctx context.Context, tenantID string) (string, error) {
	var stripeAccountID string
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(stripe_account_id, '') FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID)
	if err != nil || stripeAccountID == "" {
		return "", fmt.Errorf("no Stripe account found for this tenant")
	}

	returnURL := s.frontendURL + "/dashboard/giving/settings?setup=complete"
	refreshURL := s.frontendURL + "/dashboard/giving/settings?setup=refresh"

	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(stripeAccountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(linkParams)
	if err != nil {
		return "", fmt.Errorf("failed to create account link: %w", err)
	}

	return link.URL, nil
}

func (s *StripeService) GetConnectStatus(ctx context.Context, tenantID string) (*StripeConnectStatus, error) {
	var stripeAccountID *string
	var onboardingCompleted bool
	var testMode bool

	err := s.db.QueryRow(ctx,
		`SELECT stripe_account_id, stripe_onboarding_completed, COALESCE(stripe_test_mode, TRUE) FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID, &onboardingCompleted, &testMode)
	if err != nil {
		return nil, err
	}

	status := &StripeConnectStatus{
		Connected:           stripeAccountID != nil && *stripeAccountID != "",
		OnboardingCompleted: onboardingCompleted,
		IsTestMode:          testMode,
	}

	if status.Connected {
		status.AccountID = *stripeAccountID

		acc, err := account.GetByID(*stripeAccountID, nil)
		if err == nil {
			status.ChargesEnabled = acc.ChargesEnabled
			status.PayoutsEnabled = acc.PayoutsEnabled

			if acc.DetailsSubmitted && !onboardingCompleted {
				_, _ = s.db.Exec(ctx,
					`UPDATE tenants SET stripe_onboarding_completed = TRUE WHERE id = $1`,
					tenantID,
				)
				status.OnboardingCompleted = true
			}
		}
	}

	return status, nil
}

func (s *StripeService) SetTestMode(ctx context.Context, tenantID string, testMode bool) error {
	_, err := s.db.Exec(ctx,
		`UPDATE tenants SET stripe_test_mode = $1 WHERE id = $2`,
		testMode, tenantID,
	)
	return err
}

// --- Checkout Sessions ---

func (s *StripeService) CreateCheckoutSession(ctx context.Context, tenantID, personID, fundID string, amountCents int) (string, error) {
	var stripeAccountID *string
	err := s.db.QueryRow(ctx,
		`SELECT stripe_account_id FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID)
	if err != nil {
		return "", err
	}

	if stripeAccountID == nil || *stripeAccountID == "" {
		return "", fmt.Errorf("church has not completed Stripe Connect onboarding")
	}

	var fundName string
	err = s.db.QueryRow(ctx, `SELECT name FROM funds WHERE id = $1`, fundID).Scan(&fundName)
	if err != nil {
		return "", fmt.Errorf("fund not found")
	}

	appFeeCents := calculateAppFee(amountCents)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Donation to " + fundName),
						Description: stripe.String("Online giving"),
					},
					UnitAmount: stripe.Int64(int64(amountCents)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			ApplicationFeeAmount: stripe.Int64(int64(appFeeCents)),
		},
		SuccessURL: stripe.String(s.frontendURL + "/dashboard/giving?donation=success"),
		CancelURL:  stripe.String(s.frontendURL + "/dashboard/giving?donation=canceled"),
		Metadata: map[string]string{
			"tenant_id": tenantID,
			"person_id": personID,
			"fund_id":   fundID,
		},
	}
	params.SetStripeAccount(*stripeAccountID)

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// CreatePublicGivingSession creates a Stripe Checkout Session for the public giving page
func (s *StripeService) CreatePublicGivingSession(ctx context.Context, req PublicGiveRequest) (string, error) {
	// Get tenant info by slug
	var tenantID, tenantName string
	var stripeAccountID *string
	err := s.db.QueryRow(ctx,
		`SELECT id, name, stripe_account_id FROM tenants WHERE slug = $1`,
		req.TenantSlug,
	).Scan(&tenantID, &tenantName, &stripeAccountID)
	if err != nil {
		return "", fmt.Errorf("church not found")
	}

	if stripeAccountID == nil || *stripeAccountID == "" {
		return "", fmt.Errorf("this church has not set up online giving")
	}

	// Get fund name
	var fundName string
	err = s.db.QueryRow(ctx, `SELECT name FROM funds WHERE id = $1 AND tenant_id = $2`, req.FundID, tenantID).Scan(&fundName)
	if err != nil {
		return "", fmt.Errorf("fund not found")
	}

	appFeeCents := calculateAppFee(req.AmountCents)

	metadata := map[string]string{
		"tenant_id":  tenantID,
		"fund_id":    req.FundID,
		"source":     "public_giving",
	}
	if req.DonorName != "" {
		metadata["donor_name"] = req.DonorName
	}
	if req.DonorEmail != "" {
		metadata["donor_email"] = req.DonorEmail
	}

	successURL := s.frontendURL + "/give/" + req.TenantSlug + "/thank-you"
	cancelURL := s.frontendURL + "/give/" + req.TenantSlug

	if req.Frequency == "monthly" {
		// Recurring: use subscription mode
		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String("Monthly Donation to " + fundName),
							Description: stripe.String("Recurring monthly giving to " + tenantName),
						},
						UnitAmount: stripe.Int64(int64(req.AmountCents)),
						Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
							Interval: stripe.String("month"),
						},
					},
					Quantity: stripe.Int64(1),
				},
			},
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				ApplicationFeePercent: stripe.Float64(0.5), // 0.5% platform fee
				Metadata:              metadata,
			},
			SuccessURL: stripe.String(successURL),
			CancelURL:  stripe.String(cancelURL),
			Metadata:   metadata,
		}
		if req.DonorEmail != "" {
			params.CustomerEmail = stripe.String(req.DonorEmail)
		}
		params.SetStripeAccount(*stripeAccountID)

		sess, err := session.New(params)
		if err != nil {
			return "", fmt.Errorf("failed to create checkout session: %w", err)
		}
		return sess.URL, nil
	}

	// One-time payment
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Donation to " + fundName),
						Description: stripe.String("Online giving to " + tenantName),
					},
					UnitAmount: stripe.Int64(int64(req.AmountCents)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			ApplicationFeeAmount: stripe.Int64(int64(appFeeCents)),
			Metadata:             metadata,
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata:   metadata,
	}
	if req.DonorEmail != "" {
		params.CustomerEmail = stripe.String(req.DonorEmail)
	}
	params.SetStripeAccount(*stripeAccountID)

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// CreatePublicCheckoutSession creates a kiosk checkout session
func (s *StripeService) CreatePublicCheckoutSession(ctx context.Context, tenantID, fundID string, amountCents int, name, email *string) (string, error) {
	var stripeAccountID *string
	err := s.db.QueryRow(ctx,
		`SELECT stripe_account_id FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID)
	if err != nil {
		return "", err
	}

	if stripeAccountID == nil || *stripeAccountID == "" {
		return "", fmt.Errorf("church has not completed Stripe Connect onboarding")
	}

	var fundName string
	err = s.db.QueryRow(ctx, `SELECT name FROM funds WHERE id = $1 AND tenant_id = $2`, fundID, tenantID).Scan(&fundName)
	if err != nil {
		return "", fmt.Errorf("fund not found")
	}

	appFeeCents := calculateAppFee(amountCents)

	metadata := map[string]string{
		"tenant_id": tenantID,
		"fund_id":   fundID,
		"source":    "kiosk",
	}
	if name != nil && *name != "" {
		metadata["donor_name"] = *name
	}
	if email != nil && *email != "" {
		metadata["donor_email"] = *email
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Donation to " + fundName),
						Description: stripe.String("Kiosk donation"),
					},
					UnitAmount: stripe.Int64(int64(amountCents)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			ApplicationFeeAmount: stripe.Int64(int64(appFeeCents)),
			Metadata:             metadata,
		},
		SuccessURL: stripe.String(s.frontendURL + "/giving-kiosk/thank-you"),
		CancelURL:  stripe.String(s.frontendURL + "/giving-kiosk"),
		Metadata:   metadata,
	}

	if email != nil && *email != "" {
		params.CustomerEmail = stripe.String(*email)
	}

	params.SetStripeAccount(*stripeAccountID)

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// --- Webhook Handlers ---

func (s *StripeService) HandleCheckoutSessionCompleted(ctx context.Context, sess *stripe.CheckoutSession) error {
	tenantID := sess.Metadata["tenant_id"]
	fundID := sess.Metadata["fund_id"]
	source := sess.Metadata["source"]
	donorName := sess.Metadata["donor_name"]
	donorEmail := sess.Metadata["donor_email"]

	if tenantID == "" || fundID == "" {
		log.Printf("Skipping checkout.session.completed: missing tenant_id or fund_id in metadata")
		return nil
	}

	// Check for duplicate
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM donations WHERE stripe_checkout_session_id = $1)`,
		sess.ID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	amountCents := int(sess.AmountTotal)
	isRecurring := sess.Mode == stripe.CheckoutSessionModeSubscription
	var recurringFreq *string
	var stripeSubID *string
	if isRecurring && sess.Subscription != nil {
		freq := "monthly"
		recurringFreq = &freq
		stripeSubID = &sess.Subscription.ID
	}

	paymentMethod := "card"
	status := "completed"
	if source == "" {
		source = "online"
	}

	donation := &Donation{
		ID:                      uuid.New().String(),
		TenantID:                tenantID,
		FundID:                  fundID,
		AmountCents:             amountCents,
		Currency:                "USD",
		PaymentMethod:           &paymentMethod,
		StripeCheckoutSessionID: &sess.ID,
		Status:                  status,
		IsRecurring:             isRecurring,
		RecurringFrequency:      recurringFreq,
		StripeSubscriptionID:    stripeSubID,
		DonatedAt:               time.Now(),
		DonorName:               nilIfEmpty(donorName),
		DonorEmail:              nilIfEmpty(donorEmail),
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO donations (id, tenant_id, fund_id, amount_cents, currency,
		                        payment_method, stripe_checkout_session_id,
		                        status, is_recurring, recurring_frequency, stripe_subscription_id,
		                        donated_at, donor_name, donor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		donation.ID, donation.TenantID, donation.FundID,
		donation.AmountCents, donation.Currency, donation.PaymentMethod,
		donation.StripeCheckoutSessionID,
		donation.Status, donation.IsRecurring, donation.RecurringFrequency,
		donation.StripeSubscriptionID, donation.DonatedAt,
		donation.DonorName, donation.DonorEmail,
	)
	if err != nil {
		return fmt.Errorf("failed to record donation: %w", err)
	}

	log.Printf("Recorded donation %s: $%.2f to tenant %s (source: %s)", donation.ID, float64(amountCents)/100, tenantID, source)
	return nil
}

func (s *StripeService) HandlePaymentIntentSucceeded(ctx context.Context, pi *stripe.PaymentIntent) error {
	tenantID := pi.Metadata["tenant_id"]
	fundID := pi.Metadata["fund_id"]

	if tenantID == "" || fundID == "" {
		return nil
	}

	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM donations WHERE stripe_payment_intent_id = $1)`,
		pi.ID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	personID := pi.Metadata["person_id"]
	donorName := pi.Metadata["donor_name"]
	donorEmail := pi.Metadata["donor_email"]

	var personIDPtr *string
	if personID != "" {
		personIDPtr = &personID
	}

	paymentMethod := "card"

	donation := &Donation{
		ID:                    uuid.New().String(),
		TenantID:              tenantID,
		PersonID:              personIDPtr,
		FundID:                fundID,
		AmountCents:           int(pi.Amount),
		Currency:              string(pi.Currency),
		PaymentMethod:         &paymentMethod,
		StripePaymentIntentID: &pi.ID,
		Status:                "completed",
		IsRecurring:           false,
		DonatedAt:             time.Now(),
		DonorName:             nilIfEmpty(donorName),
		DonorEmail:            nilIfEmpty(donorEmail),
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO donations (id, tenant_id, person_id, fund_id, amount_cents, currency,
		                        payment_method, stripe_payment_intent_id,
		                        status, is_recurring, donated_at, donor_name, donor_email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		donation.ID, donation.TenantID, donation.PersonID, donation.FundID,
		donation.AmountCents, donation.Currency, donation.PaymentMethod,
		donation.StripePaymentIntentID,
		donation.Status, donation.IsRecurring, donation.DonatedAt,
		donation.DonorName, donation.DonorEmail,
	)

	return err
}

// GetPublicTenantInfo returns public info for the giving page
func (s *StripeService) GetPublicTenantInfo(ctx context.Context, slug string) (*PublicTenantInfo, error) {
	var info PublicTenantInfo
	var stripeAccountID *string

	err := s.db.QueryRow(ctx,
		`SELECT id, name, COALESCE(slug, ''), COALESCE(logo, ''), COALESCE(about, ''), stripe_account_id
		 FROM tenants WHERE slug = $1`,
		slug,
	).Scan(&info.ID, &info.Name, &info.Slug, &info.Logo, &info.About, &stripeAccountID)
	if err != nil {
		return nil, fmt.Errorf("church not found")
	}

	info.GivingEnabled = stripeAccountID != nil && *stripeAccountID != ""

	// Get active funds (bypass RLS for public access)
	rows, err := s.db.Query(ctx,
		`SELECT id, name, COALESCE(description, ''), is_default FROM funds WHERE tenant_id = $1 AND is_active = TRUE ORDER BY is_default DESC, name ASC`,
		info.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f PublicFund
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.IsDefault); err != nil {
			return nil, err
		}
		info.Funds = append(info.Funds, f)
	}

	return &info, nil
}

// --- Helpers ---

func calculateAppFee(amountCents int) int {
	fee := amountCents / 200 // 0.5%
	if fee < 30 {
		fee = 30 // Minimum 30 cents
	}
	return fee
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
