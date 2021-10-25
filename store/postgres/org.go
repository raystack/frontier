package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/shield/internal/org"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
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
	selectOrganizationForUpdateQuery = `SELECT id, name, slug, metadata, created_at, updated_at from organizations where id=$1 FOR UPDATE;`
	updateOrganizationQuery          = `UPDATE organizations set name = $2, slug = $3, metadata = $4, updated_at = now() where id = $1 RETURNING id, name, slug, metadata, created_at, updated_at;`
)

func (s Store) GetOrg(ctx context.Context, id string) (org.Organization, error) {
	return s.selectOrg(ctx, id, false, nil)
}

func (s Store) selectOrg(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (org.Organization, error) {
	var fetchedOrg Organization

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		if forUpdate {
			return txn.GetContext(ctx, &fetchedOrg, selectOrganizationForUpdateQuery, id)
		} else {
			return s.DB.GetContext(ctx, &fetchedOrg, getOrganizationsQuery, id)
		}
	})

	if errors.Is(err, sql.ErrNoRows) {
		return org.Organization{}, org.OrgDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return org.Organization{}, org.InvalidUUID
	} else if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToOrg(fetchedOrg)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (s Store) CreateOrg(ctx context.Context, orgToCreate org.Organization) (org.Organization, error) {
	var newOrg Organization

	marshaledMetadata, err := json.Marshal(orgToCreate.Metadata)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newOrg, createOrganizationQuery, orgToCreate.Name, orgToCreate.Slug, marshaledMetadata)
	})
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToOrg(newOrg)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func (s Store) ListOrg(ctx context.Context) ([]org.Organization, error) {
	var fetchedOrgs []Organization
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedOrgs, listOrganizationsQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []org.Organization{}, org.OrgDoesntExist
	}

	if err != nil {
		return []org.Organization{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedOrgs []org.Organization

	for _, o := range fetchedOrgs {
		transformedOrg, err := transformToOrg(o)
		if err != nil {
			return []org.Organization{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedOrgs = append(transformedOrgs, transformedOrg)
	}

	return transformedOrgs, nil
}

func (s Store) UpdateOrg(ctx context.Context, toUpdate org.Organization) (org.Organization, error) {
	var updateSet org.Organization
	var updatedOrg Organization
	var isModified bool

	err := s.DB.WithTxn(ctx, sql.TxOptions{Isolation: sql.LevelReadCommitted}, func(tx *sqlx.Tx) error {
		fetchedOrg, err := s.selectOrg(ctx, toUpdate.Id, true, tx)
		if err != nil {
			return err
		}

		updateSet, isModified = getUpdatedOrganizationObj(fetchedOrg, toUpdate)
		if !isModified {
			return nil
		}

		marshaledMetadata, err := json.Marshal(updateSet.Metadata)
		if err != nil {
			return fmt.Errorf("%w: %s", parseErr, err)
		}

		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return tx.GetContext(ctx, &updatedOrg, updateOrganizationQuery, updateSet.Id, updateSet.Name, updateSet.Slug, marshaledMetadata)
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return org.Organization{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	updateSet, err = transformToOrg(updatedOrg)
	if err != nil {
		return org.Organization{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return updateSet, nil
}

func transformToOrg(from Organization) (org.Organization, error) {
	var unmarshalledMetadata map[string]interface{}
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return org.Organization{}, err
	}

	return org.Organization{
		Id:        from.Id,
		Name:      from.Name,
		Slug:      from.Slug,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}

func getUpdatedOrganizationObj(existing, req org.Organization) (org.Organization, bool) {
	var isModified bool

	if req.Name != "" && req.Name != existing.Name {
		existing.Name = req.Name
		isModified = true
	}

	if req.Slug != "" && req.Slug != existing.Slug {
		existing.Slug = req.Slug
		isModified = true
	}

	// TODO: Check if "This can also clear the metadata" requirement
	if !cmp.Equal(req.Metadata, existing.Metadata) {
		existing.Metadata = req.Metadata
		isModified = true
	}

	return existing, isModified
}
