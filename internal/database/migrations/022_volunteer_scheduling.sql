-- Volunteer Teams (e.g., "Worship Team", "Tech Team", "Hospitality")
CREATE TABLE IF NOT EXISTS volunteer_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#4A8B8C',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE volunteer_teams ENABLE ROW LEVEL SECURITY;

-- Volunteer teams can only see their own tenant's data
CREATE POLICY volunteer_teams_isolation_policy ON volunteer_teams
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_volunteer_teams_tenant_id ON volunteer_teams(tenant_id);
CREATE INDEX IF NOT EXISTS idx_volunteer_teams_active ON volunteer_teams(tenant_id, is_active);

-- Updated_at trigger
CREATE TRIGGER update_volunteer_teams_updated_at
    BEFORE UPDATE ON volunteer_teams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Team Members (assign people to volunteer teams)
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES volunteer_teams(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    role VARCHAR(100), -- "Leader", "Member", "Backup", etc.
    is_active BOOLEAN DEFAULT TRUE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, person_id)
);

-- Enable RLS (inherit from volunteer_teams)
ALTER TABLE team_members ENABLE ROW LEVEL SECURITY;

CREATE POLICY team_members_isolation_policy ON team_members
    USING (EXISTS (
        SELECT 1 FROM volunteer_teams 
        WHERE volunteer_teams.id = team_members.team_id 
        AND volunteer_teams.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_person_id ON team_members(person_id);
CREATE INDEX IF NOT EXISTS idx_team_members_active ON team_members(is_active);

-- Volunteer Availability (blackout dates)
CREATE TABLE IF NOT EXISTS volunteer_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    team_id UUID REFERENCES volunteer_teams(id) ON DELETE CASCADE, -- optional, can be team-specific
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason VARCHAR(255), -- "Vacation", "Out of town", "Busy", etc.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE volunteer_availability ENABLE ROW LEVEL SECURITY;

CREATE POLICY volunteer_availability_isolation_policy ON volunteer_availability
    USING (EXISTS (
        SELECT 1 FROM people 
        WHERE people.id = volunteer_availability.person_id 
        AND people.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX IF NOT EXISTS idx_volunteer_availability_person_id ON volunteer_availability(person_id);
CREATE INDEX IF NOT EXISTS idx_volunteer_availability_dates ON volunteer_availability(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_volunteer_availability_team_id ON volunteer_availability(team_id);

-- Updated_at trigger
CREATE TRIGGER update_volunteer_availability_updated_at
    BEFORE UPDATE ON volunteer_availability
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add team_id to service_teams to link to volunteer teams
ALTER TABLE service_teams ADD COLUMN team_id UUID REFERENCES volunteer_teams(id);
CREATE INDEX IF NOT EXISTS idx_service_teams_team_id ON service_teams(team_id);

-- Add confirmation timestamp
ALTER TABLE service_teams ADD COLUMN responded_at TIMESTAMP;

-- Add notification tracking
ALTER TABLE service_teams ADD COLUMN notified_at TIMESTAMP;
ALTER TABLE service_teams ADD COLUMN notification_sent BOOLEAN DEFAULT FALSE;

-- Helper view for upcoming service assignments
CREATE VIEW upcoming_service_assignments AS
SELECT 
    st.id,
    st.service_id,
    st.person_id,
    st.role,
    st.status,
    st.team_id,
    st.notified_at,
    st.responded_at,
    p.first_name,
    p.last_name,
    p.email,
    s.service_date,
    s.service_time,
    s.tenant_id,
    vt.name as team_name
FROM service_teams st
JOIN people p ON p.id = st.person_id
JOIN services s ON s.id = st.service_id
LEFT JOIN volunteer_teams vt ON vt.id = st.team_id
WHERE s.service_date >= CURRENT_DATE
ORDER BY s.service_date, s.service_time;

-- Helper function to check if someone is available
CREATE OR REPLACE FUNCTION is_person_available(
    p_person_id UUID,
    p_date DATE
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN NOT EXISTS (
        SELECT 1 FROM volunteer_availability
        WHERE person_id = p_person_id
        AND p_date BETWEEN start_date AND end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Helper function to detect conflicts (person scheduled multiple times same day)
CREATE OR REPLACE FUNCTION get_scheduling_conflicts(
    p_service_date DATE,
    p_tenant_id UUID
) RETURNS TABLE(
    person_id UUID,
    first_name VARCHAR,
    last_name VARCHAR,
    service_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        st.person_id,
        p.first_name,
        p.last_name,
        COUNT(*) as service_count
    FROM service_teams st
    JOIN services s ON s.id = st.service_id
    JOIN people p ON p.id = st.person_id
    WHERE s.service_date = p_service_date
    AND s.tenant_id = p_tenant_id
    GROUP BY st.person_id, p.first_name, p.last_name
    HAVING COUNT(*) > 1;
END;
$$ LANGUAGE plpgsql;
