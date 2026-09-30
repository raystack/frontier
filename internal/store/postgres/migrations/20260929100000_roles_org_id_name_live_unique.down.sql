ALTER TABLE roles ADD CONSTRAINT roles_org_id_name_key UNIQUE (org_id, name);
DROP INDEX IF EXISTS uq_roles_org_id_name_live;
