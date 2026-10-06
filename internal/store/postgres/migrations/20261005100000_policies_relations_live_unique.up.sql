CREATE UNIQUE INDEX IF NOT EXISTS uq_policies_role_resource_principal_live ON policies (role_id, resource_id, resource_type, principal_id, principal_type) WHERE deleted_at IS NULL;
ALTER TABLE policies DROP CONSTRAINT IF EXISTS policies_role_id_resource_id_resource_type_principal_id_pri_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_relations_subject_object_relation_live ON relations (subject_namespace_name, subject_id, object_namespace_name, object_id, relation_name) WHERE deleted_at IS NULL;
ALTER TABLE relations DROP CONSTRAINT IF EXISTS relations_subject_namespace_name_subject_id_object_namespac_key;
