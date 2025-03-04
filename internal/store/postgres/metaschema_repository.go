package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	"github.com/raystack/salt/log"

	"github.com/pkg/errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/metaschema"
	"github.com/raystack/frontier/pkg/db"
)

var (
	userMetaSchemaName     = "user"
	groupMetaSchemaName    = "group"
	orgMetaSchemaName      = "organization"
	rolesMetaSchemaName    = "role"
	prospectMetaSchemaName = "prospect"
)

//go:embed metaschemas/user.json
var defaultUser []byte

//go:embed metaschemas/group.json
var defaultGroup []byte

//go:embed metaschemas/org.json
var defaultOrg []byte

//go:embed metaschemas/role.json
var defaultRole []byte

//go:embed metaschemas/prospect.json
var defaultProspect []byte

var defaultMetaSchemas = map[string]string{
	userMetaSchemaName:     string(defaultUser),
	groupMetaSchemaName:    string(defaultGroup),
	orgMetaSchemaName:      string(defaultOrg),
	rolesMetaSchemaName:    string(defaultRole),
	prospectMetaSchemaName: string(defaultProspect),
}

type MetaSchemaRepository struct {
	log log.Logger
	dbc *db.Client
}

func NewMetaSchemaRepository(logger log.Logger, dbc *db.Client) *MetaSchemaRepository {
	return &MetaSchemaRepository{
		log: logger,
		dbc: dbc,
	}
}

func (m MetaSchemaRepository) Get(ctx context.Context, id string) (metaschema.MetaSchema, error) {
	if strings.TrimSpace(id) == "" {
		return metaschema.MetaSchema{}, metaschema.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_METASCHEMA).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedMetaSchema MetaSchema
	if err = m.dbc.WithTimeout(ctx, TABLE_METASCHEMA, "Get", func(ctx context.Context) error {
		return m.dbc.QueryRowxContext(ctx, query, params...).StructScan(&fetchedMetaSchema)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return metaschema.MetaSchema{}, metaschema.ErrNotExist
		default:
			return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return fetchedMetaSchema.tranformtoMetadataSchema(), nil
}

func (m MetaSchemaRepository) Create(ctx context.Context, mschema metaschema.MetaSchema) (metaschema.MetaSchema, error) {
	if strings.TrimSpace(mschema.Name) == "" {
		return metaschema.MetaSchema{}, metaschema.ErrInvalidID
	}

	if strings.TrimSpace(mschema.Schema) == "" {
		return metaschema.MetaSchema{}, metaschema.ErrInvalidDetail
	}

	createQuery, params, err := dialect.Insert(TABLE_METASCHEMA).Rows(
		goqu.Record{
			"name":   mschema.Name,
			"schema": mschema.Schema,
		}).OnConflict(goqu.DoNothing()).Returning(&MetaSchema{}).ToSQL()
	if err != nil {
		return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var schemaModel MetaSchema
	if err = m.dbc.WithTimeout(ctx, TABLE_METASCHEMA, "Create", func(ctx context.Context) error {
		return m.dbc.QueryRowxContext(ctx, createQuery, params...).StructScan(&schemaModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.As(err, &ErrDuplicateKey):
			return metaschema.MetaSchema{}, metaschema.ErrConflict
		default:
			return metaschema.MetaSchema{}, err
		}
	}

	return schemaModel.tranformtoMetadataSchema(), nil
}

func (m MetaSchemaRepository) List(ctx context.Context) ([]metaschema.MetaSchema, error) {
	query, params, err := dialect.From(TABLE_METASCHEMA).ToSQL()
	if err != nil {
		return []metaschema.MetaSchema{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var schemaModels []MetaSchema
	if err = m.dbc.WithTimeout(ctx, TABLE_METASCHEMA, "List", func(ctx context.Context) error {
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

func (m MetaSchemaRepository) Delete(ctx context.Context, id string) (string, error) {
	query, params, err := dialect.Delete(TABLE_METASCHEMA).
		Where(
			goqu.Ex{
				"id": id,
			},
		).Returning("name").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var name string
	if err = m.dbc.WithTimeout(ctx, TABLE_METASCHEMA, "Delete", func(ctx context.Context) error {
		return m.dbc.QueryRowxContext(ctx, query, params...).Scan(&name)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", metaschema.ErrNotExist
		default:
			return "", fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return name, nil
}

func (m MetaSchemaRepository) Update(ctx context.Context, id string, mschema metaschema.MetaSchema) (metaschema.MetaSchema, error) {
	query, params, err := dialect.Update(TABLE_METASCHEMA).Set(
		goqu.Record{
			"schema":     mschema.Schema,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": id,
	}).Returning(&MetaSchema{}).ToSQL()

	if err != nil {
		return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var schemaModel MetaSchema
	if err = m.dbc.WithTimeout(ctx, TABLE_METASCHEMA, "Update", func(ctx context.Context) error {
		return m.dbc.QueryRowxContext(ctx, query, params...).StructScan(&schemaModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return metaschema.MetaSchema{}, metaschema.ErrNotExist
		default:
			return metaschema.MetaSchema{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return schemaModel.tranformtoMetadataSchema(), nil
}

// add default schemas to db once during database migration
func (m MetaSchemaRepository) MigrateDefaults(ctx context.Context) error {
	for name, schema := range defaultMetaSchemas {
		if _, err := m.Create(ctx, metaschema.MetaSchema{
			Name:   name,
			Schema: schema,
		}); err != nil {
			err = checkPostgresError(err)
			if errors.Is(metaschema.ErrConflict, err) {
				m.log.Debug("schema already exists", "name", name)
				continue
			}
			return errors.Wrap(err, "error in adding default schemas to db")
		}
	}
	return nil
}
