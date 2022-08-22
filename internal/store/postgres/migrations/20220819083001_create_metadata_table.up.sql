CREATE TABLE IF NOT EXISTS metadata (
    id           uuid        PRIMARY KEY     DEFAULT uuid_generate_v4(),
    user_id      uuid REFERENCES users (id),
    key          varchar     REFERENCES metadata_keys (key_name),
    created_at   timestamptz NOT NULL        DEFAULT NOW(),
    updated_at   timestamptz NOT NULL        DEFAULT NOW(),
    CONSTRAINT UQ_Metadata UNIQUE (user_id,key)
);
