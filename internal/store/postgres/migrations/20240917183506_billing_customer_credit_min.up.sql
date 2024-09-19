ALTER TABLE billing_customers ADD COLUMN IF NOT EXISTS credit_min int8 DEFAULT 0;
