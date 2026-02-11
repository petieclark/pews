-- Sermon notes and podcast feed
CREATE TABLE sermon_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    service_id UUID REFERENCES services(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    speaker VARCHAR(255) NOT NULL,
    sermon_date DATE NOT NULL,
    scripture_reference VARCHAR(500),
    notes_text TEXT,
    audio_url TEXT,
    video_url TEXT,
    series_name VARCHAR(255),
    audio_duration_seconds INTEGER,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE sermon_notes ENABLE ROW LEVEL SECURITY;

-- Sermon notes can only see their own tenant's data
CREATE POLICY sermon_notes_isolation_policy ON sermon_notes
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_sermon_notes_tenant_id ON sermon_notes(tenant_id);
CREATE INDEX idx_sermon_notes_date ON sermon_notes(sermon_date DESC);
CREATE INDEX idx_sermon_notes_series ON sermon_notes(tenant_id, series_name);
CREATE INDEX idx_sermon_notes_speaker ON sermon_notes(tenant_id, speaker);
CREATE INDEX idx_sermon_notes_published ON sermon_notes(tenant_id, published);
CREATE INDEX idx_sermon_notes_service_id ON sermon_notes(service_id);

-- Updated_at trigger
CREATE TRIGGER update_sermon_notes_updated_at
    BEFORE UPDATE ON sermon_notes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
