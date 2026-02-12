-- Add media URL columns to songs table for rehearsal playback
ALTER TABLE songs ADD COLUMN IF NOT EXISTS youtube_url TEXT;
ALTER TABLE songs ADD COLUMN IF NOT EXISTS spotify_url TEXT;
ALTER TABLE songs ADD COLUMN IF NOT EXISTS apple_music_url TEXT;
ALTER TABLE songs ADD COLUMN IF NOT EXISTS rehearsal_url TEXT;
