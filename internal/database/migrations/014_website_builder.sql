-- Add settings column to tenants table for website builder config
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}';

-- Index for faster JSON queries
CREATE INDEX IF NOT EXISTS idx_tenants_settings ON tenants USING GIN (settings);
