DROP TRIGGER IF EXISTS trg_audit_records_enforce_immutability ON audit_records;

DROP FUNCTION IF EXISTS prevent_audit_record_updates();

CREATE OR REPLACE FUNCTION prevent_audit_record_updates()
  RETURNS TRIGGER AS $$
BEGIN
      RAISE EXCEPTION 'audit_records cannot be updated to maintain audit integrity'
          USING ERRCODE = '45000',  -- User-defined error (Postgres convention: user-defined error codes are in the 45000-45999 range)
              DETAIL = 'Audit records are immutable once created';
END;
  $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_records_enforce_immutability
    BEFORE UPDATE ON audit_records
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_record_updates();

COMMENT ON TRIGGER trg_audit_records_enforce_immutability ON audit_records IS
    'Enforces immutability of audit records by preventing any UPDATE operation.';