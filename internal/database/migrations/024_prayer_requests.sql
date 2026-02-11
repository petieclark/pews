-- Prayer requests table
CREATE TABLE prayer_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    person_id UUID REFERENCES people(id), -- nullable for anonymous submissions
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    request_text TEXT NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'pending', -- pending, praying, answered, archived
    connection_card_id UUID REFERENCES connection_cards(id), -- link to original connection card if applicable
    notes TEXT, -- staff notes for follow-up
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE prayer_requests ENABLE ROW LEVEL SECURITY;

-- Prayer requests can only see their own tenant's data
CREATE POLICY prayer_requests_isolation_policy ON prayer_requests
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_prayer_requests_tenant_id ON prayer_requests(tenant_id);
CREATE INDEX idx_prayer_requests_person_id ON prayer_requests(person_id);
CREATE INDEX idx_prayer_requests_status ON prayer_requests(status);
CREATE INDEX idx_prayer_requests_is_public ON prayer_requests(is_public);
CREATE INDEX idx_prayer_requests_submitted_at ON prayer_requests(submitted_at);
CREATE INDEX idx_prayer_requests_connection_card_id ON prayer_requests(connection_card_id);

-- Updated_at trigger
CREATE TRIGGER update_prayer_requests_updated_at
    BEFORE UPDATE ON prayer_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Prayer followers (staff following specific prayer requests)
CREATE TABLE prayer_followers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prayer_request_id UUID NOT NULL REFERENCES prayer_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(prayer_request_id, user_id)
);

-- Enable RLS
ALTER TABLE prayer_followers ENABLE ROW LEVEL SECURITY;

-- Prayer followers inherit tenant access from prayer_requests
CREATE POLICY prayer_followers_isolation_policy ON prayer_followers
    USING (EXISTS (
        SELECT 1 FROM prayer_requests 
        WHERE prayer_requests.id = prayer_followers.prayer_request_id 
        AND prayer_requests.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_prayer_followers_prayer_request_id ON prayer_followers(prayer_request_id);
CREATE INDEX idx_prayer_followers_user_id ON prayer_followers(user_id);
