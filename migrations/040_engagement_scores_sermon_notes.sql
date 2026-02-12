-- Migration 040: Create engagement_scores and sermon_notes tables

CREATE TABLE IF NOT EXISTS engagement_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    score INTEGER NOT NULL DEFAULT 0,
    attendance_score INTEGER NOT NULL DEFAULT 0,
    giving_score INTEGER NOT NULL DEFAULT 0,
    group_score INTEGER NOT NULL DEFAULT 0,
    volunteer_score INTEGER NOT NULL DEFAULT 0,
    connection_score INTEGER NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, person_id)
);

CREATE INDEX IF NOT EXISTS idx_engagement_scores_tenant ON engagement_scores(tenant_id);
CREATE INDEX IF NOT EXISTS idx_engagement_scores_person ON engagement_scores(person_id);
CREATE INDEX IF NOT EXISTS idx_engagement_scores_score ON engagement_scores(tenant_id, score DESC);

CREATE TABLE IF NOT EXISTS engagement_score_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    score INTEGER NOT NULL DEFAULT 0,
    attendance_score INTEGER NOT NULL DEFAULT 0,
    giving_score INTEGER NOT NULL DEFAULT 0,
    group_score INTEGER NOT NULL DEFAULT 0,
    volunteer_score INTEGER NOT NULL DEFAULT 0,
    connection_score INTEGER NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sermon_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service_id UUID REFERENCES services(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    speaker VARCHAR(255) NOT NULL,
    sermon_date DATE NOT NULL,
    scripture_reference TEXT,
    notes_text TEXT,
    audio_url TEXT,
    video_url TEXT,
    series_name VARCHAR(255),
    audio_duration_seconds INTEGER,
    published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sermon_notes_tenant ON sermon_notes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sermon_notes_date ON sermon_notes(tenant_id, sermon_date DESC);
CREATE INDEX IF NOT EXISTS idx_sermon_notes_series ON sermon_notes(tenant_id, series_name);
CREATE INDEX IF NOT EXISTS idx_sermon_notes_published ON sermon_notes(tenant_id, published);
