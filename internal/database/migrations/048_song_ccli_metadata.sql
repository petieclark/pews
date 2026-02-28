-- Add CCLI licensing metadata fields to songs table
ALTER TABLE songs ADD COLUMN IF NOT EXISTS authors TEXT[];
ALTER TABLE songs ADD COLUMN IF NOT EXISTS copyright_year INTEGER;
ALTER TABLE songs ADD COLUMN IF NOT EXISTS publisher VARCHAR(255);
ALTER TABLE songs ADD COLUMN IF NOT EXISTS license_type VARCHAR(50);

-- Create indexes for CCLI search and filtering
CREATE INDEX IF NOT EXISTS idx_songs_ccli_number ON songs(ccli_number) WHERE ccli_number IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_songs_license_type ON songs(license_type) WHERE license_type IS NOT NULL;
