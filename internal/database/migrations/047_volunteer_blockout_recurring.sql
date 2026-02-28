-- Task #62: Enhance volunteer blockout system with recurring support
-- Add is_recurring and day_of_week fields to volunteer_availability

ALTER TABLE volunteer_availability 
ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS day_of_week INTEGER;

-- Add validation constraint: day_of_week required if is_recurring=true
ALTER TABLE volunteer_availability 
ADD CONSTRAINT chk_recurring_day_of_week 
CHECK (NOT is_recurring OR (is_recurring AND day_of_week IS NOT NULL));

-- Add constraints for day_of_week range (0=Sun, 6=Sat)
ALTER TABLE volunteer_availability 
ADD CONSTRAINT chk_day_of_week_range 
CHECK (day_of_week IS NULL OR (day_of_week >= 0 AND day_of_week <= 6));

-- Create indexes for recurring queries
CREATE INDEX IF NOT EXISTS idx_volunteer_availability_recurring 
ON volunteer_availability(person_id, is_recurring) WHERE is_recurring = TRUE;

CREATE INDEX IF NOT EXISTS idx_volunteer_availability_day_of_week 
ON volunteer_availability(day_of_week) WHERE day_of_week IS NOT NULL;

-- Update RLS policy to ensure tenant isolation still works
DROP POLICY IF EXISTS volunteer_availability_isolation_policy ON volunteer_availability;
CREATE POLICY volunteer_availability_isolation_policy ON volunteer_availability
    USING (EXISTS (
        SELECT 1 FROM people 
        WHERE people.id = volunteer_availability.person_id 
        AND people.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Enhanced availability check that handles recurring blockouts
DROP FUNCTION IF EXISTS is_person_available(UUID, DATE);
CREATE OR REPLACE FUNCTION is_person_available(
    p_person_id UUID,
    p_date DATE
) RETURNS BOOLEAN AS $$
DECLARE
    check_day INTEGER;
BEGIN
    -- Check for non-recurring blockouts (date range matches)
    IF EXISTS (
        SELECT 1 FROM volunteer_availability
        WHERE person_id = p_person_id
        AND is_recurring = FALSE
        AND p_date BETWEEN start_date AND end_date
    ) THEN
        RETURN FALSE;
    END IF;

    -- Check for recurring blockouts (day of week matches)
    check_day := EXTRACT(DOW FROM p_date)::INTEGER;  -- DOW: 0=Sun, 1=Mon, ..., 6=Sat
    
    IF EXISTS (
        SELECT 1 FROM volunteer_availability
        WHERE person_id = p_person_id
        AND is_recurring = TRUE
        AND day_of_week = check_day
    ) THEN
        RETURN FALSE;
    END IF;

    -- No conflicts found - person is available
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Add improved conflict detection function that returns matching blockout info
DROP FUNCTION IF EXISTS get_volunteer_conflicts(UUID, DATE);
CREATE OR REPLACE FUNCTION get_volunteer_conflicts(
    p_person_id UUID,
    p_date DATE
) RETURNS TABLE (
    is_blocked BOOLEAN,
    conflict_type VARCHAR,  -- 'date_range' or 'recurring'
    start_date DATE,
    end_date DATE,
    reason TEXT,
    day_of_week INTEGER
) AS $$
DECLARE
    check_day INTEGER;
BEGIN
    check_day := EXTRACT(DOW FROM p_date)::INTEGER;

    -- Check for non-recurring blockouts (date range matches)
    RETURN QUERY
    SELECT 
        TRUE as is_blocked,
        'date_range'::VARCHAR as conflict_type,
        start_date,
        end_date,
        reason,
        NULL as day_of_week
    FROM volunteer_availability
    WHERE person_id = p_person_id
    AND is_recurring = FALSE
    AND p_date BETWEEN start_date AND end_date
    LIMIT 1;

    -- If no date range conflict found, check recurring
    IF NOT FOUND THEN
        RETURN QUERY
        SELECT 
            TRUE as is_blocked,
            'recurring'::VARCHAR as conflict_type,
            NULL as start_date,
            NULL as end_date,
            reason,
            day_of_week
        FROM volunteer_availability
        WHERE person_id = p_person_id
        AND is_recurring = TRUE
        AND day_of_week = check_day
        LIMIT 1;

        -- No conflicts found - return empty row (will be filtered by caller)
        IF NOT FOUND THEN
            RETURN QUERY
            SELECT FALSE, NULL::VARCHAR, NULL, NULL, NULL::TEXT, NULL;
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Helper function to get all blockouts for a person (for API endpoint)
DROP FUNCTION IF EXISTS get_person_blockouts(UUID);
CREATE OR REPLACE FUNCTION get_person_blockouts(p_person_id UUID) 
RETURNS TABLE (
    id UUID,
    person_id UUID,
    team_id UUID,
    start_date DATE,
    end_date DATE,
    reason TEXT,
    is_recurring BOOLEAN,
    day_of_week INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    team_name VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        va.id,
        va.person_id,
        va.team_id,
        va.start_date,
        va.end_date,
        va.reason,
        va.is_recurring,
        va.day_of_week,
        va.created_at,
        va.updated_at,
        vt.name as team_name
    FROM volunteer_availability va
    LEFT JOIN volunteer_teams vt ON vt.id = va.team_id
    WHERE va.person_id = p_person_id
    ORDER BY va.start_date DESC, va.is_recurring DESC;
END;
$$ LANGUAGE plpgsql;

-- Record migration
INSERT INTO schema_migrations (version) VALUES ('047_volunteer_blockout_recurring')
ON CONFLICT (version) DO NOTHING;
