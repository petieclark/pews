-- Add event_type and room_id to events
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_type VARCHAR(50) DEFAULT 'other';
ALTER TABLE events ADD COLUMN IF NOT EXISTS room_id UUID REFERENCES rooms(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_events_type ON events(tenant_id, event_type);
CREATE INDEX IF NOT EXISTS idx_events_room ON events(room_id);
