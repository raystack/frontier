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
	"github.com/raystack/shield/core/role"

	"github.com/raystack/shield/core/serviceuser"
	"github.com/raystack/shield/pkg/db"
)

type ServiceUserRepository struct {
	dbc *db.Client
}

func NewServiceUserRepository(dbc *db.Client) *ServiceUserRepository {
	return &ServiceUserRepository{
		dbc: dbc,
	}
}

func (s ServiceUserRepository) List(ctx context.Context, flt serviceuser.Filter) ([]serviceuser.ServiceUser, error) {
	stmt := dialect.Select(
		goqu.I("s.id"),
		goqu.I("s.org_id"),
		goqu.I("s.title"),
		goqu.I("s.state"),
		goqu.I("s.metadata"),
		goqu.I("s.created_at"),
		goqu.I("s.updated_at"),
	)
	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}
	if flt.State != "" {
		stmt = stmt.Where(goqu.Ex{
			"state": flt.State,
		})
	}

	query, params, err := stmt.From(goqu.T(TABLE_SERVICEUSER).As("s")).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedServiceUsers []ServiceUser
	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSER, "List", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &fetchedServiceUsers, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedServiceUsers []serviceuser.ServiceUser
	for _, o := range fetchedServiceUsers {
		transformedServiceUser, err := o.transform()
		if err != nil {
			return nil, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedServiceUsers = append(transformedServiceUsers, transformedServiceUser)
	}

	return transformedServiceUsers, nil
}

func (s ServiceUserRepository) Create(ctx context.Context, serviceUser serviceuser.ServiceUser) (serviceuser.ServiceUser, error) {
	if strings.TrimSpace(serviceUser.ID) == "" {
		serviceUser.ID = uuid.New().String()
	}

	marshaledMetadata, err := json.Marshal(serviceUser.Metadata)
	if err != nil {
		return serviceuser.ServiceUser{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	fetchedServiceUser := ServiceUser{}
	query, params, err := dialect.Insert(TABLE_SERVICEUSER).Rows(
		goqu.Record{
			"id":       serviceUser.ID,
			"org_id":   serviceUser.OrgID,
			"title":    serviceUser.Title,
			"metadata": marshaledMetadata,
		}).OnConflict(
		goqu.DoUpdate("id", goqu.Record{
			"title":    serviceUser.Title,
			"metadata": marshaledMetadata,
		})).Returning(&ServiceUser{}).ToSQL()
	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSER, "Create", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&fetchedServiceUser)
	}); err != nil {
		return serviceuser.ServiceUser{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return fetchedServiceUser.transform()
}

func (s ServiceUserRepository) GetByID(ctx context.Context, id string) (serviceuser.ServiceUser, error) {
	if strings.TrimSpace(id) == "" {
		return serviceuser.ServiceUser{}, role.ErrInvalidID
	}

	query, params, err := dialect.Select(
		goqu.I("s.id"),
		goqu.I("s.org_id"),
		goqu.I("s.title"),
		goqu.I("s.state"),
		goqu.I("s.metadata"),
		goqu.I("s.created_at"),
		goqu.I("s.updated_at"),
	).Where(
		goqu.Ex{"s.id": id},
	).From(goqu.T(TABLE_SERVICEUSER).As("s")).ToSQL()
	if err != nil {
		return serviceuser.ServiceUser{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var serviceUserModel ServiceUser
	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSER, "Get", func(ctx context.Context) error {
		return s.dbc.GetContext(ctx, &serviceUserModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return serviceuser.ServiceUser{}, role.ErrNotExist
		}
		return serviceuser.ServiceUser{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return serviceUserModel.transform()
}

// GetByIDs returns a list of service users by their IDs.
func (s ServiceUserRepository) GetByIDs(ctx context.Context, ids []string) ([]serviceuser.ServiceUser, error) {
	if len(ids) == 0 {
		return nil, role.ErrInvalidID
	}

	query, params, err := dialect.Select(
		goqu.I("s.id"),
		goqu.I("s.org_id"),
		goqu.I("s.title"),
		goqu.I("s.state"),
		goqu.I("s.metadata"),
		goqu.I("s.created_at"),
		goqu.I("s.updated_at"),
	).Where(
		goqu.Ex{"s.id": goqu.Op{"in": ids}},
	).From(goqu.T(TABLE_SERVICEUSER).As("s")).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedUsers []ServiceUser
	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSER, "Get", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, role.ErrNotExist
		}
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []serviceuser.ServiceUser
	for _, u := range fetchedUsers {
		transformedUser, err := u.transform()
		if err != nil {
			return nil, fmt.Errorf("failed to transform user: %w", err)
		}
		transformedUsers = append(transformedUsers, transformedUser)
	}
	return transformedUsers, nil
}

func (s ServiceUserRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_SERVICEUSER).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_SERVICEUSER, "Delete", func(ctx context.Context) error {
		if _, err = s.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return serviceuser.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
