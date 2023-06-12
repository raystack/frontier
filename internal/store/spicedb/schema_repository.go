package spicedb

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/salt/log"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

type SchemaRepository struct {
	spiceDB *SpiceDB
	logger  log.Logger
}

var (
	ErrWritingSchema = errors.New("error in writing schema to spicedb")
)

func NewSchemaRepository(logger log.Logger, spiceDB *SpiceDB) *SchemaRepository {
	return &SchemaRepository{
		spiceDB: spiceDB,
		logger:  logger,
	}
}

func (r SchemaRepository) WriteSchema(ctx context.Context, schema string) error {
	if r.logger.Level() == "debug" {
		fmt.Println(schema)
	}
	if _, err := r.spiceDB.client.WriteSchema(ctx, &authzedpb.WriteSchemaRequest{Schema: schema}); err != nil {
		return fmt.Errorf("%w: %s", ErrWritingSchema, err.Error())
	}
	return nil
}
