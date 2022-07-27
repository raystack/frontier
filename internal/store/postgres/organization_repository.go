package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/user"
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

func (r OrganizationRepository) Get(ctx context.Context, id string) (organization.Organization, error) {
	var fetchedOrg Organization
	var getOrganizationsQuery string
	var err error
	id = strings.TrimSpace(id)
	isUuid := isUUID(id)

	//TODO decouple these to make these cleaner
	if isUuid {
		getOrganizationsQuery, err = buildGetOrganizationsByIDQuery(dialect)
	} else {
		getOrganizationsQuery, err = buildGetOrganizationsBySlugQuery(dialect)
	}
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id, id)
		})
	} else {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id)
		})
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return organization.Organization{}, organization.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return organization.Organization{}, organization.ErrInvalidUUID
		}
		return organization.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToOrg(fetchedOrg)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) Create(ctx context.Context, orgToCreate organization.Organization) (organization.Organization, error) {
	marshaledMetadata, err := json.Marshal(orgToCreate.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	createOrganizationQuery, err := buildCreateOrganizationQuery(dialect)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var newOrg Organization
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &newOrg, createOrganizationQuery, orgToCreate.Name, orgToCreate.Slug, marshaledMetadata)
	}); err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToOrg(newOrg)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) List(ctx context.Context) ([]organization.Organization, error) {
	var fetchedOrgs []Organization
	listOrganizationsQuery, err := buildListOrganizationsQuery(dialect)
	if err != nil {
		return []organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedOrgs, listOrganizationsQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []organization.Organization{}, organization.ErrNotExist
		}
		return []organization.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedOrgs []organization.Organization
	for _, o := range fetchedOrgs {
		transformedOrg, err := transformToOrg(o)
		if err != nil {
			return []organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedOrgs = append(transformedOrgs, transformedOrg)
	}

	return transformedOrgs, nil
}

func (r OrganizationRepository) Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error) {
	var updatedOrg Organization

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateOrganizationQuery string
	isUuid := isUUID(toUpdate.ID)

	if isUuid {
		updateOrganizationQuery, err = buildUpdateOrganizationByIDQuery(dialect)
	} else {
		updateOrganizationQuery, err = buildUpdateOrganizationBySlugQuery(dialect)
	}
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &updatedOrg, updateOrganizationQuery, toUpdate.ID, toUpdate.ID, toUpdate.Name, toUpdate.Slug, marshaledMetadata)
		})
	} else {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &updatedOrg, updateOrganizationQuery, toUpdate.ID, toUpdate.Name, toUpdate.Slug, marshaledMetadata)
		})
	}
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	toUpdate, err = transformToOrg(updatedOrg)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func (r OrganizationRepository) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	var fetchedUsers []User
	listOrganizationAdmins, err := buildListOrganizationAdmins(dialect)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	id = strings.TrimSpace(id)
	fetchedOrg, err := r.Get(ctx, id)
	if err != nil {
		return []user.User{}, err
	}
	id = fetchedOrg.ID

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, listOrganizationAdmins, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, organization.ErrNoAdminsExist
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User
	for _, u := range fetchedUsers {
		transformedUser, err := u.transformToUser()
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}
