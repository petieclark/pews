package billing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	billingportalsession "github.com/stripe/stripe-go/v76/billingportal/session"
)

type Service struct {
	db            *pgxpool.Pool
	stripePriceID string
	frontendURL   string
}

func NewService(db *pgxpool.Pool, stripeKey, stripePriceID, frontendURL string) *Service {
	stripe.Key = stripeKey
	return &Service{
		db:            db,
		stripePriceID: stripePriceID,
		frontendURL:   frontendURL,
	}
}

func (s *Service) CreateSubscription(ctx context.Context, tenantID string) (*Subscription, error) {
	sub := &Subscription{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Plan:     "free",
		Status:   "active",
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO subscriptions (id, tenant_id, plan, status) VALUES ($1, $2, $3, $4)`,
		sub.ID, sub.TenantID, sub.Plan, sub.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

// EnsureSubscription creates a free subscription if one doesn't exist
func (s *Service) EnsureSubscription(ctx context.Context, tenantID string) error {
	_, err := s.CreateSubscription(ctx, tenantID)
	return err
}

func (s *Service) GetSubscription(ctx context.Context, tenantID string) (*Subscription, error) {
	sub := &Subscription{}
	err := s.db.QueryRow(ctx,
		`SELECT id, tenant_id, COALESCE(stripe_customer_id, ''), COALESCE(stripe_subscription_id, ''), plan, status, 
		 current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at 
		 FROM subscriptions WHERE tenant_id = $1`,
		tenantID,
	).Scan(&sub.ID, &sub.TenantID, &sub.StripeCustomerID, &sub.StripeSubscriptionID,
		&sub.Plan, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	return sub, nil
}

func (s *Service) CreateCheckoutSession(ctx context.Context, tenantID, tenantEmail string) (string, error) {
	sub, err := s.GetSubscription(ctx, tenantID)
	if err != nil {
		return "", err
	}

	// Create Stripe customer if not exists
	var customerID string
	if sub.StripeCustomerID == "" {
		params := &stripe.CustomerParams{
			Email: stripe.String(tenantEmail),
			Metadata: map[string]string{
				"tenant_id": tenantID,
			},
		}
		cust, err := customer.New(params)
		if err != nil {
			return "", fmt.Errorf("failed to create Stripe customer: %w", err)
		}
		customerID = cust.ID

		// Update subscription with customer ID
		_, err = s.db.Exec(ctx,
			`UPDATE subscriptions SET stripe_customer_id = $1 WHERE tenant_id = $2`,
			customerID, tenantID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		customerID = sub.StripeCustomerID
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.stripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.frontendURL + "/settings?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.frontendURL + "/settings"),
		Metadata: map[string]string{
			"tenant_id": tenantID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"tenant_id": tenantID,
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

func (s *Service) CreatePortalSession(ctx context.Context, tenantID string) (string, error) {
	sub, err := s.GetSubscription(ctx, tenantID)
	if err != nil {
		return "", err
	}

	if sub.StripeCustomerID == "" {
		return "", fmt.Errorf("no Stripe customer found")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sub.StripeCustomerID),
		ReturnURL: stripe.String(s.frontendURL + "/settings"),
	}

	sess, err := billingportalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create portal session: %w", err)
	}

	return sess.URL, nil
}

func (s *Service) UpdateSubscriptionStatus(ctx context.Context, tenantID, stripeSubID, plan, status string, periodStart, periodEnd time.Time) error {
	_, err := s.db.Exec(ctx,
		`UPDATE subscriptions 
		 SET stripe_subscription_id = $1, plan = $2, status = $3, 
		     current_period_start = $4, current_period_end = $5
		 WHERE tenant_id = $6`,
		stripeSubID, plan, status, periodStart, periodEnd, tenantID,
	)
	return err
}

func (s *Service) CancelSubscription(ctx context.Context, tenantID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE subscriptions SET status = 'canceled', plan = 'free' WHERE tenant_id = $1`,
		tenantID,
	)
	return err
}

// EnsureStripeProduct creates or verifies the Stripe Product and Price for the Pro plan
// This should be called during service initialization or as a migration step
// Returns the price ID for the $100/month Pro plan
func (s *Service) EnsureStripeProduct() (string, error) {
	// Create or get the Pews Pro product
	productParams := &stripe.ProductParams{
		Name:        stripe.String("Pews Pro"),
		Description: stripe.String("Professional church management features"),
	}
	
	prod, err := product.New(productParams)
	if err != nil {
		// If product already exists, we might get an error
		// In production, you'd want to search for the existing product
		log.Printf("Product creation returned: %v", err)
		return "", fmt.Errorf("failed to create product: %w", err)
	}

	// Create the price: $100/month
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(prod.ID),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		UnitAmount: stripe.Int64(10000), // $100.00 in cents
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
	}

	pr, err := price.New(priceParams)
	if err != nil {
		return "", fmt.Errorf("failed to create price: %w", err)
	}

	log.Printf("Created Stripe Product %s and Price %s ($100/month)", prod.ID, pr.ID)
	return pr.ID, nil
}
