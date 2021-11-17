package schema_generator

import (
	"fmt"

	v0 "github.com/authzed/authzed-go/proto/authzed/api/v0"
	"github.com/authzed/spicedb/pkg/namespace"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
)

type role struct {
	Name       string
	Type       string
	Namespace  string
	Permission []string
}

type definition struct {
	Name  string
	Roles []role
}

type Policy struct {
	Namespace     string
	Role          string
	RoleType      string
	RoleNamespace string
	Permission    string
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

func build_policy_definitions(policies []Policy) []definition {
	definitions := []definition{}
	def_map := make(map[string]map[string][]role)

	for _, p := range policies {
		def, ok := def_map[p.Namespace]
		if !ok {
			def = make(map[string][]role)
			def_map[p.Namespace] = def
		}

		keyName := fmt.Sprintf("%s_%s_%s", p.RoleNamespace, p.Role, p.RoleType)

		r, ok := def[keyName]
		if !ok {
			r = []role{}
			def[keyName] = r
		}

		def[keyName] = append(r, role{
			Name:       p.Role,
			Type:       p.RoleType,
			Namespace:  p.RoleNamespace,
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

			role_namespace := namespace

			if r[0].Namespace != "" {
				role_namespace = r[0].Namespace
			}

			roles = append(roles, role{
				Name:       r[0].Name,
				Type:       r[0].Type,
				Namespace:  role_namespace,
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
