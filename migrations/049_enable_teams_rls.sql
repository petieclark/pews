-- Enable RLS for volunteer team tables (Issue #58)
-- This fixes toast errors on service detail page team panel

ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
CREATE POLICY teams_isolation_policy ON teams
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

ALTER TABLE team_positions ENABLE ROW LEVEL SECURITY;
CREATE POLICY team_positions_isolation_policy ON team_positions
    USING (EXISTS (
        SELECT 1 FROM teams WHERE teams.id = team_positions.team_id 
        AND teams.tenant_id = current_setting('app.current_tenant', true)::uuid
    ));

ALTER TABLE team_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY team_members_isolation_policy ON team_members
    USING (EXISTS (
        SELECT 1 FROM teams WHERE teams.id = team_members.team_id 
        AND teams.tenant_id = current_setting('app.current_tenant', true)::uuid
    ));
