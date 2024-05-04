CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    description text,
    subscribed_events text[],
    headers jsonb NOT NULL DEFAULT '{}'::jsonb,
    url text NOT NULL,
    secrets text,
    state text NOT NULL DEFAULT 'enabled',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
