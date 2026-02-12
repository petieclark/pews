-- Add unique constraint on campaign_recipients for ON CONFLICT clause
CREATE UNIQUE INDEX IF NOT EXISTS idx_campaign_recipients_campaign_person
    ON campaign_recipients(campaign_id, person_id);
