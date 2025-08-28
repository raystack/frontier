DROP INDEX CONCURRENTLY IF EXISTS auditlogs_occurred_at_idx;
DROP INDEX CONCURRENTLY IF EXISTS auditlogs_resource_type_idx;
DROP INDEX CONCURRENTLY IF EXISTS auditlogs_resource_id_idx;
DROP INDEX CONCURRENTLY IF EXISTS auditlogs_actor_id_idx;
DROP INDEX CONCURRENTLY IF EXISTS auditlogs_idempotency_key_idx;

ALTER TABLE auditlogs
    DROP COLUMN IF EXISTS occurred_at,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS resource,
    DROP COLUMN IF EXISTS idempotency_key;