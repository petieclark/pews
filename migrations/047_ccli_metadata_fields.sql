-- Add CCLI licensing metadata fields to songs table
-- Issue #74

ALTER TABLE songs ADD COLUMN IF NOT EXISTS ccli_number VARCHAR(50);
ALTER TABLE songs ADD COLUMN IF NOT EXISTS authors TEXT[];
ALTER TABLE songs ADD COLUMN IF NOT EXISTS copyright_year INTEGER;
ALTER TABLE songs ADD COLUMN IF NOT EXISTS publisher VARCHAR(255);
ALTER TABLE songs ADD COLUMN IF NOT EXISTS license_type VARCHAR(50);

-- Create index for CCLI number search (bonus requirement)
CREATE INDEX IF NOT EXISTS idx_songs_ccli_number ON songs(ccli_number) WHERE ccli_number IS NOT NULL;
