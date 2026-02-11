-- Song attachments table for chord charts and sheet music PDFs
CREATE TABLE IF NOT EXISTS song_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL DEFAULT 'application/pdf',
    file_data BYTEA NOT NULL,
    file_size INTEGER NOT NULL,
    uploaded_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_song_attachments_song ON song_attachments(song_id);
CREATE INDEX idx_song_attachments_tenant ON song_attachments(tenant_id);
