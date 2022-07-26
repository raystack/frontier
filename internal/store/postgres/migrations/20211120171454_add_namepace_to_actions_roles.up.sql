ALTER TABLE roles
    ADD COLUMN namespace_id VARCHAR REFERENCES namespaces(id);

ALTER TABLE actions
    ADD COLUMN namespace_id VARCHAR REFERENCES namespaces(id);