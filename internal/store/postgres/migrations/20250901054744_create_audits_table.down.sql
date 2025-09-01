-- Drop indexes first
DROP INDEX IF EXISTS idx_audit_records_idempotency_key;
DROP INDEX IF EXISTS idx_audit_records_target;
DROP INDEX IF EXISTS idx_audit_records_event;
DROP INDEX IF EXISTS idx_audit_records_resource;
DROP INDEX IF EXISTS idx_audit_records_actor_id;
DROP INDEX IF EXISTS idx_audit_records_organization_id;
DROP INDEX IF EXISTS idx_audit_records_occurred_at;

-- Drop the table
DROP TABLE IF EXISTS audit_records;

-- Drop the UUIDv7 function
DROP FUNCTION IF EXISTS uuid_generate_v7();