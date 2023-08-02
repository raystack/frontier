CREATE TABLE IF NOT EXISTS domains (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    org_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token text NOT NULL,
    verified boolean NOT NULL DEFAULT false,
    verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT NOW()
);