ALTER TABLE auditlogs
    ADD COLUMN IF NOT EXISTS idempotency_key uuid, -- dont make it unique constraint
    ADD COLUMN IF NOT EXISTS resource jsonb DEFAULT '{}' NOT NULL,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS occurred_at timestamp with time zone DEFAULT now() NOT NULL;

UPDATE auditlogs
SET occurred_at = created_at
WHERE occurred_at IS NULL;

CREATE INDEX CONCURRENTLY auditlogs_idempotency_key_idx
    ON auditlogs (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX CONCURRENTLY auditlogs_actor_id_idx
    ON auditlogs ((actor->>'id'));

CREATE INDEX CONCURRENTLY auditlogs_resource_id_idx
    ON auditlogs ((resource->>'id'));

CREATE INDEX CONCURRENTLY auditlogs_resource_type_idx
    ON auditlogs ((resource->>'type'));

CREATE INDEX CONCURRENTLY auditlogs_occurred_at_idx
    ON auditlogs (occurred_at);