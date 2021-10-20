package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/odpf/shield/org"
	"time"
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
	getOrganizationsQuery   = `SELECT id, name, slug, metadata, created_at, updated_at from organizations where id=$1;`
	createOrganizationQuery = `INSERT INTO organizations(name, slug, metadata) values($1, $2, $3) RETURNING id, name, slug, metadata, created_at, updated_at;`
	listOrganizationsQuery  = `SELECT id, name, slug, metadata, created_at, updated_at from organizations;`

)

var (
	parseErr = errors.New("parsing error")
	dbErr    = errors.New("error while running query")
)

func (s Store) GetOrg(ctx context.Context, id string) (org.Organization, error) {
	var fetchedOrg Organization
	err := s.DB.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id)

	if errors.Is(err, sql.ErrNoRows) {
		return org.Organization{}, org.OrgDoesntExist
	} else
	if fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return org.Organization{}, org.InvalidUUID
	}

	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var unmarshalledMetadata map[string]interface{}
	if err := json.Unmarshal(fetchedOrg.Metadata, &unmarshalledMetadata); err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return org.Organization{
		Id:        fetchedOrg.Id,
		Name:      fetchedOrg.Name,
		Slug:      fetchedOrg.Slug,
		Metadata:  unmarshalledMetadata,
		CreatedAt: fetchedOrg.CreatedAt,
		UpdatedAt: fetchedOrg.UpdatedAt,
	}, nil
}

func (s Store) CreateOrg(ctx context.Context, orgToCreate org.Organization) (org.Organization, error) {
	var newOrg Organization

	marshaledMetadata, err := json.Marshal(orgToCreate.Metadata)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	err = s.DB.GetContext(ctx, &newOrg, createOrganizationQuery, orgToCreate.Name, orgToCreate.Slug, marshaledMetadata)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var unmarshalledMetadata map[string]interface{}
	err = json.Unmarshal(newOrg.Metadata, &unmarshalledMetadata)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return org.Organization{
		Id:        newOrg.Id,
		Name:      newOrg.Name,
		Slug:      newOrg.Slug,
		Metadata:  unmarshalledMetadata,
		CreatedAt: newOrg.CreatedAt,
		UpdatedAt: newOrg.UpdatedAt,
	}, nil
}

func (s Store) ListOrg(context.Context) ([]org.Organization, error) {
	return []org.Organization{}, nil
}
