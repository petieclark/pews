-- ============================================================================
-- Pews Full Demo Seed Data
-- Populates demo-church with realistic data across ALL modules
-- Run AFTER the base seed-demo.sql and PCO import
-- Date context: February 2026
-- ============================================================================

BEGIN;

-- Store tenant_id for convenience
DO $$
DECLARE
    _tid UUID;
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';
    PERFORM set_config('app.current_tenant_id', _tid::text, true);
END $$;

-- ============================================================================
-- 1. FIX PEOPLE DUPLICATES & UPDATE MEMBERSHIP STATUSES
-- ============================================================================

-- Remove duplicates (keep earliest created)
DELETE FROM people WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY email ORDER BY created_at) as rn
        FROM people
        WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
        AND email IS NOT NULL
    ) t WHERE rn > 1
);

-- Set varied membership statuses
-- First reset all to active
UPDATE people SET membership_status = 'active'
WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

-- Now set specific statuses for variety
UPDATE people SET membership_status = 'member'
WHERE id IN (
    SELECT id FROM people
    WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
    ORDER BY created_at OFFSET 20 LIMIT 3
);

UPDATE people SET membership_status = 'visitor'
WHERE id IN (
    SELECT id FROM people
    WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
    ORDER BY created_at OFFSET 23 LIMIT 3
);

UPDATE people SET membership_status = 'inactive'
WHERE id IN (
    SELECT id FROM people
    WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
    ORDER BY created_at OFFSET 26 LIMIT 2
);

UPDATE people SET membership_status = 'regular'
WHERE id IN (
    SELECT id FROM people
    WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
    ORDER BY created_at OFFSET 28 LIMIT 2
);

-- ============================================================================
-- 2. SERVICE TYPES (ensure we have what we need)
-- ============================================================================

INSERT INTO service_types (tenant_id, name, default_time, default_day, color, is_active)
SELECT (SELECT id FROM tenants WHERE slug='demo-church'), name, default_time, default_day, color, true
FROM (VALUES
    ('Sunday Morning Service', '10:30 AM', 'sunday', '#4A8B8C'),
    ('Wednesday Night Service', '7:00 PM', 'wednesday', '#F59E0B'),
    ('Special Event', '6:00 PM', NULL, '#EF4444')
) AS v(name, default_time, default_day, color)
WHERE NOT EXISTS (
    SELECT 1 FROM service_types
    WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
    AND service_types.name = v.name
);

-- ============================================================================
-- 3. SERVICES (past 8 Sundays + next 4 + Wed + Special)
-- ============================================================================

-- Clean existing services to rebuild cleanly
DELETE FROM service_items WHERE service_id IN (
    SELECT id FROM services WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
);
DELETE FROM service_teams WHERE service_id IN (
    SELECT id FROM services WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
);
DELETE FROM services WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

