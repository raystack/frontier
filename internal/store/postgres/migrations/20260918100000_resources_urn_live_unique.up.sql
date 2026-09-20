CREATE UNIQUE INDEX IF NOT EXISTS uq_resources_urn_live ON resources (urn) WHERE deleted_at IS NULL;
ALTER TABLE resources DROP CONSTRAINT IF EXISTS resources_urn_key;
