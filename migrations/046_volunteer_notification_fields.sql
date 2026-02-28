-- Add notification tracking fields to service_team_assignments
ALTER TABLE service_team_assignments 
ADD COLUMN IF NOT EXISTS notification_sent BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS notified_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS responded_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS token_generated_at TIMESTAMP;

-- Create index for efficient notification queries
CREATE INDEX IF NOT EXISTS idx_sta_notification_pending 
ON service_team_assignments(service_id, status) 
WHERE notification_sent = false AND status IN ('pending', 'confirmed');
