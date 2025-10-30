-- Temporarily drop immutability trigger to allow update
DROP TRIGGER IF EXISTS trg_audit_records_enforce_immutability ON audit_records;

-- Convert empty string org_name values back to NULL
UPDATE audit_records
SET org_name = NULL
WHERE org_name = '';

-- Recreate immutability trigger
CREATE TRIGGER trg_audit_records_enforce_immutability
    BEFORE UPDATE ON audit_records
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_record_updates();

-- Restore partial index condition to original
DROP INDEX IF EXISTS idx_audit_records_org_name;
CREATE INDEX idx_audit_records_org_name
    ON audit_records(org_name, occurred_at DESC)
    WHERE org_name IS NOT NULL;