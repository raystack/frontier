package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/shield/core/metaschema"
	"github.com/odpf/shield/pkg/db"
	"github.com/xeipuuv/gojsonschema"
)

var (
	userMetaSchemaName  = "user"
	groupMetaSchemaName = "group"
	orgMetaSchemaName   = "organization"
	rolesMetaSchemaName = "role"
)

//go:embed metaschemas/defaultUserSchema.json
var defaultUser []byte

//go:embed metaschemas/defaultGroupSchema.json
var defaultGroup []byte

//go:embed metaschemas/defaultOrgSchema.json
var defaultOrg []byte

//go:embed metaschemas/defaultRoleSchema.json
var defaultRole []byte

var metaSchemaCache = make(map[string]string)

type MetaSchemaRepository struct {
	dbc *db.Client
}

func NewMetaSchemaRepository(dbc *db.Client) *MetaSchemaRepository {
	return &MetaSchemaRepository{
		dbc: dbc,
	}
}

func (m MetaSchemaRepository) Get(ctx context.Context, name string) (metaschema.MetaSchema, error) {
	if strings.TrimSpace(name) == "" {
		return metaschema.MetaSchema{}, metaschema.ErrInvalidName
	}

	query, params, err := dialect.From(TABLE_METASCHEMA).Where(
		goqu.Ex{
			"name": name,
		},
	).ToSQL()
	if err != nil {
		return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedMetaSchema MetaSchema
	if err = m.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METASCHEMA,
				Operation:  "Get",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return m.dbc.QueryRowxContext(ctx, query, params...).StructScan(&fetchedMetaSchema)
	}); err != nil {
		err = checkPostgresError(err)
		return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return fetchedMetaSchema.tranformtoMetadataSchema(), nil
}

func (m MetaSchemaRepository) Create(ctx context.Context, mschema metaschema.MetaSchema) (string, error) {
	if strings.TrimSpace(mschema.Name) == "" {
		return "", metaschema.ErrInvalidName
	}

	if strings.TrimSpace(mschema.Schema) == "" {
		return "", metaschema.ErrInvalidDetail
	}

	createQuery, params, err := dialect.Insert(TABLE_METASCHEMA).Rows(
		goqu.Record{
			"name":   mschema.Name,
			"schema": mschema.Schema,
		}).Returning("name").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var schemaName string
	if err = m.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METASCHEMA,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return m.dbc.QueryRowxContext(ctx, createQuery, params...).Scan(&schemaName)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return "", metaschema.ErrConflict
		default:
			return "", err
		}
	}

	metaSchemaCache[mschema.Name] = mschema.Schema
	return schemaName, nil
}

func (m MetaSchemaRepository) List(ctx context.Context) ([]metaschema.MetaSchema, error) {
	query, params, err := dialect.From(TABLE_METASCHEMA).ToSQL()
	if err != nil {
		return []metaschema.MetaSchema{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var schemaModels []MetaSchema
	if err = m.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METASCHEMA,
				Operation:  "List",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return m.dbc.SelectContext(ctx, &schemaModels, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []metaschema.MetaSchema{}, metaschema.ErrNotExist
		}
		return []metaschema.MetaSchema{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedSchemas []metaschema.MetaSchema
	for _, m := range schemaModels {
		transformedSchemas = append(transformedSchemas, m.tranformtoMetadataSchema())
	}

	return transformedSchemas, nil
}

func (m MetaSchemaRepository) Delete(ctx context.Context, name string) error {
	query, params, err := dialect.Delete(TABLE_METASCHEMA).
		Where(
			goqu.Ex{
				"name": name,
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return m.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METASCHEMA,
				Operation:  "Delete",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		result, err := m.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return fmt.Errorf("%w: %s", dbErr, metaschema.ErrNotExist)
			default:
				return fmt.Errorf("%w: %s", dbErr, err)
			}
		}

		if count, _ := result.RowsAffected(); count > 0 {
			delete(metaSchemaCache, name)
			return nil
		}

		return metaschema.ErrNotExist
	})
}

func (m MetaSchemaRepository) Update(ctx context.Context, name string, mschema metaschema.MetaSchema) (string, error) {
	query, params, err := dialect.Update(TABLE_METASCHEMA).Set(
		goqu.Record{
			"schema": mschema.Schema,
		}).Where(goqu.Ex{
		"name": name,
	}).Returning("name").ToSQL()

	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var schemaName string
	if err = m.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METASCHEMA,
				Operation:  "Update",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return m.dbc.QueryRowxContext(ctx, query, params...).Scan(&schemaName)
	}); err != nil {
		err = checkPostgresError(err)
		return "", fmt.Errorf("%w: %s", dbErr, err)
	}

	metaSchemaCache[mschema.Name] = mschema.Schema
	return schemaName, nil
}

func (m MetaSchemaRepository) InitMetaSchemas(ctx context.Context) error {
	// load schemas in memory from the db
	schemas, err := m.List(ctx)
	if err != nil {
		return errors.New("error in initialising metadata json-schemas")
	}

	for _, s := range schemas {
		metaSchemaCache[s.Name] = s.Schema
	}

	// if default metaschemas doesnt exists add to the db
	if metaSchemaCache[userMetaSchemaName] == "" {
		m.Create(ctx, metaschema.MetaSchema{
			Name:   userMetaSchemaName,
			Schema: string(defaultUser),
		})
	}

	if metaSchemaCache[groupMetaSchemaName] == "" {
		m.Create(ctx, metaschema.MetaSchema{
			Name:   groupMetaSchemaName,
			Schema: string(defaultGroup),
		})
	}

	if metaSchemaCache[orgMetaSchemaName] == "" {
		m.Create(ctx, metaschema.MetaSchema{
			Name:   orgMetaSchemaName,
			Schema: string(defaultOrg),
		})
	}

	if metaSchemaCache[rolesMetaSchemaName] == "" {
		m.Create(ctx, metaschema.MetaSchema{
			Name:   rolesMetaSchemaName,
			Schema: string(defaultRole),
		})
	}

	return nil
}

func validateMetadataSchema(marshaledMetadata []byte, schemaName string) error {
	if metaSchemaCache[schemaName] == "" {
		return fmt.Errorf("%s metaschema doesn't exists", schemaName)
	}
	metadataSchema := gojsonschema.NewStringLoader(metaSchemaCache[schemaName])
	providedSchema := gojsonschema.NewBytesLoader(marshaledMetadata)
	results, err := gojsonschema.Validate(metadataSchema, providedSchema)
	if err != nil {
		return errors.New("failed to validate metadata")
	}

	if !results.Valid() {
		return errors.New("metadata doesn't match the json-schema")
	}
	return nil
}
