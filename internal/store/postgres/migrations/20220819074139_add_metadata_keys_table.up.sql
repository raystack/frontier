CREATE TABLE IF NOT EXISTS metadata_keys (
    id           uuid        PRIMARY KEY     DEFAULT uuid_generate_v4(),
    key          varchar     UNIQUE NOT NULL,
    description  varchar,
    created_at   timestamptz NOT NULL        DEFAULT NOW(),
    updated_at   timestamptz NOT NULL        DEFAULT NOW()
);