-- Service Plans (worship planning / service builder)
CREATE TABLE service_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    notes TEXT,
    status VARCHAR(50) DEFAULT 'draft', -- draft, published
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE service_plans ENABLE ROW LEVEL SECURITY;

-- Service plans can only see their own tenant's data
CREATE POLICY service_plans_isolation_policy ON service_plans
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Indexes
CREATE INDEX idx_service_plans_tenant_id ON service_plans(tenant_id);
CREATE INDEX idx_service_plans_service_id ON service_plans(service_id);
CREATE INDEX idx_service_plans_status ON service_plans(status);

-- Updated_at trigger
CREATE TRIGGER update_service_plans_updated_at
    BEFORE UPDATE ON service_plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Service Plan Items (individual items in the plan)
CREATE TABLE service_plan_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
    item_order INTEGER NOT NULL,
    item_type VARCHAR(50) NOT NULL, -- song, scripture, prayer, announcement, video, other
    title VARCHAR(255) NOT NULL,
    duration_minutes INTEGER,
    notes TEXT,
    song_id UUID REFERENCES songs(id),
    assigned_to UUID REFERENCES users(id), -- user_id for band member
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE service_plan_items ENABLE ROW LEVEL SECURITY;

-- Service plan items inherit tenant access from service_plans
CREATE POLICY service_plan_items_isolation_policy ON service_plan_items
    USING (EXISTS (
        SELECT 1 FROM service_plans 
        WHERE service_plans.id = service_plan_items.plan_id 
        AND service_plans.tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID
    ));

-- Indexes
CREATE INDEX idx_service_plan_items_plan_id ON service_plan_items(plan_id);
CREATE INDEX idx_service_plan_items_order ON service_plan_items(plan_id, item_order);
CREATE INDEX idx_service_plan_items_song_id ON service_plan_items(song_id);
CREATE INDEX idx_service_plan_items_assigned_to ON service_plan_items(assigned_to);

-- Updated_at trigger
CREATE TRIGGER update_service_plan_items_updated_at
    BEFORE UPDATE ON service_plan_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
