ALTER TABLE users ADD COLUMN IF NOT EXISTS state varchar DEFAULT 'enabled';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS state varchar DEFAULT 'enabled';
ALTER TABLE groups ADD COLUMN IF NOT EXISTS state varchar DEFAULT 'enabled';
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS state varchar DEFAULT 'enabled';

CREATE INDEX IF NOT EXISTS idx_users_state ON users(state);
CREATE INDEX IF NOT EXISTS idx_projects_state ON projects(state);
CREATE INDEX IF NOT EXISTS idx_groups_state ON groups(state);
CREATE INDEX IF NOT EXISTS idx_organizations_state ON organizations(state);
