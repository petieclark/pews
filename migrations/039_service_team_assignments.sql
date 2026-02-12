-- Service Team Assignments: per-service volunteer scheduling
CREATE TABLE IF NOT EXISTS service_team_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id),
    position_id UUID REFERENCES team_positions(id),
    person_id UUID NOT NULL REFERENCES people(id),
    status VARCHAR(20) DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(service_id, position_id, person_id)
);

CREATE INDEX IF NOT EXISTS idx_sta_service ON service_team_assignments(service_id);
CREATE INDEX IF NOT EXISTS idx_sta_person ON service_team_assignments(person_id);
CREATE INDEX IF NOT EXISTS idx_sta_tenant ON service_team_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sta_team ON service_team_assignments(team_id);

-- RLS
ALTER TABLE service_team_assignments ENABLE ROW LEVEL SECURITY;
CREATE POLICY sta_isolation_policy ON service_team_assignments
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