-- Past 8 Sundays (completed)
INSERT INTO services (tenant_id, service_type_id, service_date, service_time, status)
SELECT
    (SELECT id FROM tenants WHERE slug='demo-church'),
    (SELECT id FROM service_types WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church') AND name = 'Sunday Morning Service' LIMIT 1),
    ('2026-02-15'::date - (w * 7))::date,
    '10:30 AM',
    'completed'
FROM generate_series(1, 8) AS w;

-- This Sunday (ready)
INSERT INTO services (tenant_id, service_type_id, service_date, service_time, status)
VALUES (
    (SELECT id FROM tenants WHERE slug='demo-church'),
    (SELECT id FROM service_types WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church') AND name = 'Sunday Morning Service' LIMIT 1),
    '2026-02-15',
    '10:30 AM',
    'ready'
);

-- Next 3 Sundays (planning/draft)
INSERT INTO services (tenant_id, service_type_id, service_date, service_time, status)
SELECT
    (SELECT id FROM tenants WHERE slug='demo-church'),
    (SELECT id FROM service_types WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church') AND name = 'Sunday Morning Service' LIMIT 1),
    ('2026-02-15'::date + (w * 7))::date,
    '10:30 AM',
    CASE WHEN w = 1 THEN 'planning' ELSE 'draft' END
FROM generate_series(1, 3) AS w;

-- 4 Wednesday nights (past 4 completed)
INSERT INTO services (tenant_id, service_type_id, service_date, service_time, status)
SELECT
    (SELECT id FROM tenants WHERE slug='demo-church'),
    (SELECT id FROM service_types WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church') AND name = 'Wednesday Night Service' LIMIT 1),
    ('2026-02-11'::date - (w * 7))::date,
    '7:00 PM',
    'completed'
FROM generate_series(0, 3) AS w;

-- Special Events: Christmas Eve (past) and Easter (upcoming)
INSERT INTO services (tenant_id, service_type_id, name, service_date, service_time, status)
VALUES
    (
        (SELECT id FROM tenants WHERE slug='demo-church'),
        (SELECT id FROM service_types WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church') AND name LIKE '%Special%' LIMIT 1),
        'Christmas Eve Candlelight Service',
        '2025-12-24',
        '6:00 PM',
        'completed'
    ),
    (
        (SELECT id FROM tenants WHERE slug='demo-church'),
        (SELECT id FROM service_types WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church') AND name LIKE '%Special%' LIMIT 1),
        'Easter Sunday Celebration',
        '2026-04-05',
        '10:30 AM',
        'planning'
    );

-- ============================================================================
-- 4. SERVICE ITEMS (order of service for each)
-- ============================================================================

-- Add items to all Sunday services using well-known worship songs
-- We'll pick songs that are likely in the PCO import
DO $$
DECLARE
    _tid UUID;
    _svc RECORD;
    _song_id UUID;
    _pos INTEGER;
    _song_titles TEXT[] := ARRAY[
        'Way Maker', 'Goodness of God', 'What a Beautiful Name',
        'Build My Life', 'Reckless Love', 'O Come to the Altar',
        'Great Are You Lord', 'Who You Say I Am', 'Living Hope',
        'Graves into Gardens', 'The Blessing', 'Holy Spirit',
        'Cornerstone', 'How Great Is Our God', 'Oceans (Where Feet May Fail)',
        'King of Kings', 'Battle Belongs', 'Do It Again',
        'Same God', 'Firm Foundation'
    ];
    _song_idx INTEGER;
    _svc_count INTEGER := 0;
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';

    FOR _svc IN
        SELECT s.id, s.service_date, st.name as type_name
        FROM services s
        JOIN service_types st ON s.service_type_id = st.id
        WHERE s.tenant_id = _tid
        ORDER BY s.service_date
    LOOP
        _svc_count := _svc_count + 1;
        _pos := 1;

        -- Welcome
        INSERT INTO service_items (service_id, item_type, title, position, duration_minutes, assigned_to)
        VALUES (_svc.id, 'other', 'Welcome & Announcements', _pos, 5, 'Pastor Mike');
        _pos := _pos + 1;

        -- Opening song
        _song_idx := ((_svc_count * 3) % array_length(_song_titles, 1)) + 1;
        SELECT id INTO _song_id FROM songs WHERE tenant_id = _tid AND title ILIKE '%' || _song_titles[_song_idx] || '%' LIMIT 1;
        IF _song_id IS NOT NULL THEN
            INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _song_id, 'G', _pos, 5);
        ELSE
            INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _pos, 5);
        END IF;
        _pos := _pos + 1;

        -- Second song
        _song_idx := ((_svc_count * 3 + 1) % array_length(_song_titles, 1)) + 1;
        SELECT id INTO _song_id FROM songs WHERE tenant_id = _tid AND title ILIKE '%' || _song_titles[_song_idx] || '%' LIMIT 1;
        IF _song_id IS NOT NULL THEN
            INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _song_id, 'C', _pos, 5);
        ELSE
            INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _pos, 5);
        END IF;
        _pos := _pos + 1;

        -- Prayer
        INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
        VALUES (_svc.id, 'prayer', 'Pastoral Prayer', _pos, 3);
        _pos := _pos + 1;

        -- Third song
        _song_idx := ((_svc_count * 3 + 2) % array_length(_song_titles, 1)) + 1;
        SELECT id INTO _song_id FROM songs WHERE tenant_id = _tid AND title ILIKE '%' || _song_titles[_song_idx] || '%' LIMIT 1;
        IF _song_id IS NOT NULL THEN
            INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _song_id, 'D', _pos, 5);
        ELSE
            INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _pos, 5);
        END IF;
        _pos := _pos + 1;

        -- Offering
        INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
        VALUES (_svc.id, 'other', 'Tithes & Offerings', _pos, 5);
        _pos := _pos + 1;

        -- Sermon
        INSERT INTO service_items (service_id, item_type, title, position, duration_minutes, assigned_to, notes)
        VALUES (_svc.id, 'sermon', 
            CASE 
                WHEN _svc.type_name LIKE '%Wednesday%' THEN 'Bible Study: Book of James'
                WHEN _svc.service_date = '2025-12-24' THEN 'The Light of the World'
                ELSE 'Sunday Message'
            END,
            _pos, 35, 'Pastor Mike',
            CASE
                WHEN _svc.service_date = '2025-12-24' THEN 'Luke 2:1-20 - Christmas Eve message'
                ELSE NULL
            END
        );
        _pos := _pos + 1;

        -- Closing song
        _song_idx := ((_svc_count * 3 + 5) % array_length(_song_titles, 1)) + 1;
        SELECT id INTO _song_id FROM songs WHERE tenant_id = _tid AND title ILIKE '%' || _song_titles[_song_idx] || '%' LIMIT 1;
        IF _song_id IS NOT NULL THEN
            INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _song_id, 'G', _pos, 5);
        ELSE
            INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
            VALUES (_svc.id, 'song', _song_titles[_song_idx], _pos, 5);
        END IF;
        _pos := _pos + 1;

        -- Benediction
        INSERT INTO service_items (service_id, item_type, title, position, duration_minutes)
        VALUES (_svc.id, 'other', 'Benediction & Dismissal', _pos, 2);
    END LOOP;
