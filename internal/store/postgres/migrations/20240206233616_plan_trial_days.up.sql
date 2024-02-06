ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS trial_days int8;

ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS trial_ends_at timestamptz;
ALTER TABLE billing_subscriptions DROP COLUMN IF EXISTS trial_days;