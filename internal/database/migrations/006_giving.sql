-- Add stripe_account_id to tenants
ALTER TABLE tenants ADD COLUMN stripe_account_id VARCHAR(255);
ALTER TABLE tenants ADD COLUMN stripe_onboarding_completed BOOLEAN DEFAULT FALSE;

-- Funds/designations
CREATE TABLE funds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS on funds
ALTER TABLE funds ENABLE ROW LEVEL SECURITY;

CREATE POLICY fund_tenant_isolation_policy ON funds
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_funds_tenant_id ON funds(tenant_id);
CREATE INDEX idx_funds_is_active ON funds(is_active);

-- Updated_at trigger
CREATE TRIGGER update_funds_updated_at
    BEFORE UPDATE ON funds
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Donations
CREATE TABLE donations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID REFERENCES people(id) ON DELETE SET NULL,
    fund_id UUID NOT NULL REFERENCES funds(id) ON DELETE RESTRICT,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),
    stripe_payment_intent_id VARCHAR(255),
    stripe_charge_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'completed',
    is_recurring BOOLEAN DEFAULT FALSE,
    recurring_frequency VARCHAR(20),
    stripe_subscription_id VARCHAR(255),
    memo TEXT,
    donated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS on donations
ALTER TABLE donations ENABLE ROW LEVEL SECURITY;

CREATE POLICY donation_tenant_isolation_policy ON donations
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_donations_tenant_id ON donations(tenant_id);
CREATE INDEX idx_donations_person_id ON donations(person_id);
CREATE INDEX idx_donations_fund_id ON donations(fund_id);
CREATE INDEX idx_donations_donated_at ON donations(donated_at DESC);
CREATE INDEX idx_donations_status ON donations(status);
CREATE INDEX idx_donations_stripe_payment_intent_id ON donations(stripe_payment_intent_id);

-- Updated_at trigger
CREATE TRIGGER update_donations_updated_at
    BEFORE UPDATE ON donations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Giving statements (annual tax receipts)
CREATE TABLE giving_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    total_cents INTEGER NOT NULL,
    generated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    pdf_url TEXT,
    UNIQUE(tenant_id, person_id, year)
);

-- Enable RLS on giving_statements
ALTER TABLE giving_statements ENABLE ROW LEVEL SECURITY;

CREATE POLICY giving_statement_tenant_isolation_policy ON giving_statements
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_giving_statements_tenant_id ON giving_statements(tenant_id);
CREATE INDEX idx_giving_statements_person_id ON giving_statements(person_id);
CREATE INDEX idx_giving_statements_year ON giving_statements(year);
