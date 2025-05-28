ALTER TABLE billing_customers ALTER COLUMN due_in_days SET DEFAULT 30;
UPDATE billing_customers SET due_in_days = 30;
