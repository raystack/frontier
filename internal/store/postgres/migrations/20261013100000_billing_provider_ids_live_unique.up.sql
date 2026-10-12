CREATE UNIQUE INDEX IF NOT EXISTS uq_billing_invoices_provider_id_live ON billing_invoices (provider_id) WHERE deleted_at IS NULL;
ALTER TABLE billing_invoices DROP CONSTRAINT IF EXISTS billing_invoices_provider_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_billing_subscriptions_provider_id_live ON billing_subscriptions (provider_id) WHERE deleted_at IS NULL;
ALTER TABLE billing_subscriptions DROP CONSTRAINT IF EXISTS billing_subscriptions_provider_id_key;
ALTER TABLE billing_subscriptions DROP CONSTRAINT IF EXISTS billing_subscriptions_provider_id_plan_id_key;
