CREATE TYPE subscription_status AS ENUM ('subscribed', 'unsubscribed');

CREATE TABLE IF NOT EXISTS audiences (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        name VARCHAR(255),
        email VARCHAR(255) NOT NULL,
        phone VARCHAR(20),
        activity VARCHAR(100) NOT NULL,
        status subscription_status NOT NULL DEFAULT 'unsubscribed',
        changed_at timestamptz NOT NULL DEFAULT NOW(),
        source VARCHAR(100) NOT NULL,
        verified BOOLEAN DEFAULT FALSE,
        created_at timestamptz NOT NULL DEFAULT NOW(),
        updated_at timestamptz NOT NULL DEFAULT NOW(),
        metadata JSONB,
        UNIQUE(email, activity)
);

CREATE INDEX IF NOT EXISTS audiences_email_idx ON audiences(email);
CREATE INDEX IF NOT EXISTS audiences_activity_idx ON audiences(activity);
CREATE INDEX IF NOT EXISTS audiences_status_idx ON audiences(status);