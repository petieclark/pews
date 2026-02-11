-- Create import_history table for tracking PCO and other imports
CREATE TABLE IF NOT EXISTS import_history (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    import_type VARCHAR(50) NOT NULL, -- 'pco_people', 'pco_songs', 'people', 'groups', 'songs', 'giving'
    imported_count INT NOT NULL DEFAULT 0,
    updated_count INT NOT NULL DEFAULT 0,
    skipped_count INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    imported_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_import_history_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Index for querying history by tenant and date
CREATE INDEX IF NOT EXISTS idx_import_history_tenant_date ON import_history(tenant_id, imported_at DESC);

-- Enable RLS
ALTER TABLE import_history ENABLE ROW LEVEL SECURITY;

-- RLS Policy
CREATE POLICY tenant_isolation_policy ON import_history
    USING (tenant_id::text = current_setting('app.current_tenant_id', TRUE));
