-- handle credits in plans
ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS on_start_credits bigint NOT NULL DEFAULT 0;

-- remove credit_amount column from billing_products table
ALTER TABLE billing_products DROP COLUMN IF EXISTS credit_amount;

-- add configuration column to billing_products table
ALTER TABLE billing_products ADD COLUMN IF NOT EXISTS config jsonb NOT NULL DEFAULT '{}'::jsonb;
