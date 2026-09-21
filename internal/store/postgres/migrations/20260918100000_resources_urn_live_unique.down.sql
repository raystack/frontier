ALTER TABLE resources ADD CONSTRAINT resources_urn_key UNIQUE (urn);
DROP INDEX IF EXISTS uq_resources_urn_live;
