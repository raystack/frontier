-- Drop index for org_name
DROP INDEX IF EXISTS idx_audit_records_org_name;

-- Remove org_name column from audit_records table
ALTER TABLE audit_records DROP COLUMN IF EXISTS org_name;