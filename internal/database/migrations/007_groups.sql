-- Groups table
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    group_type VARCHAR(50) DEFAULT 'small_group', -- small_group, bible_study, ministry_team, class
    meeting_day VARCHAR(20), -- monday, tuesday, etc.
    meeting_time VARCHAR(20), -- "7:00 PM"
    meeting_location VARCHAR(255),
    is_public BOOLEAN DEFAULT TRUE, -- visible in group finder
    max_members INTEGER,
    is_active BOOLEAN DEFAULT TRUE,
    photo_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE groups ENABLE ROW LEVEL SECURITY;

-- Groups can only see their own tenant's data
CREATE POLICY groups_isolation_policy ON groups
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_groups_tenant_id ON groups(tenant_id);
CREATE INDEX idx_groups_type ON groups(group_type);
CREATE INDEX idx_groups_active ON groups(is_active);

-- Updated_at trigger
CREATE TRIGGER update_groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Group members junction
CREATE TABLE group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- leader, co_leader, member
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(group_id, person_id)
);

-- Enable RLS
ALTER TABLE group_members ENABLE ROW LEVEL SECURITY;

-- Group members inherit tenant access from groups
CREATE POLICY group_members_isolation_policy ON group_members
    USING (EXISTS (
        SELECT 1 FROM groups 
        WHERE groups.id = group_members.group_id 
        AND groups.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_person_id ON group_members(person_id);