END $$;

-- ============================================================================
-- 5. GIVING (Funds + 100+ Donations)
-- ============================================================================

-- Clean existing giving data
DELETE FROM donations WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');
DELETE FROM funds WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

-- Create funds
INSERT INTO funds (tenant_id, name, description, is_default, is_active)
VALUES
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'General Fund', 'Tithes and general offerings supporting church operations', true, true),
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'Missions', 'Supporting local and global mission partners', false, true),
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'Building Fund', 'New sanctuary construction and facility improvements', false, true);

-- Generate donations over past 3 months
DO $$
DECLARE
    _tid UUID;
    _person RECORD;
    _fund_id UUID;
    _fund_ids UUID[];
    _amounts INTEGER[] := ARRAY[2500, 5000, 5000, 10000, 10000, 10000, 25000, 25000, 50000, 100000];
    _methods TEXT[] := ARRAY['online', 'online', 'online', 'card', 'card', 'check', 'check', 'cash'];
    _week INTEGER;
    _donation_date TIMESTAMP;
    _amount INTEGER;
    _method TEXT;
    _person_count INTEGER := 0;
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';
    SELECT array_agg(id ORDER BY name) INTO _fund_ids FROM funds WHERE tenant_id = _tid;

    -- Recurring donors: first 15 people give weekly to General Fund
    FOR _person IN
        SELECT id FROM people WHERE tenant_id = _tid AND email IS NOT NULL ORDER BY created_at LIMIT 15
    LOOP
        _person_count := _person_count + 1;
        -- Each recurring donor has a consistent amount
        _amount := _amounts[(_person_count % array_length(_amounts, 1)) + 1];

        FOR _week IN 0..11 LOOP
            _donation_date := ('2026-02-11'::date - (_week * 7) + (random() * 2)::int)::timestamp
                            + (interval '8 hours') + (random() * interval '4 hours');

            -- Skip some weeks randomly (80% give rate)
            IF random() < 0.8 THEN
                INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, is_recurring, recurring_frequency, donated_at, created_at)
                VALUES (_tid, _person.id, _fund_ids[1], _amount, 'online', 'completed', true, 'weekly', _donation_date, _donation_date);
            END IF;
        END LOOP;
    END LOOP;

    -- One-time donations from various people to various funds
    FOR _person IN
        SELECT id FROM people WHERE tenant_id = _tid AND email IS NOT NULL ORDER BY random() LIMIT 25
    LOOP
        -- 1-3 one-time donations per person
        FOR _week IN 1..(1 + (random() * 2)::int) LOOP
            _donation_date := ('2026-02-11'::date - (random() * 90)::int)::timestamp
                            + (interval '9 hours') + (random() * interval '6 hours');
            _amount := _amounts[(random() * (array_length(_amounts, 1) - 1))::int + 1];
            _method := _methods[(random() * (array_length(_methods, 1) - 1))::int + 1];
            _fund_id := _fund_ids[(random() * (array_length(_fund_ids, 1) - 1))::int + 1];

            INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at, created_at)
            VALUES (_tid, _person.id, _fund_id, _amount, _method, 'completed', _donation_date, _donation_date);
        END LOOP;
    END LOOP;

    -- A few large building fund donations
    FOR _person IN
        SELECT id FROM people WHERE tenant_id = _tid AND email IS NOT NULL ORDER BY created_at LIMIT 5
    LOOP
        _donation_date := ('2026-02-11'::date - (random() * 60)::int)::timestamp + interval '10 hours';
        INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, memo, donated_at, created_at)
        VALUES (_tid, _person.id, (SELECT id FROM funds WHERE tenant_id = _tid AND name = 'Building Fund'),
                CASE WHEN random() < 0.5 THEN 50000 ELSE 100000 END,
                'check', 'completed', 'Building fund pledge payment', _donation_date, _donation_date);
    END LOOP;
