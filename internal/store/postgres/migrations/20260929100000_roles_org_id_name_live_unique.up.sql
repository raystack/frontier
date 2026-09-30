CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_org_id_name_live ON roles (org_id, name) WHERE deleted_at IS NULL;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_org_id_name_key;
