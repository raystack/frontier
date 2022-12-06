ALTER TABLE relations
ADD CONSTRAINT relations_unique_columns UNIQUE (subject_namespace_id, subject_id, object_namespace_id, object_id, role_id);