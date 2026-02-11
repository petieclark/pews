-- Pews Demo Seed Data Script
-- Creates realistic demo data for Grace Community Church (demo.pews.app)
-- Safe to run multiple times (idempotent)

-- Disable RLS temporarily for seeding
SET app.current_tenant_id = '';

BEGIN;

-- ============================================================================
-- TENANT
-- ============================================================================

INSERT INTO tenants (id, name, slug, plan)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Grace Community Church',
    'demo-church',
    'premium'
)
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name, plan = EXCLUDED.plan;

-- Set tenant context for the rest of the script
SET app.current_tenant_id = '00000000-0000-0000-0000-000000000001';

-- ============================================================================
-- TAGS
-- ============================================================================

INSERT INTO tags (id, tenant_id, name, color)
VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Volunteer', '#4A8B8C'),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Youth', '#F59E0B'),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Worship Team', '#8B5CF6'),
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Small Group Leader', '#10B981'),
    ('10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'New Member', '#3B82F6')
ON CONFLICT (tenant_id, name) DO NOTHING;

-- ============================================================================
-- PEOPLE & HOUSEHOLDS
-- ============================================================================

-- Household 1: The Johnsons
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Michael', 'Johnson', 'michael.johnson@email.com', '555-0101', '123 Oak Street', 'Springfield', 'IL', '62701', '1978-03-15', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Sarah', 'Johnson', 'sarah.johnson@email.com', '555-0102', '123 Oak Street', 'Springfield', 'IL', '62701', '1980-07-22', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Emily', 'Johnson', 'emily.johnson@email.com', '555-0103', '123 Oak Street', 'Springfield', 'IL', '62701', '2008-11-05', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Joshua', 'Johnson', NULL, NULL, '123 Oak Street', 'Springfield', 'IL', '62701', '2012-05-18', 'male', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Johnson Family', '20000000-0000-0000-0000-000000000001', '123 Oak Street', 'Springfield', 'IL', '62701')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'head'),
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002', 'spouse'),
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000003', 'child'),
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000004', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 2: The Smiths
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'David', 'Smith', 'david.smith@email.com', '555-0201', '456 Maple Ave', 'Springfield', 'IL', '62702', '1985-09-10', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', 'Jessica', 'Smith', 'jessica.smith@email.com', '555-0202', '456 Maple Ave', 'Springfield', 'IL', '62702', '1987-02-14', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', 'Olivia', 'Smith', NULL, NULL, '456 Maple Ave', 'Springfield', 'IL', '62702', '2014-08-20', 'female', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Smith Family', '20000000-0000-0000-0000-000000000005', '456 Maple Ave', 'Springfield', 'IL', '62702')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000005', 'head'),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000006', 'spouse'),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000007', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 3: The Williams
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', 'Robert', 'Williams', 'robert.williams@email.com', '555-0301', '789 Pine Road', 'Springfield', 'IL', '62703', '1972-12-03', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000009', '00000000-0000-0000-0000-000000000001', 'Linda', 'Williams', 'linda.williams@email.com', '555-0302', '789 Pine Road', 'Springfield', 'IL', '62703', '1975-06-28', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', 'Daniel', 'Williams', 'daniel.williams@email.com', '555-0303', '789 Pine Road', 'Springfield', 'IL', '62703', '2006-04-12', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', 'Grace', 'Williams', NULL, NULL, '789 Pine Road', 'Springfield', 'IL', '62703', '2010-09-30', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', 'Noah', 'Williams', NULL, NULL, '789 Pine Road', 'Springfield', 'IL', '62703', '2015-01-15', 'male', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Williams Family', '20000000-0000-0000-0000-000000000008', '789 Pine Road', 'Springfield', 'IL', '62703')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000008', 'head'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000009', 'spouse'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000010', 'child'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000011', 'child'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000012', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 4: The Browns
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000013', '00000000-0000-0000-0000-000000000001', 'James', 'Brown', 'james.brown@email.com', '555-0401', '321 Elm Street', 'Springfield', 'IL', '62704', '1990-05-25', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000014', '00000000-0000-0000-0000-000000000001', 'Emma', 'Brown', 'emma.brown@email.com', '555-0402', '321 Elm Street', 'Springfield', 'IL', '62704', '1992-11-08', 'female', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Brown Family', '20000000-0000-0000-0000-000000000013', '321 Elm Street', 'Springfield', 'IL', '62704')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000013', 'head'),
    ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000014', 'spouse')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 5: The Davis
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000015', '00000000-0000-0000-0000-000000000001', 'Christopher', 'Davis', 'chris.davis@email.com', '555-0501', '654 Birch Lane', 'Springfield', 'IL', '62705', '1988-07-14', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000016', '00000000-0000-0000-0000-000000000001', 'Amanda', 'Davis', 'amanda.davis@email.com', '555-0502', '654 Birch Lane', 'Springfield', 'IL', '62705', '1989-03-22', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000001', 'Sophia', 'Davis', NULL, NULL, '654 Birch Lane', 'Springfield', 'IL', '62705', '2013-06-10', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000018', '00000000-0000-0000-0000-000000000001', 'Liam', 'Davis', NULL, NULL, '654 Birch Lane', 'Springfield', 'IL', '62705', '2016-02-28', 'male', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'Davis Family', '20000000-0000-0000-0000-000000000015', '654 Birch Lane', 'Springfield', 'IL', '62705')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000015', 'head'),
    ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000016', 'spouse'),
    ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000017', 'child'),
    ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000018', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 6: The Millers
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000019', '00000000-0000-0000-0000-000000000001', 'Matthew', 'Miller', 'matthew.miller@email.com', '555-0601', '987 Cedar Court', 'Springfield', 'IL', '62706', '1983-10-05', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000001', 'Jennifer', 'Miller', 'jennifer.miller@email.com', '555-0602', '987 Cedar Court', 'Springfield', 'IL', '62706', '1984-08-19', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000001', 'Ethan', 'Miller', 'ethan.miller@email.com', '555-0603', '987 Cedar Court', 'Springfield', 'IL', '62706', '2007-12-24', 'male', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', 'Miller Family', '20000000-0000-0000-0000-000000000019', '987 Cedar Court', 'Springfield', 'IL', '62706')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000019', 'head'),
    ('30000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000020', 'spouse'),
    ('30000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000021', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 7: The Thompsons
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000022', '00000000-0000-0000-0000-000000000001', 'Daniel', 'Thompson', 'daniel.thompson@email.com', '555-0701', '147 Willow Way', 'Springfield', 'IL', '62707', '1976-01-30', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000001', 'Rebecca', 'Thompson', 'rebecca.thompson@email.com', '555-0702', '147 Willow Way', 'Springfield', 'IL', '62707', '1977-09-12', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000024', '00000000-0000-0000-0000-000000000001', 'Hannah', 'Thompson', 'hannah.thompson@email.com', '555-0703', '147 Willow Way', 'Springfield', 'IL', '62707', '2005-03-18', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000025', '00000000-0000-0000-0000-000000000001', 'Isaac', 'Thompson', NULL, NULL, '147 Willow Way', 'Springfield', 'IL', '62707', '2009-11-22', 'male', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', 'Thompson Family', '20000000-0000-0000-0000-000000000022', '147 Willow Way', 'Springfield', 'IL', '62707')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000022', 'head'),
    ('30000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000023', 'spouse'),
    ('30000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000024', 'child'),
    ('30000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000025', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Household 8: The Andersons
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000026', '00000000-0000-0000-0000-000000000001', 'Andrew', 'Anderson', 'andrew.anderson@email.com', '555-0801', '258 Spruce Drive', 'Springfield', 'IL', '62708', '1981-04-07', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000027', '00000000-0000-0000-0000-000000000001', 'Michelle', 'Anderson', 'michelle.anderson@email.com', '555-0802', '258 Spruce Drive', 'Springfield', 'IL', '62708', '1982-12-16', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000028', '00000000-0000-0000-0000-000000000001', 'Ava', 'Anderson', NULL, NULL, '258 Spruce Drive', 'Springfield', 'IL', '62708', '2011-07-09', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000029', '00000000-0000-0000-0000-000000000001', 'Mason', 'Anderson', NULL, NULL, '258 Spruce Drive', 'Springfield', 'IL', '62708', '2014-10-03', 'male', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO households (id, tenant_id, name, primary_contact_id, address_line1, city, state, zip)
VALUES ('30000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', 'Anderson Family', '20000000-0000-0000-0000-000000000026', '258 Spruce Drive', 'Springfield', 'IL', '62708')
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (household_id, person_id, role)
VALUES
    ('30000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000026', 'head'),
    ('30000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000027', 'spouse'),
    ('30000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000028', 'child'),
    ('30000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000029', 'child')
ON CONFLICT (household_id, person_id) DO NOTHING;

-- Single members and visitors
INSERT INTO people (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip, birthdate, gender, membership_status)
VALUES
    ('20000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000001', 'Rachel', 'Green', 'rachel.green@email.com', '555-0901', '369 Oak Avenue', 'Springfield', 'IL', '62709', '1995-06-15', 'female', 'active'),
    ('20000000-0000-0000-0000-000000000031', '00000000-0000-0000-0000-000000000001', 'Tyler', 'Martinez', 'tyler.martinez@email.com', '555-1001', '741 Pine Plaza', 'Springfield', 'IL', '62710', '1998-02-28', 'male', 'active'),
    ('20000000-0000-0000-0000-000000000032', '00000000-0000-0000-0000-000000000001', 'Samantha', 'Lee', 'samantha.lee@email.com', '555-1101', '852 Maple Court', 'Springfield', 'IL', '62711', '1993-09-05', 'female', 'inactive'),
    ('20000000-0000-0000-0000-000000000033', '00000000-0000-0000-0000-000000000001', 'Jordan', 'Taylor', 'jordan.taylor@email.com', '555-1201', '963 Elm Boulevard', 'Springfield', 'IL', '62712', '2000-11-18', 'male', 'visitor')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- PERSON TAGS
-- ============================================================================

-- Michael Johnson: Small Group Leader, Volunteer
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000004')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- Sarah Johnson: Worship Team, Volunteer
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- Emily Johnson, Daniel Williams, Hannah Thompson, Ethan Miller: Youth
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002'),
    ('20000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000002'),
    ('20000000-0000-0000-0000-000000000024', '10000000-0000-0000-0000-000000000002'),
    ('20000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000002')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- David Smith: Worship Team
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000003')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- Jessica Smith: Small Group Leader
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000004')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- James Brown, Emma Brown: New Member
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000005'),
    ('20000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000005')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- Rachel Green, Tyler Martinez: Volunteer
INSERT INTO person_tags (person_id, tag_id)
VALUES
    ('20000000-0000-0000-0000-000000000030', '10000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000031', '10000000-0000-0000-0000-000000000001')
ON CONFLICT (person_id, tag_id) DO NOTHING;

-- ============================================================================
-- GROUPS
-- ============================================================================

INSERT INTO groups (id, tenant_id, name, description, group_type, meeting_day, meeting_time, meeting_location, is_public, is_active)
VALUES
    ('40000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Monday Night Small Group', 'A weekly gathering for fellowship, prayer, and Bible study focusing on practical Christian living.', 'small_group', 'monday', '7:00 PM', 'Johnson Home - 123 Oak Street', true, true),
    ('40000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Wednesday Bible Study', 'In-depth study of Scripture with practical application and discussion.', 'bible_study', 'wednesday', '6:30 PM', 'Church Fellowship Hall', true, true),
    ('40000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Youth Group (Grades 6-12)', 'Weekly youth gathering with games, worship, and relevant teaching for teens.', 'small_group', 'friday', '7:00 PM', 'Youth Center', true, true),
    ('40000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Worship Team', 'Musicians and vocalists who lead Sunday morning worship.', 'ministry_team', 'sunday', '8:30 AM', 'Church Sanctuary', false, true),
    ('40000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'Prayer Team', 'Dedicated intercessors meeting weekly to pray for church needs and community.', 'ministry_team', 'tuesday', '6:00 AM', 'Prayer Room', true, true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- GROUP MEMBERS
-- ============================================================================

-- Monday Night Small Group (led by Michael Johnson)
INSERT INTO group_members (group_id, person_id, role)
VALUES
    ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'leader'),
    ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002', 'co_leader'),
    ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', 'member'),
    ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000006', 'member'),
    ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000013', 'member'),
    ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000014', 'member')
ON CONFLICT (group_id, person_id) DO NOTHING;

-- Wednesday Bible Study (led by Jessica Smith)
INSERT INTO group_members (group_id, person_id, role)
VALUES
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000006', 'leader'),
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000008', 'member'),
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000009', 'member'),
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000019', 'member'),
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000020', 'member'),
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000022', 'member'),
    ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000023', 'member')
ON CONFLICT (group_id, person_id) DO NOTHING;

-- Youth Group
INSERT INTO group_members (group_id, person_id, role)
VALUES
    ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000015', 'leader'),
    ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000016', 'co_leader'),
    ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 'member'),
    ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000010', 'member'),
    ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000021', 'member'),
    ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000024', 'member')
