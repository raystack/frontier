ALTER TABLE relations
    ADD COLUMN namespace_id VARCHAR REFERENCES namespaces (id),
    ADD CONSTRAINT unique_relation_with_ns_id UNIQUE (subject_namespace_id, subject_id, role_id, object_namespace_id, object_id, namespace_id),
    DROP CONSTRAINT IF EXISTS relations_subject_namespace_id_subject_id_role_id_object_na_key;
