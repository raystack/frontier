CREATE TABLE IF NOT EXISTS flows(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    method varchar,
    start_url varchar,
    finish_url varchar,
    nonce varchar,
    created_at timestamptz NOT NULL DEFAULT NOW()
);