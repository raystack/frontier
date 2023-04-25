ALTER TABLE users DROP COLUMN IF EXISTS state;
ALTER TABLE projects DROP COLUMN IF EXISTS state;
ALTER TABLE groups DROP COLUMN IF EXISTS state;
ALTER TABLE organizations DROP COLUMN IF EXISTS state;

DROP INDEX IF EXISTS idx_users_state;
DROP INDEX IF EXISTS idx_projects_state;
DROP INDEX IF EXISTS idx_groups_state;
DROP INDEX IF EXISTS idx_organizations_state;