ALTER TABLE relations
    DROP COLUMN IF EXISTS namespace_id,
    DROP CONSTRAINT IF EXISTS unique_relation_with_ns_id,
    ADD CONSTRAINT relations_subject_namespace_id_subject_id_role_id_object_na_key UNIQUE (subject_namespace_id, subject_id, role_id, object_namespace_id, object_id);