END $$;

-- ============================================================================
-- 6. GROUPS
-- ============================================================================

DELETE FROM group_members WHERE group_id IN (SELECT id FROM groups WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church'));
DELETE FROM groups WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

INSERT INTO groups (tenant_id, name, description, group_type, meeting_day, meeting_time, meeting_location, is_public, is_active)
VALUES
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'Men''s Group', 'Iron sharpens iron — men growing together in faith and accountability', 'small_group', 'saturday', '7:00 AM', 'Fellowship Hall', true, true),
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'Women''s Bible Study', 'Deep dive into Scripture with fellowship and prayer', 'bible_study', 'tuesday', '10:00 AM', 'Room 201', true, true),
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'Young Adults', 'Community for ages 18-30 — faith, fun, and real talk', 'small_group', 'thursday', '7:00 PM', 'Youth Center', true, true),
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'Marriage Group', 'Strengthening marriages through biblical principles', 'small_group', 'friday', '6:30 PM', 'Room 105', true, true),
    ((SELECT id FROM tenants WHERE slug='demo-church'), 'New Members Class', '4-week course covering our beliefs, mission, and how to get connected', 'class', 'sunday', '12:00 PM', 'Conference Room', true, true);

-- Assign members to groups
DO $$
DECLARE
    _tid UUID;
    _group RECORD;
    _person_id UUID;
    _idx INTEGER;
    _roles TEXT[] := ARRAY['leader', 'co_leader', 'member', 'member', 'member', 'member', 'member', 'member'];
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';

    FOR _group IN SELECT id, name FROM groups WHERE tenant_id = _tid LOOP
        _idx := 0;
        FOR _person_id IN
            SELECT p.id FROM people p
            WHERE p.tenant_id = _tid AND p.email IS NOT NULL
            ORDER BY md5(_group.name || p.email) -- deterministic pseudo-random per group
            LIMIT 4 + (length(_group.name) % 5) -- 4-8 members
        LOOP
            _idx := _idx + 1;
            INSERT INTO group_members (group_id, person_id, role, joined_at)
            VALUES (_group.id, _person_id, _roles[LEAST(_idx, array_length(_roles, 1))],
                    NOW() - (random() * interval '90 days'))
            ON CONFLICT (group_id, person_id) DO NOTHING;
        END LOOP;
    END LOOP;
END $$;

-- ============================================================================
-- 7. CALENDAR EVENTS
-- ============================================================================

DELETE FROM events WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

