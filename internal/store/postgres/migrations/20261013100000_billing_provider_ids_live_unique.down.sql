ALTER TABLE billing_subscriptions ADD CONSTRAINT billing_subscriptions_provider_id_plan_id_key UNIQUE (provider_id, plan_id);
ALTER TABLE billing_subscriptions ADD CONSTRAINT billing_subscriptions_provider_id_key UNIQUE (provider_id);
DROP INDEX IF EXISTS uq_billing_subscriptions_provider_id_live;

ALTER TABLE billing_invoices ADD CONSTRAINT billing_invoices_provider_id_key UNIQUE (provider_id);
DROP INDEX IF EXISTS uq_billing_invoices_provider_id_live;
