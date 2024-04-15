package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/pkg/db"
)

type ServiceUserCredentialRepository struct {
	dbc *db.Client
}

func NewServiceUserCredentialRepository(dbc *db.Client) *ServiceUserCredentialRepository {
	return &ServiceUserCredentialRepository{
		dbc: dbc,
	}
}

func (s ServiceUserCredentialRepository) List(ctx context.Context, flt serviceuser.Filter) ([]serviceuser.Credential, error) {
	stmt := dialect.Select(
		goqu.I("s.id"),
		goqu.I("s.serviceuser_id"),
		goqu.I("s.secret_hash"),
		goqu.I("s.type"),
		goqu.I("s.public_key"),
		goqu.I("s.title"),
		goqu.I("s.metadata"),
		goqu.I("s.created_at"),
		goqu.I("s.updated_at"),
	).Where(goqu.Ex{
		"serviceuser_id": flt.ServiceUserID,
	})

	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}
	if flt.IsKey {
		stmt = stmt.Where(goqu.Ex{
			"public_key": goqu.Op{"isNot": nil},
		})
	}
	if flt.IsSecret {
		stmt = stmt.Where(goqu.Ex{
			"secret_hash": goqu.Op{"isNot": nil},
		})
	}

	query, params, err := stmt.From(goqu.T(TABLE_SERVICEUSERCREDENTIALS).As("s")).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var svUserCreds []ServiceUserCredential
	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSERCREDENTIALS, "List", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &svUserCreds, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedServiceUsers []serviceuser.Credential
	for _, o := range svUserCreds {
		transformedServiceUser, err := o.transform()
		if err != nil {
			return nil, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedServiceUsers = append(transformedServiceUsers, transformedServiceUser)
	}

	return transformedServiceUsers, nil
}

func (s ServiceUserCredentialRepository) Create(ctx context.Context, credential serviceuser.Credential) (serviceuser.Credential, error) {
	if strings.TrimSpace(credential.ID) == "" {
		credential.ID = uuid.New().String()
	}

	marshaledMetadata, err := json.Marshal(credential.Metadata)
	if err != nil {
		return serviceuser.Credential{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	publicKeyJson, err := json.Marshal(credential.PublicKey)
	if err != nil {
		return serviceuser.Credential{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	svUserCred := ServiceUserCredential{}
	query, params, err := dialect.Insert(TABLE_SERVICEUSERCREDENTIALS).Rows(
		goqu.Record{
			"id":             credential.ID,
			"serviceuser_id": credential.ServiceUserID,
			"type":           credential.Type.String(),
			"secret_hash":    credential.SecretHash,
			"public_key":     publicKeyJson,
			"title":          credential.Title,
			"metadata":       marshaledMetadata,
		}).OnConflict(
		goqu.DoUpdate("id", goqu.Record{
			"title":    credential.Title,
			"metadata": marshaledMetadata,
		})).Returning(&ServiceUserCredential{}).ToSQL()
	if err != nil {
		return serviceuser.Credential{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSERCREDENTIALS, "Create", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&svUserCred)
	}); err != nil {
		return serviceuser.Credential{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return svUserCred.transform()
}

func (s ServiceUserCredentialRepository) Get(ctx context.Context, id string) (serviceuser.Credential, error) {
	if strings.TrimSpace(id) == "" {
		return serviceuser.Credential{}, serviceuser.ErrInvalidKeyID
	}

	query, params, err := dialect.Select(
		goqu.I("s.id"),
		goqu.I("s.serviceuser_id"),
		goqu.I("s.type"),
		goqu.I("s.secret_hash"),
		goqu.I("s.public_key"),
		goqu.I("s.title"),
		goqu.I("s.metadata"),
		goqu.I("s.created_at"),
		goqu.I("s.updated_at"),
	).Where(
		goqu.Ex{"s.id": id},
	).From(goqu.T(TABLE_SERVICEUSERCREDENTIALS).As("s")).ToSQL()
	if err != nil {
		return serviceuser.Credential{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var svUserCredentialModel ServiceUserCredential
	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSERCREDENTIALS, "Get", func(ctx context.Context) error {
		return s.dbc.GetContext(ctx, &svUserCredentialModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return serviceuser.Credential{}, serviceuser.ErrCredNotExist
		}
		return serviceuser.Credential{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return svUserCredentialModel.transform()
}

func (s ServiceUserCredentialRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_SERVICEUSERCREDENTIALS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSERCREDENTIALS, "Delete", func(ctx context.Context) error {
		if _, err = s.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return serviceuser.ErrCredNotExist
		default:
			return err
		}
	}
	return nil
}