DO $$
DECLARE
    _tid UUID;
    _user_id UUID;
    _w INTEGER;
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';
    SELECT id INTO _user_id FROM users WHERE tenant_id = _tid LIMIT 1;

    -- Weekly Sunday services (past 8 + next 4)
    FOR _w IN -8..4 LOOP
        INSERT INTO events (tenant_id, title, description, location, start_time, end_time, recurring, color, event_type, created_by)
        VALUES (_tid, 'Sunday Morning Worship', 'Join us for worship, prayer, and the Word',
                'Main Sanctuary',
                ('2026-02-15'::date + (_w * 7))::timestamp + interval '10 hours 30 minutes',
                ('2026-02-15'::date + (_w * 7))::timestamp + interval '12 hours',
                'weekly', '#4A8B8C', 'service', _user_id);
    END LOOP;

    -- Weekly Wednesday nights (past 4 + next 4)
    FOR _w IN -4..3 LOOP
        INSERT INTO events (tenant_id, title, description, location, start_time, end_time, recurring, color, event_type, created_by)
        VALUES (_tid, 'Wednesday Night Bible Study', 'Midweek study through the book of James',
                'Fellowship Hall',
                ('2026-02-11'::date + (_w * 7))::timestamp + interval '19 hours',
                ('2026-02-11'::date + (_w * 7))::timestamp + interval '20 hours 30 minutes',
                'weekly', '#F59E0B', 'service', _user_id);
    END LOOP;

    -- Monthly events
    INSERT INTO events (tenant_id, title, description, location, start_time, end_time, recurring, color, event_type, created_by) VALUES
        (_tid, 'Board Meeting', 'Monthly elder/deacon board meeting', 'Conference Room',
         '2026-02-03 18:30:00', '2026-02-03 20:00:00', 'monthly', '#6B7280', 'meeting', _user_id),
        (_tid, 'Board Meeting', 'Monthly elder/deacon board meeting', 'Conference Room',
         '2026-03-03 18:30:00', '2026-03-03 20:00:00', 'monthly', '#6B7280', 'meeting', _user_id),
        (_tid, 'Women''s Brunch', 'Monthly women''s fellowship brunch', 'Fellowship Hall',
         '2026-01-17 10:00:00', '2026-01-17 12:00:00', 'monthly', '#EC4899', 'fellowship', _user_id),
        (_tid, 'Women''s Brunch', 'Monthly women''s fellowship brunch', 'Fellowship Hall',
         '2026-02-21 10:00:00', '2026-02-21 12:00:00', 'monthly', '#EC4899', 'fellowship', _user_id),
        (_tid, 'Men''s Breakfast', 'Monthly men''s breakfast and devotional', 'Fellowship Hall',
         '2026-01-10 07:00:00', '2026-01-10 09:00:00', 'monthly', '#3B82F6', 'fellowship', _user_id),
        (_tid, 'Men''s Breakfast', 'Monthly men''s breakfast and devotional', 'Fellowship Hall',
         '2026-02-14 07:00:00', '2026-02-14 09:00:00', 'monthly', '#3B82F6', 'fellowship', _user_id);

    -- Upcoming special events
    INSERT INTO events (tenant_id, title, description, location, start_time, end_time, color, event_type, created_by) VALUES
        (_tid, 'Easter Sunday Celebration', 'Resurrection Sunday — invite your neighbors!', 'Main Sanctuary',
         '2026-04-05 10:30:00', '2026-04-05 12:30:00', '#EF4444', 'service', _user_id),
        (_tid, 'Youth Lock-In', 'Overnight fun for grades 6-12. Games, worship, and pizza!', 'Youth Center',
         '2026-03-13 19:00:00', '2026-03-14 08:00:00', '#8B5CF6', 'fellowship', _user_id),
        (_tid, 'VBS Planning Meeting', 'Kickoff meeting for Vacation Bible School 2026', 'Conference Room',
         '2026-02-24 18:30:00', '2026-02-24 20:00:00', '#10B981', 'meeting', _user_id);
END $$;

-- ============================================================================
-- 8. CHECK-INS (past 8 Sunday services)
-- ============================================================================

DELETE FROM checkins WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');
DELETE FROM checkin_events WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

-- Create a check-in station if none exists
INSERT INTO checkin_stations (tenant_id, name, location, is_active)
SELECT (SELECT id FROM tenants WHERE slug='demo-church'), 'Main Lobby', 'Front Entrance', true
WHERE NOT EXISTS (
    SELECT 1 FROM checkin_stations WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
);

