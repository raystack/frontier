CREATE TABLE IF NOT EXISTS user_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    secret_hash TEXT NOT NULL UNIQUE,
    metadata JSONB DEFAULT '{}',
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Title uniqueness only for active (non-deleted) tokens.
-- Allows reuse of title after soft-delete (revoke then recreate).
-- Also serves as composite index for CountActive, List, title availability via (user_id, org_id) prefix.
CREATE UNIQUE INDEX idx_user_tokens_unique_title ON user_tokens(user_id, org_id, title) WHERE deleted_at IS NULL;

CREATE INDEX idx_user_tokens_org_active ON user_tokens(org_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_user_tokens_expires_active ON user_tokens(expires_at) WHERE deleted_at IS NULL;

-- Create a function to handle 'updated_at' timestamp
CREATE OR REPLACE FUNCTION update_user_tokens_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop the trigger if it exists
DROP TRIGGER IF EXISTS trigger_update_user_tokens_updated_at ON user_tokens;

-- Create a new trigger
CREATE TRIGGER trigger_update_user_tokens_updated_at
    BEFORE UPDATE ON user_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_user_tokens_updated_at();