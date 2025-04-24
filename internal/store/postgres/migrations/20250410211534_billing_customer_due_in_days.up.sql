ALTER TABLE billing_customers ADD COLUMN IF NOT EXISTS due_in_days integer DEFAULT 0;