-- Create check-in events and records for past 8 Sundays
DO $$
DECLARE
    _tid UUID;
    _station_id UUID;
    _svc RECORD;
    _ce_id UUID;
    _person_id UUID;
    _target_count INTEGER;
    _actual INTEGER;
    _svc_num INTEGER := 0;
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';
    SELECT id INTO _station_id FROM checkin_stations WHERE tenant_id = _tid LIMIT 1;

    FOR _svc IN
        SELECT s.id, s.service_date
        FROM services s
        JOIN service_types st ON s.service_type_id = st.id
        WHERE s.tenant_id = _tid AND st.name LIKE '%Sunday%' AND s.status = 'completed'
        ORDER BY s.service_date
    LOOP
        _svc_num := _svc_num + 1;
        -- Trending upward: 15 to 25
        _target_count := 14 + _svc_num + (random() * 3)::int;
        IF _target_count > 28 THEN _target_count := 28; END IF;

        -- Create check-in event
        INSERT INTO checkin_events (tenant_id, name, event_date, service_id, station_id, is_active)
        VALUES (_tid, 'Sunday Morning - ' || to_char(_svc.service_date, 'Mon DD'), _svc.service_date, _svc.id, _station_id, true)
        RETURNING id INTO _ce_id;

        -- Add check-ins from random people
        _actual := 0;
        FOR _person_id IN
            SELECT id FROM people WHERE tenant_id = _tid
            ORDER BY md5(id::text || _svc.service_date::text)
            LIMIT _target_count
        LOOP
            _actual := _actual + 1;
            INSERT INTO checkins (tenant_id, event_id, person_id, station_id, first_time, checked_in_at)
            VALUES (_tid, _ce_id, _person_id, _station_id,
                    _actual > (_target_count - 2) AND random() < 0.3, -- occasional first-timers
                    _svc.service_date::timestamp + interval '10 hours' + (random() * interval '30 minutes'));
        END LOOP;
    END LOOP;
END $$;

-- ============================================================================
-- 9. FOLLOW-UPS / CARE
-- ============================================================================

DELETE FROM follow_up_notes WHERE follow_up_id IN (
    SELECT id FROM follow_ups WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church')
);
DELETE FROM follow_ups WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

DO $$
DECLARE
    _tid UUID;
    _user_id UUID;
    _people_ids UUID[];
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';
    SELECT id INTO _user_id FROM users WHERE tenant_id = _tid LIMIT 1;
    SELECT array_agg(id ORDER BY created_at) INTO _people_ids
    FROM people WHERE tenant_id = _tid AND email IS NOT NULL LIMIT 10;

    INSERT INTO follow_ups (tenant_id, person_id, assigned_to, title, type, priority, status, due_date, completed_at, created_at) VALUES
        (_tid, _people_ids[1], _user_id, 'First visit follow-up — Sarah visited Jan 26', 'first_time_visitor', 'high', 'completed', '2026-01-28', '2026-01-29 14:00:00', '2026-01-26 12:00:00'),
        (_tid, _people_ids[2], _user_id, 'First visit follow-up — New family from community', 'first_time_visitor', 'high', 'in_progress', '2026-02-12', NULL, '2026-02-08 12:00:00'),
        (_tid, _people_ids[3], _user_id, 'Hospital visit — recovering from surgery', 'hospital_visit', 'high', 'completed', '2026-01-20', '2026-01-21 16:00:00', '2026-01-18 09:00:00'),
        (_tid, _people_ids[4], _user_id, 'Hospital visit — new baby congratulations', 'hospital_visit', 'medium', 'completed', '2026-02-05', '2026-02-06 11:00:00', '2026-02-03 10:00:00'),
        (_tid, _people_ids[5], _user_id, 'Pre-marriage counseling request', 'counseling', 'medium', 'in_progress', '2026-02-20', NULL, '2026-02-01 14:00:00'),
        (_tid, _people_ids[6], _user_id, 'Membership class interest — ready to join', 'membership', 'medium', 'waiting', '2026-02-18', NULL, '2026-02-05 09:00:00'),
        (_tid, _people_ids[7], _user_id, 'Check in — hasn''t attended in 3 weeks', 'general', 'low', 'new', '2026-02-15', NULL, '2026-02-10 10:00:00'),
        (_tid, _people_ids[8], _user_id, 'Grief support — lost a family member', 'general', 'high', 'new', '2026-02-13', NULL, '2026-02-10 15:00:00');
