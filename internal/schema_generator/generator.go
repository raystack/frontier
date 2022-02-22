package schema_generator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/odpf/shield/pkg/utils"

	v0 "github.com/authzed/authzed-go/proto/authzed/api/v0"
	"github.com/authzed/spicedb/pkg/namespace"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
	"github.com/odpf/shield/model"
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

func BuildSchema(def []definition) []string {
	var schema []string
	for _, d := range def {
		schema = append(schema, buildSchema(d))
	}
	return schema
}

func GetDefaultSchema() []string {
	userSchema := `definition user {}`
	schemas := []string{userSchema}
	return schemas
}

func buildSchema(d definition) string {
	var relations []*v0.Relation
	permissionsMap := make(map[string][]*v0.SetOperation_Child)
	permissionsSlice := make([]string, 0)

	inheritedNSMap := map[string]bool{}

	for _, r := range d.roles {
		if r.namespace == "" || r.namespace == d.name {
			relationReference := buildRelationReference(r)
			relations = append(relations, namespace.Relation(
				r.name,
				nil,
				relationReference...,
			))
		}

		for _, p := range r.permissions {
			perm := namespace.ComputedUserset(r.name)
			if r.namespace != "" && r.namespace != d.name {
				perm = namespace.TupleToUserset(r.namespace, r.name)
				if !inheritedNSMap[r.namespace] {
					relations = append(relations, namespace.Relation(
						r.namespace,
						nil,
						&v0.RelationReference{
							Namespace: r.namespace,
							Relation:  "...",
						},
					))
					inheritedNSMap[r.namespace] = true
				}
			}
			if _, ok := permissionsMap[p]; !ok {
				permissionsSlice = append(permissionsSlice, p)
			}
			permissionsMap[p] = append(permissionsMap[p], perm)
			sort.Slice(permissionsMap[p], func(i, j int) bool {
				return permissionsMap[p][i].String() > permissionsMap[p][j].String()
			})
		}
	}

	for _, p := range permissionsSlice {
		roles := permissionsMap[p]
		if len(roles) == 0 {
			continue
		}
		sort.Slice(roles, func(i, j int) bool {
			return roles[i].String() > roles[j].String()
		})
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

func BuildPolicyDefinitions(policies []model.Policy) ([]definition, error) {
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

		actionNs := utils.DefaultStringIfEmpty(p.Action.Namespace.Id, p.Action.NamespaceId)
		actionId := utils.DefaultStringIfEmpty(p.Action.Id, p.ActionId)
		if actionNs != "" && actionNs != namespaceId {
			return []definition{}, errors.New(fmt.Sprintf("actions (%s) namespace `%s` doesnt match with `%s`", actionId, actionNs, namespaceId))
		}

		def[keyName] = append(r, role{
			name:        p.Role.Id,
			types:       p.Role.Types,
			namespace:   p.Role.NamespaceId,
			permissions: []string{actionId},
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
				name:        strings.ReplaceAll(r[0].name, "-", "_"),
				types:       r[0].types,
				namespace:   strings.ReplaceAll(roleNamespace, "-", "_"),
				permissions: permissions,
			})
		}

		definition := definition{
			name:  strings.ReplaceAll(ns, "-", "_"),
			roles: roles,
		}

		sort.Slice(roles[:], func(i, j int) bool {
			return strings.Compare(roles[i].name, roles[j].name) < 1 && strings.Compare(roles[i].namespace, roles[j].namespace) < 1
		})

		definitions = append(definitions, definition)
	}

	sort.Slice(definitions[:], func(i, j int) bool {
		return strings.Compare(definitions[i].name, definitions[j].name) < 1
	})
	return definitions, nil
}
