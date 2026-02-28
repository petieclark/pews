-- Add key field to service_plan_items for song transposition
-- Issue #75

ALTER TABLE service_plan_items ADD COLUMN IF NOT EXISTS key VARCHAR(20);

-- Index for filtering by key if needed
CREATE INDEX IF NOT EXISTS idx_service_plan_items_key ON service_plan_items(key) WHERE key IS NOT NULL;
