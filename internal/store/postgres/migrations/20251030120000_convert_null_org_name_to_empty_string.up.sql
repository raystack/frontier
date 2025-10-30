-- Temporarily drop immutability trigger to allow update
DROP TRIGGER IF EXISTS trg_audit_records_enforce_immutability ON audit_records;

-- Convert all NULL org_name values to empty strings
UPDATE audit_records
SET org_name = ''
WHERE org_name IS NULL;

-- Recreate immutability trigger
CREATE TRIGGER trg_audit_records_enforce_immutability
    BEFORE UPDATE ON audit_records
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_record_updates();

-- Update partial index condition to exclude both NULL and empty strings
DROP INDEX IF EXISTS idx_audit_records_org_name;
CREATE INDEX idx_audit_records_org_name
    ON audit_records(org_name, occurred_at DESC)
    WHERE org_name IS NOT NULL AND org_name != '';