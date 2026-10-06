ALTER TABLE policies ADD CONSTRAINT policies_role_id_resource_id_resource_type_principal_id_pri_key UNIQUE (role_id, resource_id, resource_type, principal_id, principal_type);
DROP INDEX IF EXISTS uq_policies_role_resource_principal_live;

ALTER TABLE relations ADD CONSTRAINT relations_subject_namespace_name_subject_id_object_namespac_key UNIQUE (subject_namespace_name, subject_id, object_namespace_name, object_id, relation_name);
DROP INDEX IF EXISTS uq_relations_subject_object_relation_live;
