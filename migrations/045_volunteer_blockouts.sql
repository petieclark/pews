-- Volunteer Blockouts: allow volunteers to mark unavailable dates
CREATE TABLE IF NOT EXISTS volunteer_blockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    person_id UUID NOT NULL REFERENCES people(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    is_recurring BOOLEAN DEFAULT false,
    day_of_week INT,  -- 0=Sun..6=Sat, only used if is_recurring=true
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_vb_person ON volunteer_blockouts(person_id);
CREATE INDEX IF NOT EXISTS idx_vb_tenant ON volunteer_blockouts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vb_dates ON volunteer_blockouts(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_vb_recurring ON volunteer_blockouts(is_recurring) WHERE is_recurring = true;

-- RLS: isolate by tenant
ALTER TABLE volunteer_blockouts ENABLE ROW LEVEL SECURITY;
CREATE POLICY vb_isolation_policy ON volunteer_blockouts
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Ensure updated_at triggers exist (use existing function or create)
DROP TRIGGER IF EXISTS update_volunteer_blockouts_updated_at ON volunteer_blockouts;
CREATE TRIGGER update_volunteer_blockouts_updated_at
    BEFORE UPDATE ON volunteer_blockouts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