ON CONFLICT (group_id, person_id) DO NOTHING;

-- Worship Team
INSERT INTO group_members (group_id, person_id, role)
VALUES
    ('40000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000002', 'leader'),
    ('40000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000005', 'member'),
    ('40000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000026', 'member'),
    ('40000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000030', 'member')
ON CONFLICT (group_id, person_id) DO NOTHING;

-- Prayer Team
INSERT INTO group_members (group_id, person_id, role)
VALUES
    ('40000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000009', 'leader'),
    ('40000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000023', 'member'),
    ('40000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000027', 'member'),
    ('40000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000031', 'member')
ON CONFLICT (group_id, person_id) DO NOTHING;

-- ============================================================================
-- SONGS
-- ============================================================================

INSERT INTO songs (id, tenant_id, title, artist, default_key, tempo, ccli_number, tags, times_used)
VALUES
    ('50000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Way Maker', 'Sinach', 'C', 136, '7115744', 'worship,slow,powerful', 15),
    ('50000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Goodness of God', 'Bethel Music', 'G', 119, '7117726', 'worship,medium,testimony', 12),
    ('50000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Reckless Love', 'Cory Asbury', 'Bb', 126, '7089641', 'worship,passion,love', 10),
    ('50000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'What a Beautiful Name', 'Hillsong Worship', 'D', 68, '7068424', 'worship,slow,jesus', 18),
    ('50000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'Build My Life', 'Housefires', 'G', 70, '7070345', 'worship,devotion,surrender', 14),
    ('50000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', 'O Come to the Altar', 'Elevation Worship', 'C', 128, '7051511', 'worship,invitation,altar', 8),
    ('50000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', 'Graves Into Gardens', 'Elevation Worship', 'G', 71, '7138219', 'worship,resurrection,hope', 9),
    ('50000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', 'Raise a Hallelujah', 'Bethel Music', 'C', 138, '7119315', 'worship,warfare,victory', 11),
    ('50000000-0000-0000-0000-000000000009', '00000000-0000-0000-0000-000000000001', 'The Blessing', 'Kari Jobe', 'Bb', 132, '7147007', 'worship,blessing,peace', 7),
    ('50000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', 'King of Kings', 'Hillsong Worship', 'A', 125, '7127647', 'worship,jesus,majesty', 13),
    ('50000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', 'Great Are You Lord', 'All Sons & Daughters', 'G', 68, '6460220', 'worship,slow,adoration', 16),
    ('50000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', 'How Great Is Our God', 'Chris Tomlin', 'C', 76, '4348399', 'worship,classic,powerful', 20),
    ('50000000-0000-0000-0000-000000000013', '00000000-0000-0000-0000-000000000001', 'Amazing Grace (My Chains Are Gone)', 'Chris Tomlin', 'G', 80, '4768151', 'hymn,classic,grace', 17),
    ('50000000-0000-0000-0000-000000000014', '00000000-0000-0000-0000-000000000001', '10,000 Reasons (Bless the Lord)', 'Matt Redman', 'C', 73, '6016351', 'worship,gratitude,blessing', 19),
    ('50000000-0000-0000-0000-000000000015', '00000000-0000-0000-0000-000000000001', 'Holy Spirit', 'Francesca Battistelli', 'D', 72, '6087919', 'worship,spirit,invitation', 6),
    ('50000000-0000-0000-0000-000000000016', '00000000-0000-0000-0000-000000000001', 'In Christ Alone', 'Keith Getty', 'D', 80, '3350395', 'hymn,gospel,foundation', 14)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SERVICE TYPES
-- ============================================================================

INSERT INTO service_types (id, tenant_id, name, default_time, default_day, color, is_active)
VALUES
    ('60000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Sunday Morning Service', '10:30 AM', 'sunday', '#4A8B8C', true),
    ('60000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Wednesday Night Service', '7:00 PM', 'wednesday', '#F59E0B', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SERVICES (Past and Upcoming)
-- ============================================================================

-- Past Sunday services (4)
INSERT INTO services (id, tenant_id, service_type_id, service_date, service_time, status)
VALUES
    ('70000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '2025-01-12', '10:30 AM', 'completed'),
    ('70000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '2025-01-19', '10:30 AM', 'completed'),
    ('70000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '2025-01-26', '10:30 AM', 'completed'),
    ('70000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '2025-02-02', '10:30 AM', 'completed')
ON CONFLICT (id) DO NOTHING;

-- Past Wednesday services (2)
INSERT INTO services (id, tenant_id, service_type_id, service_date, service_time, status)
VALUES
    ('70000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000002', '2025-01-22', '7:00 PM', 'completed'),
    ('70000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000002', '2025-01-29', '7:00 PM', 'completed')
ON CONFLICT (id) DO NOTHING;

-- Upcoming Sunday services (2)
INSERT INTO services (id, tenant_id, service_type_id, service_date, service_time, status)
VALUES
    ('70000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '2025-02-09', '10:30 AM', 'confirmed'),
    ('70000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '2025-02-16', '10:30 AM', 'planning')
ON CONFLICT (id) DO NOTHING;

-- Upcoming Wednesday (2)
INSERT INTO services (id, tenant_id, service_type_id, service_date, service_time, status)
VALUES
    ('70000000-0000-0000-0000-000000000009', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000002', '2025-02-12', '7:00 PM', 'confirmed'),
    ('70000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000002', '2025-02-19', '7:00 PM', 'planning')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SERVICE ITEMS (Song setlists for completed services)
-- ============================================================================

-- Service 1 (Jan 12) setlist
INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes, assigned_to)
VALUES
    ('70000000-0000-0000-0000-000000000001', 'song', 'How Great Is Our God', '50000000-0000-0000-0000-000000000012', 'C', 1, 5, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000001', 'song', '10,000 Reasons', '50000000-0000-0000-0000-000000000014', 'C', 2, 6, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000001', 'song', 'Goodness of God', '50000000-0000-0000-0000-000000000002', 'G', 3, 7, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000001', 'prayer', 'Opening Prayer', NULL, NULL, 4, 3, 'Michael Johnson'),
    ('70000000-0000-0000-0000-000000000001', 'sermon', 'Walking in Faith', NULL, NULL, 5, 35, 'Pastor Roberts'),
    ('70000000-0000-0000-0000-000000000001', 'song', 'Build My Life', '50000000-0000-0000-0000-000000000005', 'G', 6, 6, 'Sarah Johnson')
ON CONFLICT DO NOTHING;

-- Service 2 (Jan 19) setlist
INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes, assigned_to)
VALUES
    ('70000000-0000-0000-0000-000000000002', 'song', 'Great Are You Lord', '50000000-0000-0000-0000-000000000011', 'G', 1, 5, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000002', 'song', 'Way Maker', '50000000-0000-0000-0000-000000000001', 'C', 2, 7, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000002', 'song', 'What a Beautiful Name', '50000000-0000-0000-0000-000000000004', 'D', 3, 6, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000002', 'prayer', 'Opening Prayer', NULL, NULL, 4, 3, 'Robert Williams'),
    ('70000000-0000-0000-0000-000000000002', 'sermon', 'The Power of Prayer', NULL, NULL, 5, 38, 'Pastor Roberts'),
    ('70000000-0000-0000-0000-000000000002', 'song', 'King of Kings', '50000000-0000-0000-0000-000000000010', 'A', 6, 5, 'Sarah Johnson')
ON CONFLICT DO NOTHING;

-- Service 3 (Jan 26) setlist
INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes, assigned_to)
VALUES
    ('70000000-0000-0000-0000-000000000003', 'song', 'Raise a Hallelujah', '50000000-0000-0000-0000-000000000008', 'C', 1, 6, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000003', 'song', 'Reckless Love', '50000000-0000-0000-0000-000000000003', 'Bb', 2, 7, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000003', 'song', 'Amazing Grace', '50000000-0000-0000-0000-000000000013', 'G', 3, 6, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000003', 'prayer', 'Opening Prayer', NULL, NULL, 4, 3, 'Daniel Thompson'),
    ('70000000-0000-0000-0000-000000000003', 'sermon', 'Gods Unfailing Love', NULL, NULL, 5, 40, 'Pastor Roberts'),
    ('70000000-0000-0000-0000-000000000003', 'song', 'In Christ Alone', '50000000-0000-0000-0000-000000000016', 'D', 6, 5, 'Sarah Johnson')
ON CONFLICT DO NOTHING;

-- Service 4 (Feb 2) setlist
INSERT INTO service_items (service_id, item_type, title, song_id, song_key, position, duration_minutes, assigned_to)
VALUES
    ('70000000-0000-0000-0000-000000000004', 'song', 'O Come to the Altar', '50000000-0000-0000-0000-000000000006', 'C', 1, 6, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000004', 'song', 'Graves Into Gardens', '50000000-0000-0000-0000-000000000007', 'G', 2, 7, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000004', 'song', 'The Blessing', '50000000-0000-0000-0000-000000000009', 'Bb', 3, 8, 'Sarah Johnson'),
    ('70000000-0000-0000-0000-000000000004', 'prayer', 'Opening Prayer', NULL, NULL, 4, 3, 'Andrew Anderson'),
    ('70000000-0000-0000-0000-000000000004', 'sermon', 'New Life in Christ', NULL, NULL, 5, 37, 'Pastor Roberts'),
    ('70000000-0000-0000-0000-000000000004', 'song', 'Holy Spirit', '50000000-0000-0000-0000-000000000015', 'D', 6, 6, 'Sarah Johnson')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SERVICE TEAMS (for upcoming services)
-- ============================================================================

-- Service 7 (Feb 9 - upcoming) team
INSERT INTO service_teams (service_id, person_id, role, status)
VALUES
    ('70000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000002', 'Worship Leader', 'accepted'),
    ('70000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000005', 'Keys', 'accepted'),
    ('70000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000026', 'Guitar', 'accepted'),
    ('70000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000030', 'Vocals', 'pending'),
    ('70000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000031', 'Sound Tech', 'accepted')
ON CONFLICT (service_id, person_id, role) DO NOTHING;

-- ============================================================================
-- GIVING - FUNDS
-- ============================================================================

INSERT INTO funds (id, tenant_id, name, description, is_default, is_active)
VALUES
    ('80000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'General Fund', 'General church operations and ministry', true, true),
    ('80000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Missions Fund', 'Supporting local and global mission work', false, true),
    ('80000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Building Fund', 'Facility improvements and expansion', false, true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- GIVING - DONATIONS (50+ spread across 3 months)
-- ============================================================================

-- Week of Dec 1-7, 2024
INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2024-12-01 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2024-12-01 11:15:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008', '80000000-0000-0000-0000-000000000001', 20000, 'check', 'completed', '2024-12-01 09:45:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000013', '80000000-0000-0000-0000-000000000001', 5000, 'card', 'completed', '2024-12-03 14:20:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000019', '80000000-0000-0000-0000-000000000002', 7500, 'card', 'completed', '2024-12-05 16:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000022', '80000000-0000-0000-0000-000000000001', 12500, 'check', 'completed', '2024-12-08 10:30:00')
ON CONFLICT DO NOTHING;

-- Week of Dec 8-14, 2024
INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2024-12-08 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2024-12-08 11:20:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008', '80000000-0000-0000-0000-000000000003', 50000, 'check', 'completed', '2024-12-10 09:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000026', '80000000-0000-0000-0000-000000000001', 8000, 'card', 'completed', '2024-12-12 15:45:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000030', '80000000-0000-0000-0000-000000000001', 3000, 'card', 'completed', '2024-12-14 12:00:00')
ON CONFLICT DO NOTHING;

-- Week of Dec 15-21, 2024
INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2024-12-15 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2024-12-15 11:10:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000013', '80000000-0000-0000-0000-000000000001', 5000, 'card', 'completed', '2024-12-17 13:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000019', '80000000-0000-0000-0000-000000000002', 7500, 'card', 'completed', '2024-12-19 10:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000022', '80000000-0000-0000-0000-000000000001', 12500, 'check', 'completed', '2024-12-22 10:30:00')
ON CONFLICT DO NOTHING;

-- Week of Dec 22-28, 2024 (Christmas week - more giving)
INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 25000, 'card', 'completed', '2024-12-22 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2024-12-22 11:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008', '80000000-0000-0000-0000-000000000002', 20000, 'check', 'completed', '2024-12-22 09:45:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000015', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2024-12-24 08:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000026', '80000000-0000-0000-0000-000000000001', 12000, 'card', 'completed', '2024-12-25 20:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000030', '80000000-0000-0000-0000-000000000002', 5000, 'card', 'completed', '2024-12-27 14:30:00')
ON CONFLICT DO NOTHING;

-- January 2025 - Weeks 1-4
INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2025-01-05 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2025-01-05 11:15:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008', '80000000-0000-0000-0000-000000000001', 20000, 'check', 'completed', '2025-01-05 09:45:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000013', '80000000-0000-0000-0000-000000000001', 5000, 'card', 'completed', '2025-01-07 14:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000019', '80000000-0000-0000-0000-000000000002', 7500, 'card', 'completed', '2025-01-09 16:20:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000022', '80000000-0000-0000-0000-000000000001', 12500, 'check', 'completed', '2025-01-12 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2025-01-12 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2025-01-12 11:20:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000026', '80000000-0000-0000-0000-000000000001', 8000, 'card', 'completed', '2025-01-14 15:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000030', '80000000-0000-0000-0000-000000000001', 3000, 'card', 'completed', '2025-01-16 12:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2025-01-19 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2025-01-19 11:05:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000013', '80000000-0000-0000-0000-000000000001', 5000, 'card', 'completed', '2025-01-21 13:45:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000019', '80000000-0000-0000-0000-000000000002', 7500, 'card', 'completed', '2025-01-23 09:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000022', '80000000-0000-0000-0000-000000000001', 12500, 'check', 'completed', '2025-01-26 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2025-01-26 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2025-01-26 11:25:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000026', '80000000-0000-0000-0000-000000000001', 8000, 'card', 'completed', '2025-01-28 14:15:00')
ON CONFLICT DO NOTHING;

-- February 2025 - Week 1-2
INSERT INTO donations (tenant_id, person_id, fund_id, amount_cents, payment_method, status, donated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 15000, 'card', 'completed', '2025-02-02 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '80000000-0000-0000-0000-000000000001', 10000, 'card', 'completed', '2025-02-02 11:15:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008', '80000000-0000-0000-0000-000000000001', 20000, 'check', 'completed', '2025-02-02 09:45:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000013', '80000000-0000-0000-0000-000000000001', 5000, 'card', 'completed', '2025-02-04 13:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000019', '80000000-0000-0000-0000-000000000002', 7500, 'card', 'completed', '2025-02-06 10:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000022', '80000000-0000-0000-0000-000000000001', 12500, 'check', 'completed', '2025-02-09 10:30:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000026', '80000000-0000-0000-0000-000000000001', 8000, 'card', 'completed', '2025-02-09 15:00:00'),
    ('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000030', '80000000-0000-0000-0000-000000000001', 3000, 'card', 'completed', '2025-02-10 12:00:00')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- CHECK-IN STATIONS
-- ============================================================================

INSERT INTO checkin_stations (id, tenant_id, name, location, is_active)
VALUES
    ('90000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Main Entrance', 'Church Lobby', true),
    ('90000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Kids Check-In', 'Childrens Wing', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CHECK-IN EVENTS (Last 4 Sundays)
-- ============================================================================

INSERT INTO checkin_events (id, tenant_id, name, event_date, service_id, station_id, is_active)
VALUES
    ('A0000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Sunday Service - Jan 12', '2025-01-12', '70000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', false),
    ('A0000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Kids Check-In - Jan 12', '2025-01-12', '70000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000002', false),
    ('A0000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Sunday Service - Jan 19', '2025-01-19', '70000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', false),
    ('A0000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Kids Check-In - Jan 19', '2025-01-19', '70000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000002', false),
    ('A0000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'Sunday Service - Jan 26', '2025-01-26', '70000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000001', false),
    ('A0000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', 'Kids Check-In - Jan 26', '2025-01-26', '70000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000002', false),
    ('A0000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', 'Sunday Service - Feb 2', '2025-02-02', '70000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000001', false),
    ('A0000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', 'Kids Check-In - Feb 2', '2025-02-02', '70000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000002', false)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CHECK-INS (20+ across 4 Sundays)
-- ============================================================================

-- Jan 12 check-ins
INSERT INTO checkins (tenant_id, event_id, person_id, station_id, checked_in_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', '2025-01-12 10:15:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', '2025-01-12 10:15:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000002', '2025-01-12 10:18:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000002', '2025-01-12 10:18:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005', '90000000-0000-0000-0000-000000000001', '2025-01-12 10:20:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000006', '90000000-0000-0000-0000-000000000001', '2025-01-12 10:20:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008', '90000000-0000-0000-0000-000000000001', '2025-01-12 10:22:00')
ON CONFLICT DO NOTHING;

-- Jan 19 check-ins
INSERT INTO checkins (tenant_id, event_id, person_id, station_id, checked_in_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', '2025-01-19 10:12:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', '2025-01-19 10:12:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000002', '2025-01-19 10:14:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000002', '2025-01-19 10:14:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000013', '90000000-0000-0000-0000-000000000001', '2025-01-19 10:18:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000014', '90000000-0000-0000-0000-000000000001', '2025-01-19 10:18:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000019', '90000000-0000-0000-0000-000000000001', '2025-01-19 10:25:00')
ON CONFLICT DO NOTHING;

-- Jan 26 check-ins
INSERT INTO checkins (tenant_id, event_id, person_id, station_id, checked_in_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000005', '90000000-0000-0000-0000-000000000001', '2025-01-26 10:10:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000006', '90000000-0000-0000-0000-000000000001', '2025-01-26 10:10:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000007', '90000000-0000-0000-0000-000000000002', '2025-01-26 10:12:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000022', '90000000-0000-0000-0000-000000000001', '2025-01-26 10:15:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000023', '90000000-0000-0000-0000-000000000001', '2025-01-26 10:15:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000026', '90000000-0000-0000-0000-000000000001', '2025-01-26 10:17:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000030', '90000000-0000-0000-0000-000000000001', '2025-01-26 10:20:00')
ON CONFLICT DO NOTHING;

-- Feb 2 check-ins
INSERT INTO checkins (tenant_id, event_id, person_id, station_id, checked_in_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', '2025-02-02 10:16:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', '2025-02-02 10:16:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000002', '2025-02-02 10:18:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000002', '2025-02-02 10:18:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000008', '90000000-0000-0000-0000-000000000001', '2025-02-02 10:19:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000015', '90000000-0000-0000-0000-000000000001', '2025-02-02 10:21:00'),
    ('00000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000031', '90000000-0000-0000-0000-000000000001', '2025-02-02 10:24:00')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- COMMUNICATION - MESSAGE TEMPLATES
-- ============================================================================

INSERT INTO message_templates (id, tenant_id, name, subject, body, channel, category)
VALUES
    ('B0000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'New Visitor Welcome', 'Welcome to Grace Community Church!', 
     'Hi {first_name},\n\nWe are so glad you visited us! We would love to connect with you and help you get involved.\n\nBlessings,\nGrace Community Church',
     'email', 'welcome'),
    ('B0000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Small Group Invitation', 'Join a Small Group!',
     'Hi {first_name},\n\nWe noticed you might be interested in joining a small group. We have several groups that meet throughout the week.\n\nCheck them out at our website!\n\nGrace Community Church',
     'email', 'follow_up')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- COMMUNICATION - CAMPAIGNS
-- ============================================================================

INSERT INTO campaigns (id, tenant_id, name, channel, subject, body, status, sent_at, recipient_count, target_type)
VALUES
    ('C0000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Christmas Eve Service Announcement', 'email', 
     'Join us for Christmas Eve!', 
     'We would love to see you at our special Christmas Eve service on December 24th at 7:00 PM. Bring your family and friends!',
     'sent', '2024-12-15 09:00:00', 25, 'all'),
    ('C0000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'New Year Small Group Kickoff', 'email',
     'Start the New Year in Community',
     'Join us for our small group kickoff in January! Groups are forming now. Sign up today!',
     'sent', '2025-01-05 10:00:00', 30, 'all'),
    ('C0000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Volunteer Appreciation', 'email',
     'Thank You for Serving!',
     'We want to say a huge THANK YOU to all our volunteers who make our church family thrive. You are appreciated!',
     'sent', '2025-01-20 14:00:00', 12, 'tag')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- COMMUNICATION - JOURNEYS
-- ============================================================================

INSERT INTO journeys (id, tenant_id, name, description, trigger_type, is_active)
VALUES
    ('D0000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'First-Time Visitor Follow-Up', 
     'Automated 3-email sequence for first-time visitors',
     'checkin_first_time', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO journey_steps (journey_id, position, step_type, delay_days, template_id)
VALUES
    ('D0000000-0000-0000-0000-000000000001', 1, 'send_email', 1, 'B0000000-0000-0000-0000-000000000001'),
    ('D0000000-0000-0000-0000-000000000001', 2, 'send_email', 7, 'B0000000-0000-0000-0000-000000000002')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- STREAMS
-- ============================================================================

INSERT INTO streams (id, tenant_id, title, description, service_id, status, scheduled_start, actual_start, actual_end, stream_type, embed_url, chat_enabled, giving_enabled)
VALUES
    ('E0000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 
     'Sunday Service - January 12', 
     'Join us live for Sunday morning worship and teaching',
     '70000000-0000-0000-0000-000000000001', 
     'ended', 
     '2025-01-12 10:30:00', 
     '2025-01-12 10:32:00', 
     '2025-01-12 11:45:00',
     'youtube',
     'https://www.youtube.com/embed/dQw4w9WgXcQ',
     true, true),
    ('E0000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
     'Sunday Service - January 26',
     'Join us live for Sunday morning worship and teaching',
     '70000000-0000-0000-0000-000000000003',
     'ended',
     '2025-01-26 10:30:00',
     '2025-01-26 10:31:00',
     '2025-01-26 11:50:00',
     'youtube',
     'https://www.youtube.com/embed/dQw4w9WgXcQ',
     true, true),
    ('E0000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001',
     'Sunday Service - February 9',
     'Join us live for Sunday morning worship and teaching',
     '70000000-0000-0000-0000-000000000007',
     'scheduled',
     '2025-02-09 10:30:00',
     NULL,
     NULL,
     'youtube',
     'https://www.youtube.com/embed/dQw4w9WgXcQ',
     true, true)
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Summary
DO $$
BEGIN
    RAISE NOTICE 'Demo seed data successfully created for Grace Community Church!';
    RAISE NOTICE 'Tenant: demo-church';
    RAISE NOTICE '- 33 people across 8 households';
    RAISE NOTICE '- 5 groups with members';
    RAISE NOTICE '- 16 worship songs';
    RAISE NOTICE '- 10 services (6 past, 4 upcoming)';
    RAISE NOTICE '- 50+ donations across 3 funds';
    RAISE NOTICE '- 27+ check-ins across 4 Sundays';
    RAISE NOTICE '- 3 communication campaigns';
    RAISE NOTICE '- 1 automated journey';
    RAISE NOTICE '- 3 streams (2 past, 1 upcoming)';
END $$;
