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
	Name       string
	Types      []string
	Namespace  string
	Permission []string
}

type definition struct {
	Name  string
	Roles []role
}

func build_schema(d definition) string {
	relations := []*v0.Relation{}
	permissions := make(map[string][]*v0.SetOperation_Child)

	for _, r := range d.Roles {

		if r.Namespace == "" {
			relationReference := []*v0.RelationReference{}
			for _, t := range r.Types {
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

			relations = append(relations, namespace.Relation(
				r.Name,
				nil,
				relationReference...,
			))
		}

		for _, p := range r.Permission {
			perm := namespace.ComputedUserset(r.Name)
			if r.Namespace != "" {
				perm = namespace.TupleToUserset(r.Namespace, r.Name)
				relations = append(relations, namespace.Relation(
					r.Namespace,
					nil,
					&v0.RelationReference{
						Namespace: r.Namespace,
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

	n := namespace.Namespace(d.Name, relations...)

	schema_defintion, _ := generator.GenerateSource(n)
	return schema_defintion
}

func build_policy_definitions(policies []model.Policy) []definition {
	definitions := []definition{}
	def_map := make(map[string]map[string][]role)

	for _, p := range policies {
		def, ok := def_map[p.Namespace.Slug]
		if !ok {
			def = make(map[string][]role)
			def_map[p.Namespace.Slug] = def
		}

		keyName := fmt.Sprintf("%s_%s", p.Role.Namespace, p.Role.Id)

		r, ok := def[keyName]
		if !ok {
			r = []role{}
			def[keyName] = r
		}

		def[keyName] = append(r, role{
			Name:       p.Role.Id,
			Types:      p.Role.Types,
			Namespace:  p.Role.Namespace,
			Permission: []string{p.Action.Slug},
		})
	}

	for ns, def := range def_map {
		roles := []role{}
		for _, r := range def {
			permissions := []string{}
			for _, p := range r {
				permissions = append(permissions, p.Permission...)
			}

			role_namespace := ns

			if r[0].Namespace != "" {
				role_namespace = r[0].Namespace
			}

			roles = append(roles, role{
				Name:       r[0].Name,
				Types:      r[0].Types,
				Namespace:  role_namespace,
				Permission: permissions,
			})
		}
		definition := definition{
			Name:  ns,
			Roles: roles,
		}

		definitions = append(definitions, definition)
	}

	return definitions
}
