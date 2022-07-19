package spicedb

import (
	"context"
	"strings"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/internal/store/spicedb/schema_generator"
)

func (db *SpiceDB) AddPolicy(ctx context.Context, policies []policy.Policy) error {
	schemas, err := db.generateSchema(policies)
	if err != nil {
		return err
	}
	schema := strings.Join(schemas, "\n")
	request := &authzedpb.WriteSchemaRequest{Schema: schema}
	_, err = db.client.WriteSchema(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (db *SpiceDB) generateSchema(policies []policy.Policy) ([]string, error) {
	definitions, err := schema_generator.BuildPolicyDefinitions(policies)
	if err != nil {
		return []string{}, err
	}
	schemas := schema_generator.BuildSchema(definitions)
	schemas = append(schemas, schema_generator.GetDefaultSchema()...)
	return schemas, nil
}
