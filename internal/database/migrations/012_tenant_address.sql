-- Add address and EIN fields to tenants table for tax statements
ALTER TABLE tenants
ADD COLUMN IF NOT EXISTS address_line1 VARCHAR(255),
ADD COLUMN IF NOT EXISTS address_line2 VARCHAR(255),
ADD COLUMN IF NOT EXISTS city VARCHAR(100),
ADD COLUMN IF NOT EXISTS state VARCHAR(2),
ADD COLUMN IF NOT EXISTS zip VARCHAR(10),
ADD COLUMN IF NOT EXISTS ein VARCHAR(20);

-- Add comment for EIN field
COMMENT ON COLUMN tenants.ein IS 'Employer Identification Number for tax statements';
