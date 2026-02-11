-- Add default_locale to tenants table
ALTER TABLE tenants ADD COLUMN default_locale VARCHAR(5) DEFAULT 'en' NOT NULL;

-- Add comment explaining supported locales
COMMENT ON COLUMN tenants.default_locale IS 'Default locale for tenant. Supported: en, es, pt, ko';

-- Index for faster locale lookups
CREATE INDEX idx_tenants_default_locale ON tenants(default_locale);