END $$;

-- ============================================================================
-- 10. PRAYER REQUESTS
-- ============================================================================

DELETE FROM prayer_requests WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

DO $$
DECLARE
    _tid UUID;
    _people_ids UUID[];
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';
    SELECT array_agg(id ORDER BY created_at) INTO _people_ids
    FROM people WHERE tenant_id = _tid AND email IS NOT NULL LIMIT 12;

    INSERT INTO prayer_requests (tenant_id, person_id, name, request_text, is_public, status, submitted_at) VALUES
        (_tid, _people_ids[1], 'Michael', 'Please pray for my mother — she is having heart surgery next week. We trust God but would love the church family lifting her up.', true, 'praying', '2026-02-08 09:00:00'),
        (_tid, _people_ids[2], 'Sarah', 'Praying for wisdom as our family considers a job relocation. Want to be where God wants us.', true, 'pending', '2026-02-10 14:00:00'),
        (_tid, _people_ids[3], 'David', 'Praise report! After months of treatment, my scans came back clear. God is faithful!', true, 'answered', '2026-01-25 11:00:00'),
        (_tid, _people_ids[4], 'Jennifer', 'Our son is deploying overseas next month. Praying for his safety and for peace for our family.', true, 'praying', '2026-02-03 16:00:00'),
        (_tid, _people_ids[5], 'Rachel', 'Pray for our missions team heading to Guatemala in March. Pray for safe travel and open hearts.', true, 'praying', '2026-02-01 10:00:00'),
        (_tid, _people_ids[6], 'James', 'Going through a career transition. Trusting God for the next door to open.', true, 'pending', '2026-02-09 08:00:00'),
        (_tid, _people_ids[7], 'Amanda', 'Please pray for unity in our community — there''s been a lot of division and we need God''s peace.', true, 'praying', '2026-01-28 12:00:00'),
        (_tid, _people_ids[8], 'Chris', 'Answered prayer! After 2 years of trying, we are expecting a baby! God is so good.', true, 'answered', '2026-02-05 09:00:00'),
        (_tid, _people_ids[9], 'Daniel', 'Struggling with anxiety. Asking for prayer for peace and for the right counselor.', false, 'praying', '2026-02-07 20:00:00'),
        (_tid, _people_ids[10], 'Rebecca', 'Praying for our church — for continued growth and for God to send us the right volunteers for VBS.', true, 'pending', '2026-02-11 07:00:00');
END $$;

-- ============================================================================
-- 11. CAMPAIGNS / COMMUNICATIONS
-- ============================================================================

