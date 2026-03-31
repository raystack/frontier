ALTER TABLE billing_customers DROP COLUMN IF EXISTS payment_mode;
DROP TYPE IF EXISTS payment_mode_type;
