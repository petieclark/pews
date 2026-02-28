-- Add share_token column to service_plans for public access
ALTER TABLE service_plans ADD COLUMN IF NOT EXISTS share_token UUID;

-- Create index for faster token lookups
CREATE INDEX IF NOT EXISTS idx_service_plans_share_token ON service_plans(share_token);
