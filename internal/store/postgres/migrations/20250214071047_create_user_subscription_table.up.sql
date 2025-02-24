DO $$
BEGIN
    -- Create ENUM type if not exists
CREATE TYPE subscription_status AS ENUM ('subscribed', 'unsubscribed');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the audiences table
CREATE TABLE IF NOT EXISTS audiences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    activity VARCHAR(100) NOT NULL,
    status subscription_status NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source VARCHAR(100) NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB,
    UNIQUE(email, activity)
    );

-- Create necessary indexes
CREATE INDEX IF NOT EXISTS audiences_email_idx ON audiences(email);
CREATE INDEX IF NOT EXISTS audiences_activity_idx ON audiences(activity);
CREATE INDEX IF NOT EXISTS audiences_status_idx ON audiences(status);
CREATE INDEX IF NOT EXISTS audiences_email_activity_idx ON audiences(email, activity);

-- Create function to handle timestamps
CREATE OR REPLACE FUNCTION update_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    -- Always update updated_at
    NEW.updated_at := NOW();

    -- Update changed_at only if status changes
    IF NEW.status IS DISTINCT FROM OLD.status THEN
        NEW.changed_at := NOW();
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if it exists
DROP TRIGGER IF EXISTS trigger_update_timestamps ON audiences;

-- Create new trigger
CREATE TRIGGER trigger_update_timestamps
    BEFORE UPDATE ON audiences
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamps();
