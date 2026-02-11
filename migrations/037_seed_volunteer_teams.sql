-- Seed volunteer teams for demo tenants
-- Only inserts if no teams exist yet
DO $$
DECLARE
    v_tenant_id UUID;
    v_team_id UUID;
BEGIN
    -- Get first tenant
    SELECT id INTO v_tenant_id FROM tenants LIMIT 1;
    IF v_tenant_id IS NULL THEN RETURN; END IF;
    
    -- Skip if teams already exist
    IF EXISTS (SELECT 1 FROM teams WHERE tenant_id = v_tenant_id) THEN RETURN; END IF;

    -- Worship Team
    INSERT INTO teams (tenant_id, name, description, color) VALUES
        (v_tenant_id, 'Worship Team', 'Leading the congregation in musical worship', '#4A8B8C')
        RETURNING id INTO v_team_id;
    INSERT INTO team_positions (team_id, name, sort_order) VALUES
        (v_team_id, 'Worship Leader', 0), (v_team_id, 'Vocalist', 1),
        (v_team_id, 'Guitarist', 2), (v_team_id, 'Bassist', 3),
        (v_team_id, 'Drummer', 4), (v_team_id, 'Keys', 5),
        (v_team_id, 'Audio Tech', 6);

    -- Production Team
    INSERT INTO teams (tenant_id, name, description, color) VALUES
        (v_tenant_id, 'Production Team', 'Technical production for services and events', '#1B3A4B')
        RETURNING id INTO v_team_id;
    INSERT INTO team_positions (team_id, name, sort_order) VALUES
        (v_team_id, 'Slides/ProPresenter', 0), (v_team_id, 'Camera', 1),
        (v_team_id, 'Livestream', 2), (v_team_id, 'Lighting', 3),
        (v_team_id, 'Stage', 4);

    -- Hospitality Team
    INSERT INTO teams (tenant_id, name, description, color) VALUES
        (v_tenant_id, 'Hospitality Team', 'Welcoming guests and creating a warm environment', '#8FBCB0')
        RETURNING id INTO v_team_id;
    INSERT INTO team_positions (team_id, name, sort_order) VALUES
        (v_team_id, 'Greeter', 0), (v_team_id, 'Usher', 1),
        (v_team_id, 'Welcome Desk', 2), (v_team_id, 'Café', 3);

    -- Kids Ministry Team
    INSERT INTO teams (tenant_id, name, description, color) VALUES
        (v_tenant_id, 'Kids Ministry Team', 'Nurturing children in faith through engaging programs', '#F39C12')
        RETURNING id INTO v_team_id;
    INSERT INTO team_positions (team_id, name, sort_order) VALUES
        (v_team_id, 'Lead Teacher', 0), (v_team_id, 'Assistant', 1),
        (v_team_id, 'Check-In', 2), (v_team_id, 'Nursery', 3);
END $$;
