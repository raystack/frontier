ALTER TABLE billing_plans DROP COLUMN IF EXISTS trial_days;

ALTER TABLE billing_subscriptions DROP COLUMN IF EXISTS trial_ends_at;
ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS trial_days int8;
