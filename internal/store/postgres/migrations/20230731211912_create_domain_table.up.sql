CREATE TABLE IF NOT EXISTS domains (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    org_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token text NOT NULL,
    state text NOT NULL DEFAULT 'pending',
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW(),
    CONSTRAINT org_id_name_unique UNIQUE (org_id, name)
);
CREATE INDEX IF NOT EXISTS domains_org_id_idx ON domains(org_id);