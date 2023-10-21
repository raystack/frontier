CREATE TABLE IF NOT EXISTS billing_customers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id uuid NOT NULL REFERENCES organizations(id),
    provider text NOT NULL,
    provider_id text NOT NULL UNIQUE,
    name text NOT NULL,
    email text NOT NULL,
    phone text,
    address jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    currency text NOT NULL DEFAULT 'usd',
    state text NOT NULL DEFAULT 'active',
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS billing_customers_org_id_idx ON billing_customers(org_id);

CREATE TABLE IF NOT EXISTS billing_plans (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL UNIQUE,
    title text,
    description text NOT NULL,
    interval text,
    state text NOT NULL DEFAULT 'active',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id uuid NOT NULL REFERENCES billing_customers(id),
    provider_id text NOT NULL UNIQUE,
    plan_id uuid NOT NULL REFERENCES billing_plans(id),

    cancel_url text,
    success_url text,
    trial_days int NOT NULL DEFAULT 0,

    state text NOT NULL DEFAULT 'active',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    canceled_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS billing_subscriptions_customer_id_idx ON billing_subscriptions(customer_id);

CREATE TABLE IF NOT EXISTS billing_features (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id uuid,

    name text NOT NULL UNIQUE,
    title text,
    description text NOT NULL,

    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    state text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS billing_features_plan_id_idx ON billing_features(plan_id);

CREATE TABLE IF NOT EXISTS billing_prices (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_id uuid NOT NULL REFERENCES billing_features(id),
    provider_id text NOT NULL UNIQUE,
    name text,
    billing_scheme text NOT NULL,
    currency text NOT NULL DEFAULT 'usd',
    amount bigint NOT NULL DEFAULT 0,
    usage_type text NOT NULL,
    metered_aggregate text,
    tier_mode text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    state text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS billing_prices_feature_id_idx ON billing_prices(feature_id);