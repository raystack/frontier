CREATE TABLE IF NOT EXISTS groups
(
    id         uuid        PRIMARY KEY     DEFAULT uuid_generate_v4(),
    name       varchar     UNIQUE NOT NULL,
    slug       varchar     UNIQUE NOT NULL,
    org_id     uuid                        REFERENCES organizations(id),
    metadata   jsonb,
    version    int         NOT NULL        DEFAULT 0,
    created_at timestamptz NOT NULL        DEFAULT NOW(),
    updated_at timestamptz NOT NULL        DEFAULT NOW(),
    deleted_at timestamptz
);
