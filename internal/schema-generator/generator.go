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
