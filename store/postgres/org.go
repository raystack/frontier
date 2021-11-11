package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/shield/internal/org"
	"github.com/odpf/shield/model"
)

type Organization struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	Metadata  []byte    `db:"metadata"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	getOrganizationsQuery            = `SELECT id, name, slug, metadata, created_at, updated_at from organizations where id=$1;`
	createOrganizationQuery          = `INSERT INTO organizations(name, slug, metadata) values($1, $2, $3) RETURNING id, name, slug, metadata, created_at, updated_at;`
	listOrganizationsQuery           = `SELECT id, name, slug, metadata, created_at, updated_at from organizations;`
	updateOrganizationQuery          = `UPDATE organizations set name = $2, slug = $3, metadata = $4, updated_at = now() where id = $1 RETURNING id, name, slug, metadata, created_at, updated_at;`
)

func (s Store) GetOrg(ctx context.Context, id string) (model.Organization, error) {
	var fetchedOrg Organization
	//fetchedOrg, _, err := s.selectOrg(ctx, id, false, nil)
	//return fetchedOrg, err
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id)
	})

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
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
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

func (s Store) UpdateOrg(ctx context.Context, toUpdate model.Organization) (model.Organization, error) {
	var updatedOrg Organization

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedOrg, updateOrganizationQuery, toUpdate.Id, toUpdate.Name, toUpdate.Slug, marshaledMetadata)
	})

	if err != nil {
		return model.Organization{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	toUpdate, err = transformToOrg(updatedOrg)
	if err != nil {
		return model.Organization{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func transformToOrg(from Organization) (model.Organization, error) {
	var unmarshalledMetadata map[string]string
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
