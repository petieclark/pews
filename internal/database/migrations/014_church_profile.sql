-- Add church profile fields to tenants table
-- Note: address_line1, address_line2, city, state, zip, ein already exist from previous migrations
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS website VARCHAR(255);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS logo TEXT; -- base64 encoded logo
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS about TEXT;

-- Index on email for quick lookups
CREATE INDEX IF NOT EXISTS idx_tenants_email ON tenants(email) WHERE email IS NOT NULL;
