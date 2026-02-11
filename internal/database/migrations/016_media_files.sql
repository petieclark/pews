-- Media Files Table
CREATE TABLE IF NOT EXISTS media_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    url TEXT NOT NULL,
    folder VARCHAR(255) DEFAULT '',
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    tags JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_media_files_tenant ON media_files(tenant_id);
CREATE INDEX idx_media_files_folder ON media_files(tenant_id, folder);
CREATE INDEX idx_media_files_content_type ON media_files(tenant_id, content_type);
CREATE INDEX idx_media_files_tags ON media_files USING gin(tags);
CREATE INDEX idx_media_files_created_at ON media_files(tenant_id, created_at DESC);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_media_files_updated_at
    BEFORE UPDATE ON media_files
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
