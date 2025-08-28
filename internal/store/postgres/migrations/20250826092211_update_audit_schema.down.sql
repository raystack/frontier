ALTER TABLE auditlogs
    DROP COLUMN IF EXISTS occurred_at,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS resource,
    DROP COLUMN IF EXISTS idempotency_key;