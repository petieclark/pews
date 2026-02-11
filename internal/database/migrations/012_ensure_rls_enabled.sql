-- Ensure RLS is enabled on all tenant-scoped tables
-- This migration is idempotent - it will enable RLS if not already enabled

-- Core tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants') THEN ALTER TABLE tenants ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN ALTER TABLE users ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenant_modules') THEN ALTER TABLE tenant_modules ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'subscriptions') THEN ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- People tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'people') THEN ALTER TABLE people ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'households') THEN ALTER TABLE households ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'household_members') THEN ALTER TABLE household_members ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tags') THEN ALTER TABLE tags ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'person_tags') THEN ALTER TABLE person_tags ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Giving tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'funds') THEN ALTER TABLE funds ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'donations') THEN ALTER TABLE donations ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'donation_items') THEN ALTER TABLE donation_items ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'recurring_donations') THEN ALTER TABLE recurring_donations ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Groups tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'groups') THEN ALTER TABLE groups ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'group_members') THEN ALTER TABLE group_members ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Services tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'service_types') THEN ALTER TABLE service_types ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'services') THEN ALTER TABLE services ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'service_items') THEN ALTER TABLE service_items ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'service_team') THEN ALTER TABLE service_team ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'songs') THEN ALTER TABLE songs ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Check-ins tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'checkin_events') THEN ALTER TABLE checkin_events ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'checkin_stations') THEN ALTER TABLE checkin_stations ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'checkin_records') THEN ALTER TABLE checkin_records ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'checkin_alerts') THEN ALTER TABLE checkin_alerts ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'authorized_pickups') THEN ALTER TABLE authorized_pickups ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Communication tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'templates') THEN ALTER TABLE templates ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'campaigns') THEN ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'campaign_recipients') THEN ALTER TABLE campaign_recipients ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'journeys') THEN ALTER TABLE journeys ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'journey_steps') THEN ALTER TABLE journey_steps ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'journey_enrollments') THEN ALTER TABLE journey_enrollments ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'connection_cards') THEN ALTER TABLE connection_cards ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Streaming tables
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'streams') THEN ALTER TABLE streams ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stream_chat') THEN ALTER TABLE stream_chat ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stream_viewers') THEN ALTER TABLE stream_viewers ENABLE ROW LEVEL SECURITY; END IF; END $$;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stream_notes') THEN ALTER TABLE stream_notes ENABLE ROW LEVEL SECURITY; END IF; END $$;

-- Note: This migration assumes all RLS policies are already created in their respective
-- table creation migrations (001-011). This migration only ensures RLS is ENABLED.
-- If RLS was previously disabled for testing/demo purposes, this re-enables it.
