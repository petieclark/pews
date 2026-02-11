-- Service types (e.g., "Sunday Morning", "Wednesday Night", "Special Event")
CREATE TABLE service_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    default_time VARCHAR(20), -- "10:30 AM"
    default_day VARCHAR(20), -- "sunday"
    color VARCHAR(7) DEFAULT '#4A8B8C',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE service_types ENABLE ROW LEVEL SECURITY;

-- Service types can only see their own tenant's data
CREATE POLICY service_types_isolation_policy ON service_types
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_service_types_tenant_id ON service_types(tenant_id);
CREATE INDEX idx_service_types_active ON service_types(tenant_id, is_active);

-- Updated_at trigger
CREATE TRIGGER update_service_types_updated_at
    BEFORE UPDATE ON service_types
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Individual service instances
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    service_type_id UUID NOT NULL REFERENCES service_types(id),
    name VARCHAR(255), -- override name, e.g., "Easter Sunday Service"
    service_date DATE NOT NULL,
    service_time VARCHAR(20),
    notes TEXT,
    status VARCHAR(50) DEFAULT 'planning', -- planning, confirmed, completed
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE services ENABLE ROW LEVEL SECURITY;

-- Services can only see their own tenant's data
CREATE POLICY services_isolation_policy ON services
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_services_tenant_id ON services(tenant_id);
CREATE INDEX idx_services_date ON services(service_date);
CREATE INDEX idx_services_type ON services(service_type_id);
CREATE INDEX idx_services_status ON services(status);

-- Updated_at trigger
CREATE TRIGGER update_services_updated_at
    BEFORE UPDATE ON services
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Songs database
CREATE TABLE songs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255),
    default_key VARCHAR(10), -- "G", "C", "Bb"
    tempo INTEGER, -- BPM
    ccli_number VARCHAR(50),
    lyrics TEXT,
    notes TEXT,
    tags VARCHAR(255), -- comma-separated: "worship,fast,opener"
    last_used DATE,
    times_used INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE songs ENABLE ROW LEVEL SECURITY;

-- Songs can only see their own tenant's data
CREATE POLICY songs_isolation_policy ON songs
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_songs_tenant_id ON songs(tenant_id);
CREATE INDEX idx_songs_title ON songs(title);
CREATE INDEX idx_songs_artist ON songs(artist);

-- Updated_at trigger
CREATE TRIGGER update_songs_updated_at
    BEFORE UPDATE ON songs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Service items (order of service)
CREATE TABLE service_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL, -- song, prayer, reading, sermon, announcement, other
    title VARCHAR(255) NOT NULL,
    song_id UUID REFERENCES songs(id),
    song_key VARCHAR(10), -- key for this performance
    position INTEGER NOT NULL, -- order in service
    duration_minutes INTEGER,
    notes TEXT,
    assigned_to VARCHAR(255) -- person name or role
);

-- Enable RLS
ALTER TABLE service_items ENABLE ROW LEVEL SECURITY;

-- Service items inherit tenant access from services
CREATE POLICY service_items_isolation_policy ON service_items
    USING (EXISTS (
        SELECT 1 FROM services 
        WHERE services.id = service_items.service_id 
        AND services.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_service_items_service_id ON service_items(service_id);
CREATE INDEX idx_service_items_position ON service_items(service_id, position);
CREATE INDEX idx_service_items_song_id ON service_items(song_id);

-- Team scheduling
CREATE TABLE service_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id),
    role VARCHAR(100) NOT NULL, -- "Worship Leader", "Keys", "Sound", "Greeter"
    status VARCHAR(50) DEFAULT 'pending', -- pending, accepted, declined
    notes TEXT,
    UNIQUE(service_id, person_id, role)
);

-- Enable RLS
ALTER TABLE service_teams ENABLE ROW LEVEL SECURITY;

-- Service teams inherit tenant access from services
CREATE POLICY service_teams_isolation_policy ON service_teams
    USING (EXISTS (
        SELECT 1 FROM services 
        WHERE services.id = service_teams.service_id 
        AND services.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_service_teams_service_id ON service_teams(service_id);
CREATE INDEX idx_service_teams_person_id ON service_teams(person_id);
CREATE INDEX idx_service_teams_status ON service_teams(status);
