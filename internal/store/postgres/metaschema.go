package postgres

import (
	"github.com/odpf/shield/core/metaschema"
)

type MetaSchema struct {
	Name   string
	Schema string
}

func (from MetaSchema) tranformtoMetadataSchema() metaschema.MetaSchema {
	return metaschema.MetaSchema{
		Name:   from.Name,
		Schema: from.Schema,
	}
}
