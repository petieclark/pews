-- People/members table
CREATE TABLE people (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip VARCHAR(20),
    birthdate DATE,
    gender VARCHAR(20),
    membership_status VARCHAR(50) DEFAULT 'active',
    photo_url TEXT,
    notes TEXT,
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE people ENABLE ROW LEVEL SECURITY;

-- People can only see their own tenant's data
CREATE POLICY people_isolation_policy ON people
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_people_tenant_id ON people(tenant_id);
CREATE INDEX idx_people_email ON people(email);
CREATE INDEX idx_people_name ON people(first_name, last_name);

-- Updated_at trigger
CREATE TRIGGER update_people_updated_at
    BEFORE UPDATE ON people
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Households
CREATE TABLE households (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    primary_contact_id UUID REFERENCES people(id),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE households ENABLE ROW LEVEL SECURITY;

-- Households can only see their own tenant's data
CREATE POLICY households_isolation_policy ON households
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_households_tenant_id ON households(tenant_id);

-- Updated_at trigger
CREATE TRIGGER update_households_updated_at
    BEFORE UPDATE ON households
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Household members junction
CREATE TABLE household_members (
    household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    PRIMARY KEY (household_id, person_id)
);

-- Enable RLS
ALTER TABLE household_members ENABLE ROW LEVEL SECURITY;

-- Household members inherit tenant access from people
CREATE POLICY household_members_isolation_policy ON household_members
    USING (EXISTS (
        SELECT 1 FROM people 
        WHERE people.id = household_members.person_id 
        AND people.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_household_members_household_id ON household_members(household_id);
CREATE INDEX idx_household_members_person_id ON household_members(person_id);

-- Tags for people
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#4A8B8C',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Enable RLS
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;

-- Tags can only see their own tenant's data
CREATE POLICY tags_isolation_policy ON tags
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_tags_tenant_id ON tags(tenant_id);

-- Person tags junction
CREATE TABLE person_tags (
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (person_id, tag_id)
);

-- Enable RLS
ALTER TABLE person_tags ENABLE ROW LEVEL SECURITY;

-- Person tags inherit tenant access from people
CREATE POLICY person_tags_isolation_policy ON person_tags
    USING (EXISTS (
        SELECT 1 FROM people 
        WHERE people.id = person_tags.person_id 
        AND people.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_person_tags_person_id ON person_tags(person_id);
CREATE INDEX idx_person_tags_tag_id ON person_tags(tag_id);
