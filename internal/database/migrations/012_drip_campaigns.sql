-- Drip campaigns table
CREATE TABLE drip_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    trigger_event VARCHAR(50) NOT NULL, -- new_member, connection_card, first_visit
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE drip_campaigns ENABLE ROW LEVEL SECURITY;

CREATE POLICY drip_campaigns_isolation_policy ON drip_campaigns
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_drip_campaigns_tenant_id ON drip_campaigns(tenant_id);
CREATE INDEX idx_drip_campaigns_trigger_event ON drip_campaigns(trigger_event);
CREATE INDEX idx_drip_campaigns_is_active ON drip_campaigns(is_active);

-- Updated_at trigger
CREATE TRIGGER update_drip_campaigns_updated_at
    BEFORE UPDATE ON drip_campaigns
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Drip campaign steps
CREATE TABLE drip_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES drip_campaigns(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    delay_days INTEGER NOT NULL DEFAULT 0,
    action_type VARCHAR(50) NOT NULL, -- email, sms, follow_up
    subject VARCHAR(500), -- for email
    body TEXT NOT NULL,
    template_id UUID REFERENCES message_templates(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE drip_steps ENABLE ROW LEVEL SECURITY;

CREATE POLICY drip_steps_isolation_policy ON drip_steps
    USING (EXISTS (
        SELECT 1 FROM drip_campaigns 
        WHERE drip_campaigns.id = drip_steps.campaign_id 
        AND drip_campaigns.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_drip_steps_campaign_id ON drip_steps(campaign_id);
CREATE INDEX idx_drip_steps_step_order ON drip_steps(step_order);

-- Updated_at trigger
CREATE TRIGGER update_drip_steps_updated_at
    BEFORE UPDATE ON drip_steps
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Drip campaign enrollments
CREATE TABLE drip_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES drip_campaigns(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id),
    status VARCHAR(50) DEFAULT 'active', -- active, completed, paused, cancelled
    enrolled_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    UNIQUE(campaign_id, person_id)
);

-- Enable RLS
ALTER TABLE drip_enrollments ENABLE ROW LEVEL SECURITY;

CREATE POLICY drip_enrollments_isolation_policy ON drip_enrollments
    USING (EXISTS (
        SELECT 1 FROM drip_campaigns 
        WHERE drip_campaigns.id = drip_enrollments.campaign_id 
        AND drip_campaigns.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_drip_enrollments_campaign_id ON drip_enrollments(campaign_id);
CREATE INDEX idx_drip_enrollments_person_id ON drip_enrollments(person_id);
CREATE INDEX idx_drip_enrollments_status ON drip_enrollments(status);

-- Drip step executions (track when each step was executed)
CREATE TABLE drip_step_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES drip_enrollments(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES drip_steps(id),
    status VARCHAR(50) DEFAULT 'pending', -- pending, sent, failed
    scheduled_at TIMESTAMP NOT NULL,
    executed_at TIMESTAMP,
    error_message TEXT,
    UNIQUE(enrollment_id, step_id)
);

-- Enable RLS
ALTER TABLE drip_step_executions ENABLE ROW LEVEL SECURITY;

CREATE POLICY drip_step_executions_isolation_policy ON drip_step_executions
    USING (EXISTS (
        SELECT 1 FROM drip_enrollments
        JOIN drip_campaigns ON drip_campaigns.id = drip_enrollments.campaign_id
        WHERE drip_enrollments.id = drip_step_executions.enrollment_id 
        AND drip_campaigns.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_drip_step_executions_enrollment_id ON drip_step_executions(enrollment_id);
CREATE INDEX idx_drip_step_executions_step_id ON drip_step_executions(step_id);
CREATE INDEX idx_drip_step_executions_status ON drip_step_executions(status);
CREATE INDEX idx_drip_step_executions_scheduled_at ON drip_step_executions(scheduled_at);
