ALTER TABLE relations
    ADD COLUMN namespace_id VARCHAR REFERENCES namespaces (id),
    DROP CONSTRAINT IF EXISTS relations_subject_namespace_id_subject_id_role_id_object_na_key;

CREATE UNIQUE INDEX unique_relation_with_ns_id ON relations (subject_namespace_id, subject_id, object_namespace_id, object_id, COALESCE(role_id, ''), COALESCE(namespace_id, ''))
