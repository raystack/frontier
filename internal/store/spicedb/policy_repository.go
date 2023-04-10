package spicedb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/internal/store/spicedb/schema_generator"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

type PolicyRepository struct {
	spiceDB *SpiceDB
}

var (
	ErrWritingSchema = errors.New("error in writing schema to spicedb")
)

func NewPolicyRepository(spiceDB *SpiceDB) *PolicyRepository {
	return &PolicyRepository{
		spiceDB: spiceDB,
	}
}

func (r PolicyRepository) WriteSchema(ctx context.Context, schema schema.NamespaceConfigMapType) error {
	generatedSchema, err := schema_generator.GenerateSchema(schema)
	if err != nil {
		return err
	}
	request := &authzedpb.WriteSchemaRequest{Schema: strings.Join(generatedSchema, "\n")}
	fmt.Println(strings.Join(generatedSchema, "\n"))
	if _, err := r.spiceDB.client.WriteSchema(ctx, request); err != nil {
		return fmt.Errorf("%w: %s", ErrWritingSchema, err.Error())
	}

	return nil
}
