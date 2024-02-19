ALTER TABLE billing_subscriptions DROP COLUMN IF EXISTS current_period_start_at;
ALTER TABLE billing_subscriptions DROP COLUMN IF EXISTS current_period_end_at;
ALTER TABLE billing_subscriptions DROP COLUMN IF EXISTS billing_cycle_anchor_at;

DROP TABLE IF EXISTS billing_invoices;