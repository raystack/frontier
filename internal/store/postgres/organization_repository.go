package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
)

type OrganizationRepository struct {
	dbc *db.Client
}

func NewOrganizationRepository(dbc *db.Client) *OrganizationRepository {
	return &OrganizationRepository{
		dbc: dbc,
	}
}

func (r OrganizationRepository) GetByID(ctx context.Context, id string) (organization.Organization, error) {
	if id == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
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

func (r OrganizationRepository) GetBySlug(ctx context.Context, slug string) (organization.Organization, error) {
	if slug == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"slug": slug,
	}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
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
	if str.IsStringEmpty(org.Name) || str.IsStringEmpty(org.Slug) {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_ORGANIZATIONS).Rows(
		goqu.Record{
			"name":     org.Name,
			"slug":     org.Slug,
			"metadata": marshaledMetadata,
		}).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
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

func (r OrganizationRepository) List(ctx context.Context) ([]organization.Organization, error) {
	query, params, err := dialect.From(TABLE_ORGANIZATIONS).ToSQL()
	if err != nil {
		return []organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var orgModels []Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
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
	if str.IsStringEmpty(org.ID) {
		return organization.Organization{}, organization.ErrInvalidID
	}

	if str.IsStringEmpty(org.Name) || str.IsStringEmpty(org.Slug) {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
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
	if str.IsStringEmpty(org.Slug) {
		return organization.Organization{}, organization.ErrInvalidID
	}

	if str.IsStringEmpty(org.Name) {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
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

func (r OrganizationRepository) ListAdminsByOrgID(ctx context.Context, orgID string) ([]user.User, error) {
	if orgID == "" {
		return []user.User{}, organization.ErrInvalidID
	}

	query, params, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).
		From(goqu.T(TABLE_RELATIONS).As("r")).
		Join(goqu.T(TABLE_USERS).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            orgID,
		"r.role_id":              role.DefinitionOrganizationAdmin.ID,
		"r.subject_namespace_id": namespace.DefinitionUser.ID,
		"r.object_namespace_id":  namespace.DefinitionOrg.ID,
	}).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModels []User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &userModels, query, params...)
	}); err != nil {
		// List should not return error if empty
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, nil
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User
	for _, u := range userModels {
		transformedUser, err := u.transformToUser()
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}
