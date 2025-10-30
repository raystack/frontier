-- Add org_name column to audit_records table
ALTER TABLE audit_records ADD COLUMN org_name VARCHAR(255);

-- Temporarily drop immutability trigger to allow backfill
DROP TRIGGER IF EXISTS trg_audit_records_enforce_immutability ON audit_records;

-- Backfill org_name for existing records by joining with organizations table
UPDATE audit_records ar
SET org_name = o.name
FROM organizations o
WHERE ar.org_id = o.id
  AND ar.org_name IS NULL;

-- Recreate immutability trigger
CREATE TRIGGER trg_audit_records_enforce_immutability
    BEFORE UPDATE ON audit_records
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_record_updates();

-- Create index on org_name
CREATE INDEX idx_audit_records_org_name
    ON audit_records(org_name, occurred_at DESC)
    WHERE org_name IS NOT NULL;