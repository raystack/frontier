-- billing_features
ALTER TABLE billing_features
    RENAME TO billing_products;

DROP INDEX IF EXISTS billing_features_plan_ids_idx;
DROP INDEX IF EXISTS billing_features_provider_id_idx;
CREATE INDEX IF NOT EXISTS billing_products_plan_ids_idx ON billing_products USING GIN(plan_ids);
CREATE INDEX IF NOT EXISTS billing_products_provider_id_idx ON billing_products(provider_id);

-- billing_prices
DROP INDEX IF EXISTS billing_prices_feature_id_idx;
ALTER TABLE billing_prices
    RENAME COLUMN feature_id TO product_id;
CREATE INDEX IF NOT EXISTS billing_prices_product_id_idx ON billing_prices(product_id);
