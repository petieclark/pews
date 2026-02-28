-- Add key field to service_plan_items for song transposition display
ALTER TABLE service_plan_items ADD COLUMN IF NOT EXISTS key VARCHAR(10);

-- Create index for queries filtering by key
CREATE INDEX IF NOT EXISTS idx_service_plan_items_key ON service_plan_items(key) WHERE key IS NOT NULL;
