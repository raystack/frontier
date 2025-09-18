DROP TRIGGER IF EXISTS trg_audit_records_enforce_immutability ON audit_records;

DROP FUNCTION IF EXISTS prevent_audit_record_updates();