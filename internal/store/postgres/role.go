package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"

	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/str"
)

type Role struct {
	Id          string         `db:"id"`
	Name        string         `db:"name"`
	Types       pq.StringArray `db:"types"`
	Namespace   Namespace      `db:"namespace"`
	NamespaceID string         `db:"namespace_id"`
	Metadata    []byte         `db:"metadata"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func buildRoleSelectStatement(dialect goqu.DialectWrapper) *goqu.SelectDataset {
	roleSelectStatement := dialect.Select(
		goqu.I("r.id"),
		goqu.I("r.name"),
		goqu.I("r.types"),
		goqu.I("r.namespace_id"),
		goqu.I("r.metadata"),
		goqu.I("namespaces.id").As(goqu.C("namespace.id")),
		goqu.I("namespaces.name").As(goqu.C("namespace.name")),
	).From(goqu.T(TABLE_ROLES).As("r"))

	return roleSelectStatement
}

func buildRoleJoinStatement(selectStatement *goqu.SelectDataset) *goqu.SelectDataset {
	roleJoinStatement := selectStatement.Join(goqu.T(TABLE_NAMESPACE), goqu.On(
		goqu.I("namespaces.id").Eq(goqu.I("r.namespace_id"))))

	return roleJoinStatement
}

func buildCreateRoleQuery(dialect goqu.DialectWrapper) (string, error) {
	createRoleQuery, _, err := dialect.Insert(TABLE_ROLES).Rows(
		goqu.Record{
			"id":           goqu.L("$1"),
			"name":         goqu.L("$2"),
			"types":        goqu.L("$3"),
			"namespace_id": goqu.L("$4"),
			"metadata":     goqu.L("$5"),
		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
		"name": goqu.L("$2"),
	})).Returning("id").ToSQL()

	return createRoleQuery, err
}
func buildGetRoleQuery(dialect goqu.DialectWrapper) (string, error) {
	selectStatement := buildRoleSelectStatement(dialect)
	joinStatement := buildRoleJoinStatement(selectStatement)
	getRoleQuery, _, err := joinStatement.Where(goqu.Ex{
		"r.id": goqu.L("$1"),
	}).ToSQL()

	return getRoleQuery, err
}
func buildListRolesQuery(dialect goqu.DialectWrapper) (string, error) {
	selectStatement := buildRoleSelectStatement(dialect)
	joinStatement := buildRoleJoinStatement(selectStatement)
	listRolesQuery, _, err := joinStatement.ToSQL()

	return listRolesQuery, err
}
func buildUpdateRoleQuery(dialect goqu.DialectWrapper) (string, error) {
	updateRoleQuery, _, err := dialect.Update(TABLE_ROLES).Set(
		goqu.Record{
			"name":         goqu.L("$2"),
			"types":        goqu.L("$3"),
			"namespace_id": goqu.L("$4"),
			"metadata":     goqu.L("$5"),
			"updated_at":   goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).Returning("id").ToSQL()

	return updateRoleQuery, err
}

func (s Store) GetRole(ctx context.Context, id string) (role.Role, error) {
	var fetchedRole Role
	getRoleQuery, err := buildGetRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRole, getRoleQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return role.Role{}, project.ErrNotExist
	} else if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedRole, err := transformToRole(fetchedRole)
	if err != nil {
		return role.Role{}, err
	}

	return transformedRole, nil
}

func (s Store) CreateRole(ctx context.Context, roleToCreate role.Role) (role.Role, error) {
	marshaledMetadata, err := json.Marshal(roleToCreate.Metadata)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newRole Role
	var fetchedRole Role

	nsId := str.DefaultStringIfEmpty(roleToCreate.Namespace.Id, roleToCreate.NamespaceId)
	createRoleQuery, err := buildCreateRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newRole, createRoleQuery, roleToCreate.Id, roleToCreate.Name, pq.StringArray(roleToCreate.Types), nsId, marshaledMetadata)
	})
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	getRoleQuery, err := buildGetRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRole, getRoleQuery, newRole.Id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return role.Role{}, project.ErrNotExist
	} else if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedRole, err := transformToRole(fetchedRole)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedRole, nil
}

func (s Store) ListRoles(ctx context.Context) ([]role.Role, error) {
	var fetchedRoles []Role
	listRolesQuery, err := buildListRolesQuery(dialect)
	if err != nil {
		return []role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedRoles, listRolesQuery)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return []role.Role{}, project.ErrNotExist
	} else if err != nil {
		return []role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRoles []role.Role
	for _, o := range fetchedRoles {
		transformedOrg, err := transformToRole(o)
		if err != nil {
			return []role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedRoles = append(transformedRoles, transformedOrg)
	}

	return transformedRoles, nil
}

func (s Store) UpdateRole(ctx context.Context, toUpdate role.Role) (role.Role, error) {
	var updatedRole Role
	var fetchedRole Role

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	updateRoleQuery, err := buildUpdateRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedRole, updateRoleQuery, toUpdate.Id, toUpdate.Name, pq.StringArray(toUpdate.Types), toUpdate.NamespaceId, marshaledMetadata)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return role.Role{}, role.ErrNotExist
	} else if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	getRoleQuery, err := buildGetRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRole, getRoleQuery, updatedRole.Id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return role.Role{}, project.ErrNotExist
	} else if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToRole(updatedRole)
	if err != nil {
		return role.Role{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func transformToRole(from Role) (role.Role, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return role.Role{}, err
		}
	}

	namespace, err := transformToNamespace(from.Namespace)
	if err != nil {
		return role.Role{}, err
	}
	return role.Role{
		Id:          from.Id,
		Name:        from.Name,
		Types:       from.Types,
		Namespace:   namespace,
		NamespaceId: from.NamespaceID,
		Metadata:    unmarshalledMetadata,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
