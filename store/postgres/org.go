package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/org"
	"github.com/odpf/shield/model"
)

type Organization struct {
	Id        string       `db:"id"`
	Name      string       `db:"name"`
	Slug      string       `db:"slug"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// *Get Organizations Query
func buildGetOrganizationsBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	getOrganizationsBySlugQuery, _, err := dialect.From(TABLE_ORG).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).ToSQL()

	return getOrganizationsBySlugQuery, err
}

func buildGetOrganizationsByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	getOrganizationsByIdQuery, _, err := dialect.From(TABLE_ORG).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).ToSQL()

	return getOrganizationsByIdQuery, err
}

// *Create Organization Query
func buildCreateOrganizationQuery(dialect goqu.DialectWrapper) (string, error) {
	createOrganizationQuery, _, err := dialect.Insert(TABLE_ORG).Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"metadata": goqu.L("$3"),
		}).Returning(&Organization{}).ToSQL()

	return createOrganizationQuery, err
}

// *List Organization Query
func buildListOrganizationsQuery(dialect goqu.DialectWrapper) (string, error) {
	listOrganizationsQuery, _, err := dialect.From(TABLE_ORG).ToSQL()

	return listOrganizationsQuery, err
}

func buildListOrganizationAdmins(dialect goqu.DialectWrapper) (string, error) {
	listOrganizationAdmins, _, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.T(TABLE_RELATION).As("r")).
		Join(goqu.T(TABLE_USER).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              definition.OrganizationAdminRole.Id,
		"r.subject_namespace_id": definition.UserNamespace.Id,
		"r.object_namespace_id":  definition.OrgNamespace.Id,
	}).ToSQL()

	return listOrganizationAdmins, err
}

// *Update Organization Query
func buildUpdateOrganizationBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	updateOrganizationQuery, _, err := dialect.Update(TABLE_ORG).Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"metadata":   goqu.L("$4"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).Returning(&Organization{}).ToSQL()

	return updateOrganizationQuery, err
}

func buildUpdateOrganizationByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	updateOrganizationQuery, _, err := dialect.Update(TABLE_ORG).Set(
		goqu.Record{
			"name":       goqu.L("$3"),
			"slug":       goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"slug": goqu.L("$1"),
		"id":   goqu.L("$2"),
	}).Returning(&Organization{}).ToSQL()

	return updateOrganizationQuery, err
}

// GetOrg Supports Slug
func (s Store) GetOrg(ctx context.Context, id string) (model.Organization, error) {
	var fetchedOrg Organization
	var getOrganizationsQuery string
	var err error
	id = strings.TrimSpace(id)
	isUuid := isUUID(id)

	if isUuid {
		getOrganizationsQuery, err = buildGetOrganizationsByIdQuery(dialect)
	} else {
		getOrganizationsQuery, err = buildGetOrganizationsBySlugQuery(dialect)
	}
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id, id)
		})
	} else {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id)
		})
	}

	if errors.Is(err, sql.ErrNoRows) {
		return model.Organization{}, org.OrgDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Organization{}, org.InvalidUUID
	} else if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToOrg(fetchedOrg)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (s Store) CreateOrg(ctx context.Context, orgToCreate model.Organization) (model.Organization, error) {
	marshaledMetadata, err := json.Marshal(orgToCreate.Metadata)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	createOrganizationQuery, err := buildCreateOrganizationQuery(dialect)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var newOrg Organization
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newOrg, createOrganizationQuery, orgToCreate.Name, orgToCreate.Slug, marshaledMetadata)
	})

	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToOrg(newOrg)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (s Store) ListOrg(ctx context.Context) ([]model.Organization, error) {
	var fetchedOrgs []Organization
	listOrganizationsQuery, err := buildListOrganizationsQuery(dialect)
	if err != nil {
		return []model.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedOrgs, listOrganizationsQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Organization{}, org.OrgDoesntExist
	}

	if err != nil {
		return []model.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedOrgs []model.Organization

	for _, o := range fetchedOrgs {
		transformedOrg, err := transformToOrg(o)
		if err != nil {
			return []model.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedOrgs = append(transformedOrgs, transformedOrg)
	}

	return transformedOrgs, nil
}

// UpdateOrg Supports Slug
func (s Store) UpdateOrg(ctx context.Context, toUpdate model.Organization) (model.Organization, error) {
	var updatedOrg Organization

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateOrganizationQuery string
	isUuid := isUUID(toUpdate.Id)

	if isUuid {
		updateOrganizationQuery, err = buildUpdateOrganizationByIdQuery(dialect)
	} else {
		updateOrganizationQuery, err = buildUpdateOrganizationBySlugQuery(dialect)
	}
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &updatedOrg, updateOrganizationQuery, toUpdate.Id, toUpdate.Id, toUpdate.Name, toUpdate.Slug, marshaledMetadata)
		})
	} else {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &updatedOrg, updateOrganizationQuery, toUpdate.Id, toUpdate.Name, toUpdate.Slug, marshaledMetadata)
		})
	}
	if err != nil {
		return model.Organization{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	toUpdate, err = transformToOrg(updatedOrg)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

// ListOrgAdmins Supports Slug
func (s Store) ListOrgAdmins(ctx context.Context, id string) ([]model.User, error) {
	var fetchedUsers []User
	listOrganizationAdmins, err := buildListOrganizationAdmins(dialect)
	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	id = strings.TrimSpace(id)
	isUuid := isUUID(id)
	if !isUuid {
		fetchedOrg, err := s.GetOrg(ctx, id)
		if err != nil {
			return []model.User{}, err
		}
		id = fetchedOrg.Id
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listOrganizationAdmins, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.User{}, org.NoAdminsExist
	}

	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User
	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []model.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func transformToOrg(from Organization) (model.Organization, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return model.Organization{}, err
	}

	return model.Organization{
		Id:        from.Id,
		Name:      from.Name,
		Slug:      from.Slug,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
