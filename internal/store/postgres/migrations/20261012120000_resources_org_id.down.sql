DROP INDEX IF EXISTS idx_resources_org_id;
ALTER TABLE resources DROP COLUMN IF EXISTS org_id;
