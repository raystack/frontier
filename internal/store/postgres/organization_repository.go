package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/shield/core/user"

	"github.com/odpf/shield/core/metaschema"
	"github.com/odpf/shield/core/project"

	"github.com/doug-martin/goqu/v9"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/pkg/db"
)

type OrganizationRepository struct {
	dbc *db.Client
}

func NewOrganizationRepository(dbc *db.Client) *OrganizationRepository {
	return &OrganizationRepository{
		dbc: dbc,
	}
}

var notDisabledOrgExp = goqu.Or(
	goqu.Ex{
		"state": nil,
	},
	goqu.Ex{
		"state": goqu.Op{"neq": organization.Disabled},
	},
)

func (r OrganizationRepository) GetByID(ctx context.Context, id string) (organization.Organization, error) {
	if strings.TrimSpace(id) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"id": id,
	}).Where(notDisabledOrgExp).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "GetByID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.GetContext(ctx, &orgModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return organization.Organization{}, organization.ErrInvalidUUID
		default:
			return organization.Organization{}, err
		}
	}

	transformedOrg, err := orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) GetByIDs(ctx context.Context, ids []string) ([]organization.Organization, error) {
	if len(ids) == 0 {
		return nil, project.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"id": goqu.Op{"in": ids},
	}).Where(notDisabledOrgExp).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgs []Organization
	// TODO(kushsharma): clean up this unnecessary newrelic blot over each query
	// abstract it over database using a facade
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "GetByIDs",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &orgs, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, organization.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return nil, organization.ErrInvalidUUID
		default:
			return nil, err
		}
	}

	var transformedOrgs []organization.Organization
	for _, o := range orgs {
		to, err := o.transformToOrg()
		if err != nil {
			return nil, err
		}
		transformedOrgs = append(transformedOrgs, to)
	}
	return transformedOrgs, nil
}

func (r OrganizationRepository) GetBySlug(ctx context.Context, slug string) (organization.Organization, error) {
	if strings.TrimSpace(slug) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"slug": slug,
	}).Where(notDisabledOrgExp).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "GetBySlug",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}
		return r.dbc.GetContext(ctx, &orgModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return organization.Organization{}, organization.ErrInvalidUUID
		default:
			return organization.Organization{}, err
		}
	}

	transformedOrg, err := orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) Create(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	if strings.TrimSpace(org.Name) == "" || strings.TrimSpace(org.Slug) == "" {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	if err = validateMetadataSchema(marshaledMetadata, orgMetaSchemaName); err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", metaschema.ErrInvalidMetaSchema, err)
	}

	insertRow := goqu.Record{
		"name":     org.Name,
		"slug":     org.Slug,
		"metadata": marshaledMetadata,
	}
	if org.State != "" {
		insertRow["state"] = org.State
	}
	query, params, err := dialect.Insert(TABLE_ORGANIZATIONS).Rows(insertRow).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&orgModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return organization.Organization{}, organization.ErrConflict
		default:
			return organization.Organization{}, err
		}
	}

	transformedOrg, err := orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) List(ctx context.Context, flt organization.Filter) ([]organization.Organization, error) {
	stmt := dialect.From(TABLE_ORGANIZATIONS)
	if flt.State == "" {
		stmt = stmt.Where(notDisabledOrgExp)
	} else {
		stmt = stmt.Where(goqu.Ex{
			"state": flt.State.String(),
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return []organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModels []Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "List",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}
		return r.dbc.SelectContext(ctx, &orgModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []organization.Organization{}, nil
		}
		return []organization.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedOrgs []organization.Organization
	for _, o := range orgModels {
		transformedOrg, err := o.transformToOrg()
		if err != nil {
			return []organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedOrgs = append(transformedOrgs, transformedOrg)
	}

	return transformedOrgs, nil
}

func (r OrganizationRepository) UpdateByID(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	if strings.TrimSpace(org.ID) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	if strings.TrimSpace(org.Name) == "" || strings.TrimSpace(org.Slug) == "" {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	if err = validateMetadataSchema(marshaledMetadata, orgMetaSchemaName); err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", metaschema.ErrInvalidMetaSchema, err)
	}

	query, params, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"name":       org.Name,
			"slug":       org.Slug,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": org.ID,
	}).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "UpdateByID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&orgModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, errDuplicateKey):
			return organization.Organization{}, organization.ErrConflict
		case errors.Is(err, errInvalidTexRepresentation):
			return organization.Organization{}, organization.ErrInvalidUUID
		default:
			return organization.Organization{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	org, err = orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return org, nil
}

func (r OrganizationRepository) UpdateBySlug(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	if strings.TrimSpace(org.Slug) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	if strings.TrimSpace(org.Name) == "" {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	if err = validateMetadataSchema(marshaledMetadata, orgMetaSchemaName); err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", metaschema.ErrInvalidMetaSchema, err)
	}

	query, params, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"name":       org.Name,
			"slug":       org.Slug,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(
		goqu.Ex{
			"slug": org.Slug,
		}).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "UpdateBySlug",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&orgModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, errDuplicateKey):
			return organization.Organization{}, organization.ErrConflict
		default:
			return organization.Organization{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	org, err = orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return org, nil
}

func (r OrganizationRepository) SetState(ctx context.Context, id string, state organization.State) error {
	query, params, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"state": state.String(),
		}).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "SetState",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user.ErrNotExist
		default:
			return err
		}
	}
	return nil
}

func (r OrganizationRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_ORGANIZATIONS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_ORGANIZATIONS,
				Operation:  "Delete",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
