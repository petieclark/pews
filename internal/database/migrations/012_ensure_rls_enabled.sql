-- Ensure RLS is enabled on all tenant-scoped tables
-- This migration is idempotent - it will enable RLS if not already enabled

-- Core tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;

-- People tables
ALTER TABLE people ENABLE ROW LEVEL SECURITY;
ALTER TABLE households ENABLE ROW LEVEL SECURITY;
ALTER TABLE household_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE person_tags ENABLE ROW LEVEL SECURITY;

-- Giving tables
ALTER TABLE funds ENABLE ROW LEVEL SECURITY;
ALTER TABLE donations ENABLE ROW LEVEL SECURITY;
ALTER TABLE donation_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE recurring_donations ENABLE ROW LEVEL SECURITY;

-- Groups tables
ALTER TABLE groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE group_members ENABLE ROW LEVEL SECURITY;

-- Services tables
ALTER TABLE service_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE services ENABLE ROW LEVEL SECURITY;
ALTER TABLE service_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE service_team ENABLE ROW LEVEL SECURITY;
ALTER TABLE songs ENABLE ROW LEVEL SECURITY;

-- Check-ins tables
ALTER TABLE checkin_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE checkin_stations ENABLE ROW LEVEL SECURITY;
ALTER TABLE checkin_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE checkin_alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE authorized_pickups ENABLE ROW LEVEL SECURITY;

-- Communication tables
ALTER TABLE templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaign_recipients ENABLE ROW LEVEL SECURITY;
ALTER TABLE journeys ENABLE ROW LEVEL SECURITY;
ALTER TABLE journey_steps ENABLE ROW LEVEL SECURITY;
ALTER TABLE journey_enrollments ENABLE ROW LEVEL SECURITY;
ALTER TABLE connection_cards ENABLE ROW LEVEL SECURITY;

-- Streaming tables
ALTER TABLE streams ENABLE ROW LEVEL SECURITY;
ALTER TABLE stream_chat ENABLE ROW LEVEL SECURITY;
ALTER TABLE stream_viewers ENABLE ROW LEVEL SECURITY;
ALTER TABLE stream_notes ENABLE ROW LEVEL SECURITY;

-- Note: This migration assumes all RLS policies are already created in their respective
-- table creation migrations (001-011). This migration only ensures RLS is ENABLED.
-- If RLS was previously disabled for testing/demo purposes, this re-enables it.
