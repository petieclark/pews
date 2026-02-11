-- Add TOTP 2FA fields to users table
ALTER TABLE users 
    ADD COLUMN totp_secret VARCHAR(255),
    ADD COLUMN totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN recovery_codes TEXT[];

-- Index for 2FA enabled users
CREATE INDEX idx_users_totp_enabled ON users(totp_enabled) WHERE totp_enabled = TRUE;
