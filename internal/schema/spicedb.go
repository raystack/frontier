package schema

import (
	sdbcore "github.com/authzed/authzed-go/proto/authzed/api/v0"
	sdbnamespace "github.com/authzed/spicedb/pkg/namespace"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
)

func GenerateSchema(namespaceConfig NamespaceConfigMapType) []string {
	definitionSchemaStringified := make([]string, 0)
	for name, config := range namespaceConfig {
		roles := make([]*sdbcore.Relation, 0)
		permissions := make([]*sdbcore.Relation, 0)
		inheritedNamespaces := make([]*sdbcore.Relation, 0)

		// generate spicedb relations
		for roleName, principals := range config.Roles {
			relationList := make([]*sdbcore.RelationReference, 0)
			for _, p := range principals {
				relationList = append(relationList, sdbnamespace.RelationReference(p, "..."))
			}

			roles = append(roles, sdbnamespace.Relation(roleName, nil, relationList...))
		}

		// generate spicedb permissions
		for permissioName, permissionRoles := range config.Permissions {
			rolesList := make([]*sdbcore.SetOperation_Child, 0)
			for _, role := range permissionRoles {
				rolesList = append(rolesList, sdbnamespace.ComputedUserset(role))
			}

			permissions = append(permissions, sdbnamespace.Relation(permissioName, sdbnamespace.Union(rolesList[0], rolesList[1:]...)))
		}

		// generate inheritance
		for _, namespace := range config.InheritedNamespaces {
			inheritedNamespaces = append(inheritedNamespaces, sdbnamespace.Relation(namespace, nil, sdbnamespace.RelationReference(namespace, "...")))
		}

		source, _ := generator.GenerateSource(sdbnamespace.Namespace(name, append(roles, append(permissions, inheritedNamespaces...)...)...))
		definitionSchemaStringified = append(definitionSchemaStringified, source)
	}

	return definitionSchemaStringified
}
