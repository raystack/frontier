package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"database/sql"

	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
)

type RoleRepository struct {
	dbc *db.Client
}

func NewRoleRepository(dbc *db.Client) *RoleRepository {
	return &RoleRepository{
		dbc: dbc,
	}
}

func (r RoleRepository) Get(ctx context.Context, id string) (role.Role, error) {
	var fetchedRole Role
	getRoleQuery, err := buildGetRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedRole, getRoleQuery, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role.Role{}, project.ErrNotExist
		}
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedRole, err := transformToRole(fetchedRole)
	if err != nil {
		return role.Role{}, err
	}

	return transformedRole, nil
}

func (r RoleRepository) Create(ctx context.Context, roleToCreate role.Role) (role.Role, error) {
	marshaledMetadata, err := json.Marshal(roleToCreate.Metadata)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newRole Role
	var fetchedRole Role

	nsID := str.DefaultStringIfEmpty(roleToCreate.Namespace.ID, roleToCreate.NamespaceID)
	createRoleQuery, err := buildCreateRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &newRole, createRoleQuery, roleToCreate.ID, roleToCreate.Name, roleToCreate.Types, nsID, marshaledMetadata)
	}); err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	getRoleQuery, err := buildGetRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedRole, getRoleQuery, newRole.ID)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role.Role{}, project.ErrNotExist
		}
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedRole, err := transformToRole(fetchedRole)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedRole, nil
}

func (r RoleRepository) List(ctx context.Context) ([]role.Role, error) {
	var fetchedRoles []Role
	listRolesQuery, err := buildListRolesQuery(dialect)
	if err != nil {
		return []role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRoles, listRolesQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []role.Role{}, project.ErrNotExist
		}
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

func (r RoleRepository) Update(ctx context.Context, toUpdate role.Role) (role.Role, error) {
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

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &updatedRole, updateRoleQuery, toUpdate.ID, toUpdate.Name, toUpdate.Types, toUpdate.NamespaceID, marshaledMetadata)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role.Role{}, role.ErrNotExist
		}
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	getRoleQuery, err := buildGetRoleQuery(dialect)
	if err != nil {
		return role.Role{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedRole, getRoleQuery, updatedRole.ID)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role.Role{}, project.ErrNotExist
		}
		return role.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToRole(updatedRole)
	if err != nil {
		return role.Role{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}
