-- Stream configurations
CREATE TABLE streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    service_id UUID REFERENCES services(id), -- optional link to a service
    status VARCHAR(50) DEFAULT 'scheduled', -- scheduled, live, ended, archived
    scheduled_start TIMESTAMP,
    actual_start TIMESTAMP,
    actual_end TIMESTAMP,
    -- Stream source (support multiple providers)
    stream_type VARCHAR(50) DEFAULT 'youtube', -- youtube, facebook, vimeo, rtmp_custom
    stream_url TEXT, -- YouTube/FB embed URL or RTMP ingest URL
    stream_key VARCHAR(255), -- for RTMP
    embed_url TEXT, -- the viewer-facing embed URL
    -- Engagement
    chat_enabled BOOLEAN DEFAULT TRUE,
    giving_enabled BOOLEAN DEFAULT TRUE, -- show "Give Now" button
    connection_card_enabled BOOLEAN DEFAULT TRUE, -- show connection card prompt
    viewer_count INTEGER DEFAULT 0,
    peak_viewers INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE streams ENABLE ROW LEVEL SECURITY;

-- Streams can only see their own tenant's data
CREATE POLICY streams_isolation_policy ON streams
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_streams_tenant_id ON streams(tenant_id);
CREATE INDEX idx_streams_status ON streams(status);
CREATE INDEX idx_streams_scheduled_start ON streams(scheduled_start);

-- Updated_at trigger
CREATE TRIGGER update_streams_updated_at
    BEFORE UPDATE ON streams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Stream chat messages
CREATE TABLE stream_chat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    person_id UUID REFERENCES people(id), -- null for anonymous/guest viewers
    guest_name VARCHAR(100), -- for non-logged-in viewers
    message TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE stream_chat ENABLE ROW LEVEL SECURITY;

-- Stream chat inherits tenant access from streams
CREATE POLICY stream_chat_isolation_policy ON stream_chat
    USING (EXISTS (
        SELECT 1 FROM streams 
        WHERE streams.id = stream_chat.stream_id 
        AND streams.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_stream_chat_stream_id ON stream_chat(stream_id);
CREATE INDEX idx_stream_chat_created_at ON stream_chat(created_at);

-- Stream viewer tracking
CREATE TABLE stream_viewers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    person_id UUID REFERENCES people(id),
    guest_name VARCHAR(100),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP,
    duration_seconds INTEGER
);

-- Enable RLS
ALTER TABLE stream_viewers ENABLE ROW LEVEL SECURITY;

-- Stream viewers inherit tenant access from streams
CREATE POLICY stream_viewers_isolation_policy ON stream_viewers
    USING (EXISTS (
        SELECT 1 FROM streams 
        WHERE streams.id = stream_viewers.stream_id 
        AND streams.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_stream_viewers_stream_id ON stream_viewers(stream_id);
CREATE INDEX idx_stream_viewers_person_id ON stream_viewers(person_id);

-- Stream notes/sermon notes (viewers can take notes during stream)
CREATE TABLE stream_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    person_id UUID REFERENCES people(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE stream_notes ENABLE ROW LEVEL SECURITY;

-- Stream notes inherit tenant access from streams
CREATE POLICY stream_notes_isolation_policy ON stream_notes
    USING (EXISTS (
        SELECT 1 FROM streams 
        WHERE streams.id = stream_notes.stream_id 
        AND streams.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_stream_notes_stream_id ON stream_notes(stream_id);
CREATE INDEX idx_stream_notes_person_id ON stream_notes(person_id);

-- Updated_at trigger
CREATE TRIGGER update_stream_notes_updated_at
    BEFORE UPDATE ON stream_notes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
