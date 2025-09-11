-- Rename fields in the audit_records table
ALTER TABLE audit_records RENAME COLUMN request_id TO req_id;
ALTER TABLE audit_records RENAME COLUMN organization_id TO org_id;