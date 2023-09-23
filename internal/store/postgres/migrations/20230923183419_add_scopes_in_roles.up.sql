ALTER TABLE roles ADD COLUMN scopes text[];
CREATE INDEX roles_scopes_idx ON roles USING gin(scopes);