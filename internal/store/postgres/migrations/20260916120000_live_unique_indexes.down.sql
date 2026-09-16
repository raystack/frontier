ALTER TABLE organizations ADD CONSTRAINT organizations_name_key UNIQUE (name);
DROP INDEX IF EXISTS uq_organizations_name_live;

ALTER TABLE projects ADD CONSTRAINT projects_name_key UNIQUE (name);
DROP INDEX IF EXISTS uq_projects_name_live;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
DROP INDEX IF EXISTS uq_users_email_live;

ALTER TABLE users ADD CONSTRAINT users_name_key UNIQUE (name);
DROP INDEX IF EXISTS uq_users_name_live;

ALTER TABLE domains ADD CONSTRAINT org_id_name_unique UNIQUE (org_id, name);
DROP INDEX IF EXISTS uq_domains_org_id_name_live;

ALTER TABLE groups ADD CONSTRAINT groups_org_id_name_key UNIQUE (org_id, name);
DROP INDEX IF EXISTS uq_groups_org_id_name_live;
