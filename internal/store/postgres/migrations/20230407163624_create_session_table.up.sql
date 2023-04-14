CREATE TABLE IF NOT EXISTS sessions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    authenticated_at timestamptz,
    expires_at timestamptz DEFAULT (NOW() + INTERVAL '7 days'),
    created_at timestamptz NOT NULL DEFAULT NOW()
);
