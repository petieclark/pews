-- People Tags system
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#4A8B8C',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);
CREATE INDEX IF NOT EXISTS idx_tags_tenant ON tags(tenant_id);

CREATE TABLE IF NOT EXISTS person_tags (
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (person_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_person_tags_tag ON person_tags(tag_id);

-- Seed default tags for all existing tenants
INSERT INTO tags (tenant_id, name, color)
SELECT t.id, seed.name, seed.color
FROM tenants t
CROSS JOIN (VALUES
    ('First-Time Visitor', '#3b82f6'),
    ('Worship Team', '#8b5cf6'),
    ('Youth', '#f59e0b'),
    ('Volunteer', '#10b981'),
    ('New Member', '#06b6d4'),
    ('Leadership', '#ef4444'),
    ('Small Group Leader', '#ec4899')
) AS seed(name, color)
ON CONFLICT (tenant_id, name) DO NOTHING;
