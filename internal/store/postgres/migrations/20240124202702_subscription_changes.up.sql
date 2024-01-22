-- add change in subscriptions table
ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS changes jsonb NOT NULL DEFAULT '{}'::jsonb;
