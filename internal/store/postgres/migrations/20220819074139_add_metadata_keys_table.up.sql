CREATE TABLE IF NOT EXISTS metadata_keys (
    key          varchar     UNIQUE NOT NULL,
    description  varchar,
    created_at   timestamptz NOT NULL        DEFAULT NOW(),
    updated_at   timestamptz NOT NULL        DEFAULT NOW()
);
