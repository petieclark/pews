-- Stripe test mode for tenants
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_test_mode BOOLEAN DEFAULT TRUE;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_test_account_id VARCHAR(255);

-- Add stripe_checkout_session_id to donations for webhook reconciliation
ALTER TABLE donations ADD COLUMN IF NOT EXISTS stripe_checkout_session_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_donations_stripe_checkout_session_id ON donations(stripe_checkout_session_id);

-- Add donor_name and donor_email for anonymous/public donations
ALTER TABLE donations ADD COLUMN IF NOT EXISTS donor_name VARCHAR(255);
ALTER TABLE donations ADD COLUMN IF NOT EXISTS donor_email VARCHAR(255);
