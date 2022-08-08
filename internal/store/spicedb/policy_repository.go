package spicedb

import (
	"context"
	"strings"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/internal/store/spicedb/schema_generator"
)

type PolicyRepository struct {
	spiceDB *SpiceDB
}

func NewPolicyRepository(spiceDB *SpiceDB) *PolicyRepository {
	return &PolicyRepository{
		spiceDB: spiceDB,
	}
}

func (r PolicyRepository) Add(ctx context.Context, policies []policy.Policy) error {
	schemas, err := generateSchema(policies)
	if err != nil {
		return err
	}
	schema := strings.Join(schemas, "\n")
	request := &authzedpb.WriteSchemaRequest{Schema: schema}
	if _, err = r.spiceDB.client.WriteSchema(ctx, request); err != nil {
		return err
	}
	return nil
}

func generateSchema(policies []policy.Policy) ([]string, error) {
	definitions, err := schema_generator.BuildPolicyDefinitions(policies)
	if err != nil {
		return []string{}, err
	}
	schemas := schema_generator.BuildSchema(definitions)
	schemas = append(schemas, schema_generator.GetDefaultSchema()...)
	return schemas, nil
}
