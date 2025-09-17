-- Create UUIDv7 function for time-ordered IDs (portable implementation)
-- Reference: https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
-- Note: Requires uuid-ossp or pgcrypto extension for gen_random_uuid() (should already exist from base migration)
-- Uses random v4 UUID as a starting point, then overlays timestamp and sets version 7
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID AS $$
BEGIN
  RETURN encode(
    set_bit(
      set_bit(
        overlay(uuid_send(gen_random_uuid())
                placing substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
                from 1 for 6
        ),
        52, 1
      ),
      53, 1
    ),
    'hex')::uuid;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE audit_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    idempotency_key UUID UNIQUE, -- nullable for Frontier's internal calls.
    event VARCHAR(255) NOT NULL,
    actor_id UUID NOT NULL,
    actor_type VARCHAR(50) NOT NULL,
    actor_name VARCHAR(255) NOT NULL,
    actor_metadata JSONB DEFAULT '{}',
    resource_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    resource_metadata JSONB DEFAULT '{}',
    target_id VARCHAR(255),
    target_type VARCHAR(100),
    target_name VARCHAR(255),
    target_metadata JSONB DEFAULT '{}',
    organization_id UUID,
    request_id VARCHAR(255),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'
);

-- Primary indexes for common query patterns
CREATE INDEX idx_audit_records_occurred_at 
    ON audit_records(occurred_at DESC);

CREATE INDEX idx_audit_records_organization_id 
    ON audit_records(organization_id, occurred_at DESC);

CREATE INDEX idx_audit_records_actor_id
    ON audit_records(actor_id, occurred_at DESC);

-- Resource filtering
CREATE INDEX idx_audit_records_resource 
    ON audit_records(resource_id, resource_type, occurred_at DESC);

-- Event filtering
CREATE INDEX idx_audit_records_event 
    ON audit_records(event, occurred_at DESC);

-- Target resource queries
CREATE INDEX idx_audit_records_target 
    ON audit_records(target_id, target_type) 
    WHERE target_id IS NOT NULL;

-- Idempotency key index for deduplication
CREATE INDEX idx_audit_records_idempotency_key 
    ON audit_records(idempotency_key) 
    WHERE idempotency_key IS NOT NULL;