-- billing_features to billing_products
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

-- billing_features

CREATE TABLE IF NOT EXISTS billing_features (
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE,
    product_ids text[],

    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);
CREATE INDEX IF NOT EXISTS billing_features_product_ids_idx ON billing_features USING GIN(product_ids);
CREATE INDEX IF NOT EXISTS billing_features_name_idx ON billing_features(name);