ALTER TABLE projects ADD CONSTRAINT IF NOT EXISTS projects_name_key UNIQUE (name);
ALTER TABLE groups ADD CONSTRAINT IF NOT EXISTS groups_name_key UNIQUE (name);
ALTER TABLE organizations ADD CONSTRAINT IF NOT EXISTS organizations_name_key UNIQUE (name);