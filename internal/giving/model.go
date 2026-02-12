package giving

import (
	"time"
)

type Fund struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Donation struct {
	ID                      string     `json:"id"`
	TenantID                string     `json:"tenant_id"`
	PersonID                *string    `json:"person_id"`
	PersonName              *string    `json:"person_name,omitempty"`
	FundID                  string     `json:"fund_id"`
	FundName                string     `json:"fund_name,omitempty"`
	AmountCents             int        `json:"amount_cents"`
	AmountDisplay           string     `json:"amount_display,omitempty"`
	Currency                string     `json:"currency"`
	PaymentMethod           *string    `json:"payment_method"`
	StripePaymentIntentID   *string    `json:"stripe_payment_intent_id,omitempty"`
	StripeChargeID          *string    `json:"stripe_charge_id,omitempty"`
	StripeCheckoutSessionID *string    `json:"stripe_checkout_session_id,omitempty"`
	Status                  string     `json:"status"`
	IsRecurring             bool       `json:"is_recurring"`
	RecurringFrequency      *string    `json:"recurring_frequency,omitempty"`
	StripeSubscriptionID    *string    `json:"stripe_subscription_id,omitempty"`
	Memo                    *string    `json:"memo,omitempty"`
	DonorName               *string    `json:"donor_name,omitempty"`
	DonorEmail              *string    `json:"donor_email,omitempty"`
	DonatedAt               time.Time  `json:"donated_at"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type GivingStatement struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	PersonID    string    `json:"person_id"`
	PersonName  string    `json:"person_name,omitempty"`
	Year        int       `json:"year"`
	TotalCents  int       `json:"total_cents"`
	GeneratedAt time.Time `json:"generated_at"`
	PDFURL      *string   `json:"pdf_url,omitempty"`
}

type GivingStats struct {
	TotalThisMonth  int            `json:"total_this_month"`
	TotalThisYear   int            `json:"total_this_year"`
	TotalAllTime    int            `json:"total_all_time"`
	DonationCount   int            `json:"donation_count"`
	DonorCount      int            `json:"donor_count"`
	AverageDonation int            `json:"average_donation"`
	FundBreakdown   []FundSummary  `json:"fund_breakdown"`
	MonthlyTrend    []MonthlyTotal `json:"monthly_trend"`
}

type FundSummary struct {
	FundID      string `json:"fund_id"`
	FundName    string `json:"fund_name"`
	TotalCents  int    `json:"total_cents"`
	DonorCount  int    `json:"donor_count"`
	Percentage  float64 `json:"percentage"`
}

type MonthlyTotal struct {
	Month      string `json:"month"`
	TotalCents int    `json:"total_cents"`
	Count      int    `json:"count"`
}

type StripeConnectStatus struct {
	Connected            bool   `json:"connected"`
	AccountID            string `json:"account_id,omitempty"`
	OnboardingCompleted  bool   `json:"onboarding_completed"`
	ChargesEnabled       bool   `json:"charges_enabled"`
	PayoutsEnabled       bool   `json:"payouts_enabled"`
	IsTestMode           bool   `json:"is_test_mode"`
}

type KioskConfig struct {
	Enabled          bool     `json:"enabled"`
	QuickAmounts     []int    `json:"quick_amounts"`      // Amount in cents
	DefaultFundID    *string  `json:"default_fund_id"`
	ThankYouMessage  string   `json:"thank_you_message"`
}

// Public giving page types

type PublicGiveRequest struct {
	TenantSlug  string `json:"tenant_slug"`
	FundID      string `json:"fund_id"`
	AmountCents int    `json:"amount_cents"`
	DonorName   string `json:"donor_name"`
	DonorEmail  string `json:"donor_email"`
	Frequency   string `json:"frequency"` // "one-time" or "monthly"
}

type PublicTenantInfo struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Slug          string       `json:"slug"`
	Logo          string       `json:"logo"`
	About         string       `json:"about"`
	GivingEnabled bool         `json:"giving_enabled"`
	Funds         []PublicFund `json:"funds"`
}

type PublicFund struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}
