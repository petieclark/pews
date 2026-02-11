package giving

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/accountlink"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/paymentintent"
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

// Stripe Connect

func (s *StripeService) GetTenantName(ctx context.Context, tenantID string) (string, error) {
	var name string
	err := s.db.QueryRow(ctx, `SELECT name FROM tenants WHERE id = $1`, tenantID).Scan(&name)
	return name, err
}

func (s *StripeService) CreateConnectOnboardingLink(ctx context.Context, tenantID, tenantName, tenantEmail string) (string, error) {
	// Check if account already exists
	var stripeAccountID string
	err := s.db.QueryRow(ctx,
		`SELECT stripe_account_id FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID)
	if err != nil {
		return "", err
	}

	// Create account if doesn't exist
	if stripeAccountID == "" {
		params := &stripe.AccountParams{
			Type: stripe.String(string(stripe.AccountTypeExpress)),
			Country: stripe.String("US"),
			Email: stripe.String(tenantEmail),
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

		// Save to database
		_, err = s.db.Exec(ctx,
			`UPDATE tenants SET stripe_account_id = $1 WHERE id = $2`,
			acc.ID, tenantID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to save Stripe account ID: %w", err)
		}

		stripeAccountID = acc.ID
	}

	// Create account link with hardcoded return URLs (TODO: use env var)
	returnURL := "http://localhost:5273/dashboard/giving/settings?setup=complete"
	refreshURL := "http://localhost:5273/dashboard/giving/settings?setup=refresh"

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
	// Get existing account ID
	var stripeAccountID string
	err := s.db.QueryRow(ctx,
		`SELECT stripe_account_id FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID)
	if err != nil || stripeAccountID == "" {
		return "", fmt.Errorf("no Stripe account found for this tenant")
	}

	// Create new account link
	returnURL := "http://localhost:5273/dashboard/giving/settings?setup=complete"
	refreshURL := "http://localhost:5273/dashboard/giving/settings?setup=refresh"

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

	err := s.db.QueryRow(ctx,
		`SELECT stripe_account_id, stripe_onboarding_completed FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&stripeAccountID, &onboardingCompleted)
	if err != nil {
		return nil, err
	}

	status := &StripeConnectStatus{
		Connected:           stripeAccountID != nil && *stripeAccountID != "",
		OnboardingCompleted: onboardingCompleted,
	}

	if status.Connected {
		status.AccountID = *stripeAccountID

		// Get account details from Stripe
		acc, err := account.GetByID(*stripeAccountID, nil)
		if err == nil {
			status.ChargesEnabled = acc.ChargesEnabled
			status.PayoutsEnabled = acc.PayoutsEnabled
			
			// Update onboarding status if details submitted
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

// Donations via Stripe

func (s *StripeService) CreateCheckoutSession(ctx context.Context, tenantID, personID, fundID string, amountCents int) (string, error) {
	// Get tenant's Stripe account ID
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

	// Get fund name
	var fundName string
	err = s.db.QueryRow(ctx, `SELECT name FROM funds WHERE id = $1`, fundID).Scan(&fundName)
	if err != nil {
		return "", fmt.Errorf("fund not found")
	}

	// Calculate application fee (1%)
	appFeeCents := amountCents / 100
	if appFeeCents < 30 {
		appFeeCents = 30 // Minimum 30 cents
	}

	// Create checkout session
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

func (s *StripeService) HandlePaymentIntentSucceeded(ctx context.Context, pi *stripe.PaymentIntent) error {
	tenantID := pi.Metadata["tenant_id"]
	personID := pi.Metadata["person_id"]
	fundID := pi.Metadata["fund_id"]

	if tenantID == "" || fundID == "" {
		return fmt.Errorf("missing metadata")
	}

	// Check if donation already recorded
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM donations WHERE stripe_payment_intent_id = $1)`,
		pi.ID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already recorded
	}

	// Create donation record
	var personIDPtr *string
	if personID != "" {
		personIDPtr = &personID
	}

	paymentMethod := "card"
	chargeID := ""
	if pi.LatestCharge != nil {
		chargeID = pi.LatestCharge.ID
		if pi.LatestCharge.PaymentMethodDetails != nil {
			if string(pi.LatestCharge.PaymentMethodDetails.Type) == "card" {
				paymentMethod = "card"
			} else if string(pi.LatestCharge.PaymentMethodDetails.Type) == "us_bank_account" {
				paymentMethod = "ach"
			}
		}
	}

	donation := &Donation{
		ID:                    uuid.New().String(),
		TenantID:              tenantID,
		PersonID:              personIDPtr,
		FundID:                fundID,
		AmountCents:           int(pi.Amount),
		Currency:              string(pi.Currency),
		PaymentMethod:         &paymentMethod,
		StripePaymentIntentID: &pi.ID,
		StripeChargeID:        &chargeID,
		Status:                "completed",
		IsRecurring:           false,
		DonatedAt:             time.Now(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO donations (id, tenant_id, person_id, fund_id, amount_cents, currency,
		                        payment_method, stripe_payment_intent_id, stripe_charge_id,
		                        status, is_recurring, donated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		donation.ID, donation.TenantID, donation.PersonID, donation.FundID,
		donation.AmountCents, donation.Currency, donation.PaymentMethod,
		donation.StripePaymentIntentID, donation.StripeChargeID,
		donation.Status, donation.IsRecurring, donation.DonatedAt,
	)

	return err
}

func (s *StripeService) CreatePaymentIntent(ctx context.Context, tenantID, personID, fundID string, amountCents int) (string, error) {
	// Get tenant's Stripe account ID
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

	// Calculate application fee (1%)
	appFeeCents := amountCents / 100
	if appFeeCents < 30 {
		appFeeCents = 30 // Minimum 30 cents
	}

	params := &stripe.PaymentIntentParams{
		Amount:               stripe.Int64(int64(amountCents)),
		Currency:             stripe.String("usd"),
		ApplicationFeeAmount: stripe.Int64(int64(appFeeCents)),
		Metadata: map[string]string{
			"tenant_id": tenantID,
			"person_id": personID,
			"fund_id":   fundID,
		},
	}
	params.SetStripeAccount(*stripeAccountID)

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	return pi.ClientSecret, nil
}

// Public checkout (kiosk)

func (s *StripeService) CreatePublicCheckoutSession(ctx context.Context, tenantID, fundID string, amountCents int, name, email *string) (string, error) {
	// Get tenant's Stripe account ID
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

	// Get fund name
	var fundName string
	err = s.db.QueryRow(ctx, `SELECT name FROM funds WHERE id = $1 AND tenant_id = $2`, fundID, tenantID).Scan(&fundName)
	if err != nil {
		return "", fmt.Errorf("fund not found")
	}

	// Calculate application fee (1%)
	appFeeCents := amountCents / 100
	if appFeeCents < 30 {
		appFeeCents = 30 // Minimum 30 cents
	}

	// Build metadata
	metadata := map[string]string{
		"tenant_id": tenantID,
		"fund_id":   fundID,
		"kiosk":     "true",
	}
	if name != nil && *name != "" {
		metadata["donor_name"] = *name
	}
	if email != nil && *email != "" {
		metadata["donor_email"] = *email
	}

	// Create checkout session
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

	// Add email if provided
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
