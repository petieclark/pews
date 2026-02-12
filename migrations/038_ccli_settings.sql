CREATE TABLE IF NOT EXISTS ccli_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) UNIQUE,
    license_number VARCHAR(50),
    auto_report BOOLEAN DEFAULT FALSE,
    report_frequency VARCHAR(20) DEFAULT 'quarterly',
    last_reported_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
