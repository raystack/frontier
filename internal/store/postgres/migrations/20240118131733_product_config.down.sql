-- products
ALTER TABLE billing_products DROP COLUMN IF EXISTS config;
ALTER TABLE billing_products ADD COLUMN IF NOT EXISTS credit_amount bigint;

-- plans
ALTER TABLE billing_plans DROP COLUMN IF EXISTS on_start_credits;
