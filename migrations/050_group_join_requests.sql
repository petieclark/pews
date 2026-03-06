-- Group join requests for group finder (Issue #62)
CREATE TABLE IF NOT EXISTS group_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    message TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'declined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_group_join_requests_tenant ON group_join_requests(tenant_id);
CREATE INDEX idx_group_join_requests_group ON group_join_requests(group_id);
CREATE INDEX idx_group_join_requests_status ON group_join_requests(tenant_id, status);

ALTER TABLE group_join_requests ENABLE ROW LEVEL SECURITY;
CREATE POLICY group_join_requests_isolation_policy ON group_join_requests
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
