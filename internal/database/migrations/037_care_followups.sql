-- Care / Follow-ups module
CREATE TABLE IF NOT EXISTS follow_ups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'general', -- first_time_visitor, hospital_visit, counseling, general, membership
    priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- high, medium, low
    status VARCHAR(20) NOT NULL DEFAULT 'new', -- new, in_progress, waiting, completed
    due_date DATE,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS follow_up_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follow_up_id UUID NOT NULL REFERENCES follow_ups(id) ON DELETE CASCADE,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    note TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE follow_ups ENABLE ROW LEVEL SECURITY;
ALTER TABLE follow_up_notes ENABLE ROW LEVEL SECURITY;

CREATE POLICY follow_ups_isolation ON follow_ups
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE POLICY follow_up_notes_isolation ON follow_up_notes
    USING (EXISTS (
        SELECT 1 FROM follow_ups
        WHERE follow_ups.id = follow_up_notes.follow_up_id
        AND follow_ups.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_follow_ups_tenant ON follow_ups(tenant_id);
CREATE INDEX idx_follow_ups_person ON follow_ups(person_id);
CREATE INDEX idx_follow_ups_assigned ON follow_ups(assigned_to);
CREATE INDEX idx_follow_ups_status ON follow_ups(tenant_id, status);
CREATE INDEX idx_follow_ups_due ON follow_ups(tenant_id, due_date);
CREATE INDEX idx_follow_up_notes_follow_up ON follow_up_notes(follow_up_id);

-- Updated_at trigger
CREATE TRIGGER update_follow_ups_updated_at
    BEFORE UPDATE ON follow_ups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
