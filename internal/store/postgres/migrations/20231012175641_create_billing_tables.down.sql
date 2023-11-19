DROP TABLE IF EXISTS billing_prices CASCADE;
DROP TABLE IF EXISTS billing_features CASCADE;
DROP TABLE IF EXISTS billing_subscriptions CASCADE;
DROP TABLE IF EXISTS billing_plans CASCADE;
DROP TABLE IF EXISTS billing_checkouts CASCADE;
DROP TABLE IF EXISTS billing_customers;
DROP TABLE IF EXISTS billing_transactions;

DROP INDEX IF EXISTS policies_resource_id_resource_type_idx;