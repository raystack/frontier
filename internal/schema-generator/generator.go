package schema_generator

import (
	v0 "github.com/authzed/authzed-go/proto/authzed/api/v0"
	"github.com/authzed/spicedb/pkg/namespace"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
)

type role struct {
	Name       string
	Type       string
	Permission []string
}

type definition struct {
	Name  string
	Roles []role
}

type Policy struct {
	Namespace  string
	Role       string
	RoleType   string
	Permission string
}

func build_schema(d definition) string {
	relations := []*v0.Relation{}
	permissions := make(map[string][]*v0.SetOperation_Child)

	for _, r := range d.Roles {
		relations = append(relations, namespace.Relation(
			r.Name,
			nil,
			&v0.RelationReference{
				Namespace: r.Type,
				Relation:  "...",
			},
		))
		for _, p := range r.Permission {
			permissions[p] = append(permissions[p], namespace.ComputedUserset(r.Name))

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

func build_policy_definitions(policies []Policy) []definition {
	definitions := []definition{}
	def_map := make(map[string]map[string][]role)

	for _, p := range policies {
		def, ok := def_map[p.Namespace]
		if !ok {
			def = make(map[string][]role)
			def_map[p.Namespace] = def
		}

		r, ok := def[p.Role]
		if !ok {
			r = []role{}
			def[p.Role] = r
		}
		def[p.Role] = append(r, role{
			Name:       p.Role,
			Type:       p.RoleType,
			Permission: []string{p.Permission},
		})
	}

	for namespace, def := range def_map {
		roles := []role{}
		for _, r := range def {
			permissions := []string{}
			for _, p := range r {
				permissions = append(permissions, p.Permission...)
			}
			roles = append(roles, role{
				Name:       r[0].Name,
				Type:       r[0].Type,
				Permission: permissions,
			})
		}
		definition := definition{
			Name:  namespace,
			Roles: roles,
		}

		definitions = append(definitions, definition)
	}

	return definitions
}
