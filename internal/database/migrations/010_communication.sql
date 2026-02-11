-- Message templates
CREATE TABLE message_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(500), -- for email
    body TEXT NOT NULL,
    channel VARCHAR(20) NOT NULL, -- email, sms
    category VARCHAR(50), -- welcome, follow_up, announcement, giving, custom
    variables TEXT, -- comma-separated: "first_name,church_name,group_name"
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE message_templates ENABLE ROW LEVEL SECURITY;

CREATE POLICY message_templates_isolation_policy ON message_templates
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_message_templates_tenant_id ON message_templates(tenant_id);
CREATE INDEX idx_message_templates_channel ON message_templates(channel);
CREATE INDEX idx_message_templates_category ON message_templates(category);

-- Updated_at trigger
CREATE TRIGGER update_message_templates_updated_at
    BEFORE UPDATE ON message_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Campaigns (bulk sends)
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES message_templates(id),
    channel VARCHAR(20) NOT NULL, -- email, sms
    subject VARCHAR(500),
    body TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'draft', -- draft, scheduled, sending, sent, failed
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    recipient_count INTEGER DEFAULT 0,
    opened_count INTEGER DEFAULT 0,
    clicked_count INTEGER DEFAULT 0,
    -- Targeting
    target_type VARCHAR(50) NOT NULL, -- all, tag, group, list, manual
    target_id VARCHAR(255), -- tag name, group ID, etc.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;

CREATE POLICY campaigns_isolation_policy ON campaigns
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_campaigns_tenant_id ON campaigns(tenant_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_scheduled_at ON campaigns(scheduled_at);

-- Updated_at trigger
CREATE TRIGGER update_campaigns_updated_at
    BEFORE UPDATE ON campaigns
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Campaign recipients
CREATE TABLE campaign_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id),
    status VARCHAR(50) DEFAULT 'pending', -- pending, sent, delivered, opened, clicked, bounced, failed
    sent_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP
);

-- Enable RLS
ALTER TABLE campaign_recipients ENABLE ROW LEVEL SECURITY;

CREATE POLICY campaign_recipients_isolation_policy ON campaign_recipients
    USING (EXISTS (
        SELECT 1 FROM campaigns 
        WHERE campaigns.id = campaign_recipients.campaign_id 
        AND campaigns.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_campaign_recipients_campaign_id ON campaign_recipients(campaign_id);
CREATE INDEX idx_campaign_recipients_person_id ON campaign_recipients(person_id);
CREATE INDEX idx_campaign_recipients_status ON campaign_recipients(status);

-- Automated journeys (sequences)
CREATE TABLE journeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(50) NOT NULL, -- first_visit, tag_added, group_joined, manual, checkin_first_time
    trigger_value VARCHAR(255), -- tag name, group ID, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE journeys ENABLE ROW LEVEL SECURITY;

CREATE POLICY journeys_isolation_policy ON journeys
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_journeys_tenant_id ON journeys(tenant_id);
CREATE INDEX idx_journeys_is_active ON journeys(is_active);
CREATE INDEX idx_journeys_trigger_type ON journeys(trigger_type);

-- Updated_at trigger
CREATE TRIGGER update_journeys_updated_at
    BEFORE UPDATE ON journeys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Journey steps
CREATE TABLE journey_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journey_id UUID NOT NULL REFERENCES journeys(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    step_type VARCHAR(50) NOT NULL, -- send_email, send_sms, wait, add_tag, add_to_group
    delay_days INTEGER DEFAULT 0, -- days to wait before executing
    delay_hours INTEGER DEFAULT 0,
    template_id UUID REFERENCES message_templates(id),
    config JSONB DEFAULT '{}', -- additional config (tag name, group ID, custom message)
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE journey_steps ENABLE ROW LEVEL SECURITY;

CREATE POLICY journey_steps_isolation_policy ON journey_steps
    USING (EXISTS (
        SELECT 1 FROM journeys 
        WHERE journeys.id = journey_steps.journey_id 
        AND journeys.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_journey_steps_journey_id ON journey_steps(journey_id);
CREATE INDEX idx_journey_steps_position ON journey_steps(position);

-- Journey enrollments (people in a journey)
CREATE TABLE journey_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journey_id UUID NOT NULL REFERENCES journeys(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id),
    current_step INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active', -- active, completed, paused, exited
    enrolled_at TIMESTAMP NOT NULL DEFAULT NOW(),
    next_step_at TIMESTAMP,
    completed_at TIMESTAMP,
    UNIQUE(journey_id, person_id)
);

-- Enable RLS
ALTER TABLE journey_enrollments ENABLE ROW LEVEL SECURITY;

CREATE POLICY journey_enrollments_isolation_policy ON journey_enrollments
    USING (EXISTS (
        SELECT 1 FROM journeys 
        WHERE journeys.id = journey_enrollments.journey_id 
        AND journeys.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_journey_enrollments_journey_id ON journey_enrollments(journey_id);
CREATE INDEX idx_journey_enrollments_person_id ON journey_enrollments(person_id);
CREATE INDEX idx_journey_enrollments_status ON journey_enrollments(status);
CREATE INDEX idx_journey_enrollments_next_step_at ON journey_enrollments(next_step_at);

-- Digital connection cards
CREATE TABLE connection_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    is_first_visit BOOLEAN DEFAULT TRUE,
    how_heard VARCHAR(255), -- "Friend", "Online", "Drive by"
    prayer_request TEXT,
    interested_in TEXT, -- comma-separated: "small groups, volunteering, membership"
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE,
    person_id UUID REFERENCES people(id) -- linked after processing
);

-- Enable RLS
ALTER TABLE connection_cards ENABLE ROW LEVEL SECURITY;

CREATE POLICY connection_cards_isolation_policy ON connection_cards
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_connection_cards_tenant_id ON connection_cards(tenant_id);
CREATE INDEX idx_connection_cards_processed ON connection_cards(processed);
CREATE INDEX idx_connection_cards_submitted_at ON connection_cards(submitted_at);
