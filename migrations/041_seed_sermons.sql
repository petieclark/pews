-- Migration 041: Seed sample sermons
-- Uses the first tenant found (demo/dev environment)

INSERT INTO sermon_notes (tenant_id, title, speaker, sermon_date, scripture_reference, notes_text, series_name, published)
SELECT t.id,
       'Walking by Faith',
       'Pastor James Mitchell',
       '2026-02-08',
       '2 Corinthians 5:7',
       'Faith is not the absence of doubt, but the courage to move forward despite it. In this message, we explore what it means to trust God''s plan even when the path ahead is unclear.',
       'Faith Forward',
       true
FROM tenants t LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO sermon_notes (tenant_id, title, speaker, sermon_date, scripture_reference, notes_text, series_name, published)
SELECT t.id,
       'The Power of Community',
       'Pastor James Mitchell',
       '2026-02-01',
       'Acts 2:42-47',
       'The early church thrived because believers devoted themselves to fellowship, breaking of bread, and prayer. We examine how authentic community transforms lives and strengthens our faith journey.',
       'Better Together',
       true
FROM tenants t LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO sermon_notes (tenant_id, title, speaker, sermon_date, scripture_reference, notes_text, series_name, audio_url, published)
SELECT t.id,
       'Grace That Transforms',
       'Rev. Sarah Coleman',
       '2026-01-25',
       'Ephesians 2:8-10',
       'Grace isn''t just a theological concept — it''s the transforming power of God at work in our everyday lives. This sermon explores how receiving grace changes how we see ourselves and others.',
       'Faith Forward',
       'https://example.com/sermons/grace-transforms.mp3',
       true
FROM tenants t LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO sermon_notes (tenant_id, title, speaker, sermon_date, scripture_reference, notes_text, series_name, video_url, published)
SELECT t.id,
       'Finding Rest in a Restless World',
       'Pastor James Mitchell',
       '2026-01-18',
       'Matthew 11:28-30',
       'Jesus invites us: "Come to me, all who are weary and burdened, and I will give you rest." In a culture of hustle and burnout, discover the countercultural invitation to sabbath rest.',
       'Better Together',
       'https://example.com/sermons/finding-rest.mp4',
       true
FROM tenants t LIMIT 1
ON CONFLICT DO NOTHING;
