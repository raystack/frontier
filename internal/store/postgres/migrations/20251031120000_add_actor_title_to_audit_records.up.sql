-- Add actor_title column to audit_records table
ALTER TABLE audit_records ADD COLUMN actor_title VARCHAR(255);

-- Temporarily drop immutability trigger to allow update
DROP TRIGGER IF EXISTS trg_audit_records_enforce_immutability ON audit_records;

-- Backfill actor_title for user principals from users table
UPDATE audit_records ar
SET actor_title = COALESCE(u.title, '')
FROM users u
WHERE ar.actor_id = u.id
  AND ar.actor_title IS NULL;

-- Backfill actor_title for service user principals from service users table
UPDATE audit_records ar
SET actor_title = COALESCE(su.title, '')
FROM serviceusers su
WHERE ar.actor_id = su.id
  AND ar.actor_title IS NULL;

-- Set NULL values to empty string
UPDATE audit_records
SET actor_title = ''
WHERE actor_title IS NULL;

-- Recreate immutability trigger
CREATE TRIGGER trg_audit_records_enforce_immutability
    BEFORE UPDATE ON audit_records
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_record_updates();