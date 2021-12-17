package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/odpf/shield/internal/project"
	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/model"
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

const roleSelectStatement = `r.id, r.name, r.types, r.namespace_id, r.metadata, namespaces.id "namespace.id", namespaces.name "namespace.name"`
const roleJoinStatement = `JOIN namespaces on namespaces.id = r.namespace_id`

var (
	createRoleQuery = `INSERT into roles(id, name, types, namespace_id, metadata) 
		values($1, $2, $3, $4, $5) 
		ON CONFLICT (id) DO UPDATE SET name=$2
		RETURNING id;`
	getRoleQuery    = fmt.Sprintf(`SELECT %s FROM roles r %s WHERE r.id = $1`, roleSelectStatement, roleJoinStatement)
	listRolesQuery  = fmt.Sprintf(`SELECT %s FROM roles r %s`, roleSelectStatement, roleJoinStatement)
	updateRoleQuery = `UPDATE roles SET name = $2, types = $3, namespace_id = $4, metadata = $5, updated_at = now() where id = $1 RETURNING id;`
)

func (s Store) GetRole(ctx context.Context, id string) (model.Role, error) {
	var fetchedRole Role
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRole, getRoleQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Role{}, project.ProjectDoesntExist
	} else if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedRole, err := transformToRole(fetchedRole)
	if err != nil {
		return model.Role{}, err
	}

	return transformedRole, nil
}

func (s Store) CreateRole(ctx context.Context, roleToCreate model.Role) (model.Role, error) {
	marshaledMetadata, err := json.Marshal(roleToCreate.Metadata)
	if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newRole Role
	var fetchedRole Role

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newRole, createRoleQuery, roleToCreate.Id, roleToCreate.Name, pq.StringArray(roleToCreate.Types), roleToCreate.NamespaceId, marshaledMetadata)
	})
	if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRole, getRoleQuery, newRole.Id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Role{}, project.ProjectDoesntExist
	} else if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedRole, err := transformToRole(fetchedRole)
	if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedRole, nil
}

func (s Store) ListRoles(ctx context.Context) ([]model.Role, error) {
	var fetchedRoles []Role
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedRoles, listRolesQuery)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return []model.Role{}, project.ProjectDoesntExist
	} else if err != nil {
		return []model.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRoles []model.Role
	for _, o := range fetchedRoles {
		transformedOrg, err := transformToRole(o)
		if err != nil {
			return []model.Role{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedRoles = append(transformedRoles, transformedOrg)
	}

	return transformedRoles, nil
}

func (s Store) UpdateRole(ctx context.Context, toUpdate model.Role) (model.Role, error) {
	var updatedRole Role
	var fetchedRole Role

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedRole, updateRoleQuery, toUpdate.Id, toUpdate.Name, pq.StringArray(toUpdate.Types), toUpdate.NamespaceId, marshaledMetadata)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Role{}, roles.RoleDoesntExist
	} else if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRole, getRoleQuery, updatedRole.Id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Role{}, project.ProjectDoesntExist
	} else if err != nil {
		return model.Role{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToRole(updatedRole)
	if err != nil {
		return model.Role{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func transformToRole(from Role) (model.Role, error) {
	var unmarshalledMetadata map[string]string
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return model.Role{}, err
		}
	}

	namespace, err := transformToNamespace(from.Namespace)
	if err != nil {
		return model.Role{}, err
	}
	return model.Role{
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
