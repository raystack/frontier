package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/user"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/db"
)

type RoleRepository struct {
	dbc *db.Client
}

func NewRoleRepository(dbc *db.Client) *RoleRepository {
	return &RoleRepository{
		dbc: dbc,
	}
}

func (r RoleRepository) buildListQuery(dialect goqu.DialectWrapper) *goqu.SelectDataset {
	roleSelectStatement := dialect.Select(
		goqu.I("r.id"),
		goqu.I("r.org_id"),
		goqu.I("r.name"),
		goqu.I("r.permissions"),
		goqu.I("r.state"),
		goqu.I("r.metadata"),
	).From(goqu.T(TABLE_ROLES).As("r"))
	return roleSelectStatement
}

func (r RoleRepository) Get(ctx context.Context, id string) (role.Role, error) {
	if strings.TrimSpace(id) == "" {
		return role.Role{}, role.ErrInvalidID
	}

	query, params, err := r.buildListQuery(dialect).
		Where(
			goqu.Ex{"r.id": id},
		).ToSQL()
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var roleModel Role
	if err = r.dbc.WithTimeout(ctx, TABLE_ROLES, "Get", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &roleModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role.Role{}, role.ErrNotExist
		}
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return roleModel.transformToRole()
}

func (r RoleRepository) GetByName(ctx context.Context, orgID, name string) (role.Role, error) {
	if strings.TrimSpace(name) == "" {
		return role.Role{}, role.ErrInvalidDetail
	}
	if len(orgID) == 0 {
		orgID = uuid.Nil.String()
	}
	query, params, err := r.buildListQuery(dialect).
		Where(
			goqu.Ex{"r.name": name},
			goqu.Ex{"r.org_id": orgID},
		).ToSQL()
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var roleModel Role
	if err = r.dbc.WithTimeout(ctx, TABLE_ROLES, "GetByName", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &roleModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role.Role{}, role.ErrNotExist
		}
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return roleModel.transformToRole()
}

func (r RoleRepository) Upsert(ctx context.Context, rl role.Role) (string, error) {
	if strings.TrimSpace(rl.ID) == "" {
		rl.ID = uuid.New().String()
	}
	if strings.TrimSpace(rl.Name) == "" {
		return "", role.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(rl.Metadata)
	if err != nil {
		return "", fmt.Errorf("%w: %s", parseErr, err)
	}

	marshaledPermissions, err := json.Marshal(rl.Permissions)
	if err != nil {
		return "", fmt.Errorf("%w: %s", parseErr, err)
	}

	query, _, err := dialect.Insert(TABLE_ROLES).Rows(
		goqu.Record{
			"id":          goqu.L("$1"),
			"org_id":      goqu.L("$2"),
			"name":        goqu.L("$3"),
			"permissions": goqu.L("$4"),
			"state":       goqu.L("$5"),
			"metadata":    goqu.L("$6"),
		}).OnConflict(goqu.DoUpdate("org_id, name", goqu.Record{
		"permissions": goqu.L("$4"),
		"state":       goqu.L("$5"),
		"metadata":    goqu.L("$6"),
	})).Returning("id").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var roleID string
	if err = r.dbc.WithTimeout(ctx, TABLE_ROLES, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, rl.ID, rl.OrgID, rl.Name, marshaledPermissions, rl.State, marshaledMetadata).Scan(&roleID)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return "", role.ErrConflict
		case errors.Is(err, ErrForeignKeyViolation):
			return "", role.ErrInvalidDetail
		default:
			return "", err
		}
	}

	return roleID, nil
}

func (r RoleRepository) List(ctx context.Context, flt role.Filter) ([]role.Role, error) {
	stmt := r.buildListQuery(dialect)
	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}

	query, params, err := stmt.ToSQL()
	if err != nil {
		return []role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRoles []Role
	if err = r.dbc.WithTimeout(ctx, TABLE_ROLES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRoles, query, params...)
	}); err != nil {
		return []role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRoles []role.Role
	for _, o := range fetchedRoles {
		transformedOrg, err := o.transformToRole()
		if err != nil {
			return []role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedRoles = append(transformedRoles, transformedOrg)
	}

	return transformedRoles, nil
}

func (r RoleRepository) Update(ctx context.Context, rl role.Role) (string, error) {
	if strings.TrimSpace(rl.ID) == "" {
		return "", role.ErrInvalidID
	}
	if strings.TrimSpace(rl.Name) == "" {
		return "", role.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(rl.Metadata)
	if err != nil {
		return "", fmt.Errorf("%w: %s", parseErr, err)
	}
	marshaledPermissions, err := json.Marshal(rl.Permissions)
	if err != nil {
		return "", fmt.Errorf("%w: %s", parseErr, err)
	}

	query, _, err := dialect.Update(TABLE_ROLES).Set(
		goqu.Record{
			"name":        goqu.L("$2"),
			"permissions": goqu.L("$3"),
			"state":       goqu.L("$4"),
			"metadata":    goqu.L("$5"),
			"updated_at":  goqu.L("now()"),
		}).Where(
		goqu.Ex{"id": goqu.L("$1")},
	).Returning("id").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var roleID string
	if err = r.dbc.WithTimeout(ctx, TABLE_ROLES, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, rl.ID, rl.Name, marshaledPermissions, rl.State, marshaledMetadata).Scan(&roleID)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", role.ErrNotExist
		case errors.Is(err, ErrForeignKeyViolation):
			return "", namespace.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return "", role.ErrConflict
		default:
			return "", err
		}
	}

	return roleID, nil
}

func (r RoleRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_ROLES).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_ROLES, "Delete", func(ctx context.Context) error {
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
