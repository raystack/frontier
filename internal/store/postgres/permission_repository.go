package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/shield/core/user"

	"github.com/google/uuid"
	"github.com/raystack/shield/core/permission"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/pkg/db"
)

var (
	ErrInvalidSlug = fmt.Errorf("invalid slug")
)

type PermissionRepository struct {
	dbc *db.Client
}

func NewPermissionRepository(dbc *db.Client) *PermissionRepository {
	return &PermissionRepository{
		dbc: dbc,
	}
}

func (r PermissionRepository) Get(ctx context.Context, id string) (permission.Permission, error) {
	if strings.TrimSpace(id) == "" {
		return permission.Permission{}, permission.ErrInvalidID
	}

	var fetchedPermission Permission
	query, params, err := dialect.Select(&permReturnedColumns{}).From(TABLE_PERMISSIONS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return permission.Permission{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_PERMISSIONS, "Get", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedPermission, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return permission.Permission{}, permission.ErrNotExist
		}
		return permission.Permission{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return fetchedPermission.transformToPermission()
}

func (r PermissionRepository) GetBySlug(ctx context.Context, slug string) (permission.Permission, error) {
	var fetchedPermission Permission
	query, params, err := dialect.Select(&permReturnedColumns{}).From(TABLE_PERMISSIONS).Where(
		goqu.Ex{
			"slug": slug,
		},
	).ToSQL()
	if err != nil {
		return permission.Permission{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_PERMISSIONS, "GetByName", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedPermission, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return permission.Permission{}, permission.ErrNotExist
		}
		return permission.Permission{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return fetchedPermission.transformToPermission()
}

func (r PermissionRepository) Upsert(ctx context.Context, perm permission.Permission) (permission.Permission, error) {
	if strings.TrimSpace(perm.Slug) == "" {
		return permission.Permission{}, ErrInvalidSlug
	}

	marshaledMetadata, err := json.Marshal(perm.Metadata)
	if err != nil {
		return permission.Permission{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	if strings.TrimSpace(perm.ID) == "" {
		perm.ID = uuid.New().String()
	}
	nsID := perm.NamespaceID
	query, params, err := dialect.Insert(TABLE_PERMISSIONS).Rows(
		goqu.Record{
			"id":             perm.ID,
			"name":           perm.Name,
			"slug":           perm.Slug,
			"namespace_name": nsID,
			"metadata":       marshaledMetadata,
		}).OnConflict(goqu.DoUpdate("slug", goqu.Record{
		"metadata": marshaledMetadata,
	})).Returning(&permReturnedColumns{}).ToSQL()
	if err != nil {
		return permission.Permission{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var actionModel Permission
	if err = r.dbc.WithTimeout(ctx, TABLE_PERMISSIONS, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&actionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return permission.Permission{}, namespace.ErrNotExist
		default:
			return permission.Permission{}, err
		}
	}

	return actionModel.transformToPermission()
}

func (r PermissionRepository) List(ctx context.Context, flt permission.Filter) ([]permission.Permission, error) {
	var fetchedActions []Permission
	stmt := dialect.Select(&permReturnedColumns{}).From(TABLE_PERMISSIONS)
	if flt.NamespaceID != "" {
		stmt = stmt.Where(goqu.Ex{
			"namespace_name": flt.NamespaceID,
		})
	}
	if len(flt.Slugs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"slug": goqu.Op{"in": flt.Slugs},
		})
	}

	query, params, err := stmt.ToSQL()
	if err != nil {
		return []permission.Permission{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_PERMISSIONS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedActions, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []permission.Permission{}, nil
		}
		return []permission.Permission{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedPermissions []permission.Permission
	for _, o := range fetchedActions {
		transPerm, err := o.transformToPermission()
		if err != nil {
			return nil, fmt.Errorf("failed to transform permission model: %w", err)
		}
		transformedPermissions = append(transformedPermissions, transPerm)
	}

	return transformedPermissions, nil
}

func (r PermissionRepository) Update(ctx context.Context, act permission.Permission) (permission.Permission, error) {
	if strings.TrimSpace(act.ID) == "" {
		return permission.Permission{}, permission.ErrInvalidID
	}

	if strings.TrimSpace(act.Name) == "" {
		return permission.Permission{}, permission.ErrInvalidDetail
	}

	query, params, err := dialect.Update(TABLE_PERMISSIONS).Set(
		goqu.Record{
			"name":           act.Name,
			"namespace_name": act.NamespaceID,
			"updated_at":     goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": act.ID,
	}).Returning(&permReturnedColumns{}).ToSQL()
	if err != nil {
		return permission.Permission{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var actionModel Permission
	if err = r.dbc.WithTimeout(ctx, TABLE_PERMISSIONS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&actionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return permission.Permission{}, permission.ErrNotExist
		case errors.Is(err, ErrForeignKeyViolation):
			return permission.Permission{}, namespace.ErrNotExist
		default:
			return permission.Permission{}, err
		}
	}

	return actionModel.transformToPermission()
}

func (r PermissionRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_PERMISSIONS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_PERMISSIONS, "Delete", func(ctx context.Context) error {
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
