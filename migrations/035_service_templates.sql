CREATE TABLE IF NOT EXISTS service_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_service_templates_tenant ON service_templates(tenant_id);

-- Add default_day and default_time to service_types if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'service_types' AND column_name = 'default_day') THEN
        ALTER TABLE service_types ADD COLUMN default_day VARCHAR(20);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'service_types' AND column_name = 'default_time') THEN
        ALTER TABLE service_types ADD COLUMN default_time VARCHAR(20);
    END IF;
END$$;
