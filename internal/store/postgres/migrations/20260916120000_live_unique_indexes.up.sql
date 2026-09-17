CREATE UNIQUE INDEX IF NOT EXISTS uq_organizations_name_live ON organizations (name) WHERE deleted_at IS NULL;
ALTER TABLE organizations DROP CONSTRAINT IF EXISTS organizations_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_projects_name_live ON projects (name) WHERE deleted_at IS NULL;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_live ON users (email) WHERE deleted_at IS NULL;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_name_live ON users (name) WHERE deleted_at IS NULL;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_domains_org_id_name_live ON domains (org_id, name) WHERE deleted_at IS NULL;
ALTER TABLE domains DROP CONSTRAINT IF EXISTS org_id_name_unique;

CREATE UNIQUE INDEX IF NOT EXISTS uq_groups_org_id_name_live ON groups (org_id, name) WHERE deleted_at IS NULL;
ALTER TABLE groups DROP CONSTRAINT IF EXISTS groups_org_id_name_key;
