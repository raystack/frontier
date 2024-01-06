-- billing_features
ALTER TABLE billing_products
    RENAME TO billing_features;

DROP INDEX IF EXISTS billing_products_plan_ids_idx;
DROP INDEX IF EXISTS billing_products_provider_id_idx;
CREATE INDEX IF NOT EXISTS billing_features_plan_ids_idx ON billing_features USING GIN(plan_ids);
CREATE INDEX IF NOT EXISTS billing_features_provider_id_idx ON billing_features(provider_id);

-- billing_prices
DROP INDEX IF EXISTS billing_prices_product_id_idx;
ALTER TABLE billing_prices
    RENAME COLUMN product_id TO feature_id;
CREATE INDEX IF NOT EXISTS billing_prices_feature_id_idx ON billing_prices(feature_id);