DELETE FROM campaign_recipients WHERE campaign_id IN (SELECT id FROM campaigns WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church'));
DELETE FROM campaigns WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

DO $$
DECLARE
    _tid UUID;
    _c1 UUID;
    _c2 UUID;
    _c3 UUID;
    _person RECORD;
BEGIN
    SELECT id INTO _tid FROM tenants WHERE slug = 'demo-church';

    -- Campaign 1: Welcome Series (sent)
    INSERT INTO campaigns (tenant_id, name, channel, subject, body, status, sent_at, recipient_count, opened_count, clicked_count, target_type, created_at)
    VALUES (_tid, 'Welcome to Grace Community!', 'email',
            'Welcome to Grace Community Church!',
            E'Hi {{first_name}},\n\nThank you for visiting Grace Community Church! We''re so glad you joined us.\n\nHere are a few ways to get connected:\n- Join a small group\n- Sign up to volunteer\n- Download our app\n\nWe''d love to get to know you better. Reply to this email or stop by the Welcome Center this Sunday!\n\nBlessings,\nPastor Mike',
            'sent', '2026-01-20 09:00:00', 8, 6, 3, 'tag', '2026-01-19 14:00:00')
    RETURNING id INTO _c1;

    -- Campaign 2: Easter Invite (sent)
    INSERT INTO campaigns (tenant_id, name, channel, subject, body, status, sent_at, recipient_count, opened_count, clicked_count, target_type, created_at)
    VALUES (_tid, 'Easter at Grace Community', 'email',
            'You''re Invited: Easter Sunday at Grace Community',
            E'Hi {{first_name}},\n\nEaster is coming on April 5th, and we''d love for you to join us!\n\n🌅 Easter Sunday Celebration\n📍 Main Sanctuary\n⏰ 10:30 AM\n\nBring your family, invite a friend, and experience the joy of Resurrection Sunday together.\n\nSpecial music, kids program, and a message of hope.\n\nSee you there!\nGrace Community Church',
            'sent', '2026-02-05 10:00:00', 45, 32, 12, 'all', '2026-02-03 16:00:00')
    RETURNING id INTO _c2;

    -- Campaign 3: Volunteer Signup (sent)
    INSERT INTO campaigns (tenant_id, name, channel, subject, body, status, sent_at, recipient_count, opened_count, clicked_count, target_type, created_at)
    VALUES (_tid, 'Volunteer Opportunities This Spring', 'email',
            'We Need You! Spring Volunteer Opportunities',
            E'Hi {{first_name}},\n\nSpring is a busy season at Grace Community and we need YOUR help!\n\nOpen positions:\n🎵 Worship Team — singers and musicians\n👋 Hospitality — greeters and coffee bar\n📚 Kids Ministry — Sunday morning teachers\n🔧 Tech Team — sound and media\n\nNo experience needed — just a willing heart!\n\nSign up at the Welcome Center or reply to this email.\n\nThanks for serving,\nPastor Mike',
            'sent', '2026-02-08 09:00:00', 38, 25, 8, 'all', '2026-02-07 11:00:00')
    RETURNING id INTO _c3;

    -- Add recipients for campaigns
    FOR _person IN
        SELECT id FROM people WHERE tenant_id = _tid AND email IS NOT NULL ORDER BY random() LIMIT 8
    LOOP
        INSERT INTO campaign_recipients (campaign_id, person_id, status, sent_at, opened_at)
        VALUES (_c1, _person.id,
                CASE WHEN random() < 0.75 THEN 'opened' ELSE 'sent' END,
                '2026-01-20 09:00:00',
                CASE WHEN random() < 0.75 THEN '2026-01-20 09:00:00'::timestamp + (random() * interval '48 hours') ELSE NULL END);
    END LOOP;

    FOR _person IN
        SELECT id FROM people WHERE tenant_id = _tid AND email IS NOT NULL ORDER BY random() LIMIT 40
    LOOP
        INSERT INTO campaign_recipients (campaign_id, person_id, status, sent_at, opened_at)
        VALUES (_c2, _person.id,
                CASE WHEN random() < 0.7 THEN 'opened' WHEN random() < 0.9 THEN 'sent' ELSE 'delivered' END,
                '2026-02-05 10:00:00',
                CASE WHEN random() < 0.7 THEN '2026-02-05 10:00:00'::timestamp + (random() * interval '72 hours') ELSE NULL END);
    END LOOP;

    FOR _person IN
        SELECT id FROM people WHERE tenant_id = _tid AND email IS NOT NULL ORDER BY random() LIMIT 35
    LOOP
        INSERT INTO campaign_recipients (campaign_id, person_id, status, sent_at, opened_at)
        VALUES (_c3, _person.id,
                CASE WHEN random() < 0.65 THEN 'opened' ELSE 'sent' END,
                '2026-02-08 09:00:00',
                CASE WHEN random() < 0.65 THEN '2026-02-08 09:00:00'::timestamp + (random() * interval '48 hours') ELSE NULL END);
    END LOOP;
END $$;

-- ============================================================================
-- 12. ENGAGEMENT SCORES (for dashboard)
-- ============================================================================

DELETE FROM engagement_scores WHERE tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

INSERT INTO engagement_scores (tenant_id, person_id, score, attendance_score, giving_score, group_score, volunteer_score, connection_score)
SELECT
    (SELECT id FROM tenants WHERE slug='demo-church'),
    p.id,
    -- Overall score: weighted mix
    LEAST(100, GREATEST(10,
        (30 + (random() * 70))::int
    )),
    (random() * 100)::int,
    (random() * 100)::int,
    (random() * 100)::int,
    (random() * 100)::int,
    (random() * 100)::int
FROM people p
WHERE p.tenant_id = (SELECT id FROM tenants WHERE slug='demo-church');

COMMIT;
