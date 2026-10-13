ALTER TABLE resources ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations (id);

UPDATE resources r
SET org_id = p.org_id
FROM projects p
WHERE r.project_id = p.id
  AND r.org_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_resources_org_id ON resources (org_id);
