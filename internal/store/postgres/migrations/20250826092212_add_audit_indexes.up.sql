-- +migrate Up notransaction

-- Create indexes concurrently to avoid locking the table
CREATE INDEX CONCURRENTLY IF NOT EXISTS auditlogs_idempotency_key_idx
    ON auditlogs (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS auditlogs_actor_id_idx
    ON auditlogs ((actor->>'id'));

CREATE INDEX CONCURRENTLY IF NOT EXISTS auditlogs_resource_id_idx
    ON auditlogs ((resource->>'id'));

CREATE INDEX CONCURRENTLY IF NOT EXISTS auditlogs_resource_type_idx
    ON auditlogs ((resource->>'type'));

CREATE INDEX CONCURRENTLY IF NOT EXISTS auditlogs_occurred_at_idx
    ON auditlogs (occurred_at);