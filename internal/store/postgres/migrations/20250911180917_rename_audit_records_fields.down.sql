-- Revert field renames in audit_records table
ALTER TABLE audit_records RENAME COLUMN req_id TO request_id;
ALTER TABLE audit_records RENAME COLUMN org_id TO organization_id;