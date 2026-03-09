CREATE TABLE IF NOT EXISTS user_pats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    secret_hash TEXT NOT NULL UNIQUE,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Title uniqueness only for active (non-deleted) PATs.
-- Allows reuse of title after soft-delete (revoke then recreate).
-- Also serves as composite index for CountActive, List, title availability via (user_id, org_id) prefix.
CREATE UNIQUE INDEX idx_user_pats_unique_title ON user_pats(user_id, org_id, title) WHERE deleted_at IS NULL;

CREATE INDEX idx_user_pats_org_active ON user_pats(org_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_user_pats_user_org_active ON user_pats(user_id, org_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_user_pats_expires_active ON user_pats(expires_at) WHERE deleted_at IS NULL;

-- Create a function to handle 'updated_at' timestamp
CREATE OR REPLACE FUNCTION update_user_pats_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop the trigger if it exists
DROP TRIGGER IF EXISTS trigger_update_user_pats_updated_at ON user_pats;

-- Create a new trigger
CREATE TRIGGER trigger_update_user_pats_updated_at
    BEFORE UPDATE ON user_pats
    FOR EACH ROW
    EXECUTE FUNCTION update_user_pats_updated_at();