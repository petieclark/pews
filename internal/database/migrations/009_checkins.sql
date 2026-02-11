-- Check-in stations (e.g., "Main Lobby", "Kids Room A", "Youth Center")
CREATE TABLE checkin_stations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE checkin_stations ENABLE ROW LEVEL SECURITY;

CREATE POLICY checkin_stations_isolation_policy ON checkin_stations
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_checkin_stations_tenant_id ON checkin_stations(tenant_id);
CREATE INDEX idx_checkin_stations_active ON checkin_stations(tenant_id, is_active);

CREATE TRIGGER update_checkin_stations_updated_at
    BEFORE UPDATE ON checkin_stations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Check-in events (e.g., "Sunday Morning 10:30 AM - Feb 9, 2025")
CREATE TABLE checkin_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    event_date DATE NOT NULL,
    service_id UUID REFERENCES services(id),
    station_id UUID REFERENCES checkin_stations(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE checkin_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY checkin_events_isolation_policy ON checkin_events
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_checkin_events_tenant_id ON checkin_events(tenant_id);
CREATE INDEX idx_checkin_events_date ON checkin_events(event_date);
CREATE INDEX idx_checkin_events_service ON checkin_events(service_id);
CREATE INDEX idx_checkin_events_active ON checkin_events(tenant_id, is_active);

CREATE TRIGGER update_checkin_events_updated_at
    BEFORE UPDATE ON checkin_events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Individual check-ins
CREATE TABLE checkins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_id UUID NOT NULL REFERENCES checkin_events(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id),
    station_id UUID REFERENCES checkin_stations(id),
    first_time BOOLEAN DEFAULT FALSE,
    checked_in_at TIMESTAMP NOT NULL DEFAULT NOW(),
    checked_out_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE checkins ENABLE ROW LEVEL SECURITY;

CREATE POLICY checkins_isolation_policy ON checkins
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_checkins_tenant_id ON checkins(tenant_id);
CREATE INDEX idx_checkins_event_id ON checkins(event_id);
CREATE INDEX idx_checkins_person_id ON checkins(person_id);
CREATE INDEX idx_checkins_station_id ON checkins(station_id);
CREATE INDEX idx_checkins_checked_in_at ON checkins(checked_in_at);
CREATE INDEX idx_checkins_first_time ON checkins(tenant_id, first_time) WHERE first_time = TRUE;

-- Medical alerts for people (allergies, medical conditions, dietary needs)
CREATE TABLE medical_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL, -- allergy, medical, dietary
    severity VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE medical_alerts ENABLE ROW LEVEL SECURITY;

CREATE POLICY medical_alerts_isolation_policy ON medical_alerts
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_medical_alerts_tenant_id ON medical_alerts(tenant_id);
CREATE INDEX idx_medical_alerts_person_id ON medical_alerts(person_id);

CREATE TRIGGER update_medical_alerts_updated_at
    BEFORE UPDATE ON medical_alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Authorized pickups (child → authorized person with relationship)
CREATE TABLE authorized_pickups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    child_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    pickup_person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    relationship VARCHAR(100) NOT NULL, -- parent, grandparent, guardian, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(child_id, pickup_person_id)
);

ALTER TABLE authorized_pickups ENABLE ROW LEVEL SECURITY;

CREATE POLICY authorized_pickups_isolation_policy ON authorized_pickups
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

CREATE INDEX idx_authorized_pickups_tenant_id ON authorized_pickups(tenant_id);
CREATE INDEX idx_authorized_pickups_child_id ON authorized_pickups(child_id);
CREATE INDEX idx_authorized_pickups_pickup_person ON authorized_pickups(pickup_person_id);

CREATE TRIGGER update_authorized_pickups_updated_at
    BEFORE UPDATE ON authorized_pickups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
