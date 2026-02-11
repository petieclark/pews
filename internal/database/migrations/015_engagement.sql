-- Engagement scores table
CREATE TABLE engagement_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    score INTEGER NOT NULL CHECK (score >= 0 AND score <= 100),
    attendance_score INTEGER NOT NULL DEFAULT 0,
    giving_score INTEGER NOT NULL DEFAULT 0,
    group_score INTEGER NOT NULL DEFAULT 0,
    volunteer_score INTEGER NOT NULL DEFAULT 0,
    connection_score INTEGER NOT NULL DEFAULT 0,
    calculated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, person_id)
);

ALTER TABLE engagement_scores ENABLE ROW LEVEL SECURITY;

CREATE POLICY engagement_scores_isolation_policy ON engagement_scores
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_engagement_scores_tenant_id ON engagement_scores(tenant_id);
CREATE INDEX idx_engagement_scores_person_id ON engagement_scores(person_id);
CREATE INDEX idx_engagement_scores_score ON engagement_scores(tenant_id, score DESC);

-- Engagement score history (for tracking changes over time)
CREATE TABLE engagement_score_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    score INTEGER NOT NULL CHECK (score >= 0 AND score <= 100),
    attendance_score INTEGER NOT NULL DEFAULT 0,
    giving_score INTEGER NOT NULL DEFAULT 0,
    group_score INTEGER NOT NULL DEFAULT 0,
    volunteer_score INTEGER NOT NULL DEFAULT 0,
    connection_score INTEGER NOT NULL DEFAULT 0,
    recorded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE engagement_score_history ENABLE ROW LEVEL SECURITY;

CREATE POLICY engagement_score_history_isolation_policy ON engagement_score_history
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_engagement_score_history_tenant_id ON engagement_score_history(tenant_id);
CREATE INDEX idx_engagement_score_history_person_id ON engagement_score_history(person_id);
CREATE INDEX idx_engagement_score_history_recorded_at ON engagement_score_history(recorded_at DESC);

-- Function to automatically save score history when engagement_scores is updated
CREATE OR REPLACE FUNCTION save_engagement_score_history()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO engagement_score_history (
        tenant_id, person_id, score, 
        attendance_score, giving_score, group_score, 
        volunteer_score, connection_score, recorded_at
    ) VALUES (
        NEW.tenant_id, NEW.person_id, NEW.score,
        NEW.attendance_score, NEW.giving_score, NEW.group_score,
        NEW.volunteer_score, NEW.connection_score, NEW.calculated_at
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER engagement_score_history_trigger
    AFTER INSERT OR UPDATE ON engagement_scores
    FOR EACH ROW
    EXECUTE FUNCTION save_engagement_score_history();

-- Add volunteer tracking to service team members (for volunteer scoring)
-- This is already handled via service_team_members table

-- Add "in_directory" flag to people for connection scoring
ALTER TABLE people ADD COLUMN in_directory BOOLEAN DEFAULT FALSE;
ALTER TABLE people ADD COLUMN profile_completed BOOLEAN DEFAULT FALSE;
