-- Digest settings table
CREATE TABLE digest_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    send_day VARCHAR(10) NOT NULL DEFAULT 'monday',
    recipients TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_digest_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Enable RLS
ALTER TABLE digest_settings ENABLE ROW LEVEL SECURITY;

-- Only tenant can see their own settings
CREATE POLICY digest_settings_isolation_policy ON digest_settings
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- One settings row per tenant
CREATE UNIQUE INDEX idx_digest_settings_tenant ON digest_settings(tenant_id);

-- Updated_at trigger
CREATE TRIGGER update_digest_settings_updated_at
    BEFORE UPDATE ON digest_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Digest history table (track sent digests)
CREATE TABLE digest_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    recipients TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_digest_history_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Enable RLS
ALTER TABLE digest_history ENABLE ROW LEVEL SECURITY;

-- Only tenant can see their own history
CREATE POLICY digest_history_isolation_policy ON digest_history
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Index for querying history
CREATE INDEX idx_digest_history_tenant_date ON digest_history(tenant_id, week_start DESC);
