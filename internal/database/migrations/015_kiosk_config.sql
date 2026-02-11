-- Add kiosk configuration to tenants
ALTER TABLE tenants ADD COLUMN kiosk_config JSONB DEFAULT '{
    "enabled": false,
    "quick_amounts": [1000, 2500, 5000, 10000, 25000],
    "default_fund_id": null,
    "thank_you_message": "Thank you for your generous gift!"
}'::jsonb;

-- Index for kiosk enabled lookups
CREATE INDEX idx_tenants_kiosk_enabled ON tenants ((kiosk_config->>'enabled'));
