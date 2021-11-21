package schema_generator

import (
	"fmt"
	"github.com/odpf/shield/model"
	"strings"

	v0 "github.com/authzed/authzed-go/proto/authzed/api/v0"
	"github.com/authzed/spicedb/pkg/namespace"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
)

type role struct {
	name        string
	types       []string
	namespace   string
	permissions []string
}

type definition struct {
	name  string
	roles []role
}

func buildSchema(d definition) string {
	var relations []*v0.Relation
	permissions := make(map[string][]*v0.SetOperation_Child)

	for _, r := range d.roles {

		if r.namespace == "" {
			relationReference := buildRelationReference(r)

			relations = append(relations, namespace.Relation(
				r.name,
				nil,
				relationReference...,
			))
		}

		for _, p := range r.permissions {
			perm := namespace.ComputedUserset(r.name)
			if r.namespace != "" {
				perm = namespace.TupleToUserset(r.namespace, r.name)
				relations = append(relations, namespace.Relation(
					r.namespace,
					nil,
					&v0.RelationReference{
						Namespace: r.namespace,
						Relation:  "...",
					},
				))
			}
			permissions[p] = append(permissions[p], perm)
		}
	}

	for p, roles := range permissions {
		if len(roles) == 0 {
			continue
		}
		relations = append(relations, namespace.Relation(
			p,
			namespace.Union(
				roles[0],
				roles[1:]...,
			),
		))
	}

	n := namespace.Namespace(d.name, relations...)

	schemaDefinition, _ := generator.GenerateSource(n)
	return schemaDefinition
}

func buildRelationReference(r role) []*v0.RelationReference {
	var relationReference []*v0.RelationReference
	for _, t := range r.types {
		roleType := strings.Split(t, "#")
		subType := "..."
		if len(roleType) > 1 {
			subType = roleType[1]
		}
		relationReference = append(relationReference, &v0.RelationReference{
			Namespace: roleType[0],
			Relation:  subType,
		})
	}
	return relationReference
}

func buildPolicyDefinitions(policies []model.Policy) []definition {
	var definitions []definition
	defMap := make(map[string]map[string][]role)

	for _, p := range policies {
		namespaceId := p.Namespace.Id
		def, ok := defMap[namespaceId]
		if !ok {
			def = make(map[string][]role)
			defMap[namespaceId] = def
		}

		keyName := fmt.Sprintf("%s_%s_%s", p.Role.Id, p.Role.NamespaceId, namespaceId)

		r, ok := def[keyName]
		if !ok {
			r = []role{}
			def[keyName] = r
		}

		def[keyName] = append(r, role{
			name:        p.Role.Id,
			types:       p.Role.Types,
			namespace:   p.Role.NamespaceId,
			permissions: []string{p.Action.Id},
		})
	}

	for ns, def := range defMap {
		var roles []role
		for _, r := range def {
			var permissions []string
			for _, p := range r {
				permissions = append(permissions, p.permissions...)
			}

			roleNamespace := ns

			if r[0].namespace != "" {
				roleNamespace = r[0].namespace
			}

			roles = append(roles, role{
				name:        r[0].name,
				types:       r[0].types,
				namespace:   roleNamespace,
				permissions: permissions,
			})
		}
		definition := definition{
			name:  ns,
			roles: roles,
		}

		definitions = append(definitions, definition)
	}

	return definitions
}
