CREATE TABLE IF NOT EXISTS metaschema (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar UNIQUE NOT NULL,
    schema varchar NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);