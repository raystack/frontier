package role

import (
	"context"
	"fmt"

	"github.com/raystack/shield/pkg/utils"

	"github.com/raystack/shield/core/permission"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type PermissionService interface {
	Get(ctx context.Context, id string) (permission.Permission, error)
}

type Service struct {
	repository        Repository
	relationService   RelationService
	permissionService PermissionService
}

func NewService(repository Repository, relationService RelationService, permissionService PermissionService) *Service {
	return &Service{
		repository:        repository,
		relationService:   relationService,
		permissionService: permissionService,
	}
}

func (s Service) Upsert(ctx context.Context, toCreate Role) (Role, error) {
	for idx, permName := range toCreate.Permissions {
		// verify if perm exists
		if perm, err := s.permissionService.Get(ctx, permName); err != nil {
			return Role{}, fmt.Errorf("%s: %w", permName, err)
		} else {
			toCreate.Permissions[idx] = perm.GenerateSlug()
		}
	}

	roleID, err := s.repository.Upsert(ctx, toCreate)
	if err != nil {
		return Role{}, err
	}

	// create relation between role and permissions
	// for example for each permission:
	// app/role:org_owner#organization_delete@app/user:*
	// app/role:org_owner#organization_update@app/user:*
	// this needs to be created for each type of principles
	for _, perm := range toCreate.Permissions {
		_, err = s.relationService.Create(ctx, relation.Relation{
			Object: relation.Object{
				ID:        roleID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principles who have role will have access
				Namespace: schema.UserPrincipal,
			},
			RelationName: perm,
		})
		if err != nil {
			return Role{}, err
		}
		// do the same with service user
		_, err = s.relationService.Create(ctx, relation.Relation{
			Object: relation.Object{
				ID:        roleID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principles who have role will have access
				Namespace: schema.ServiceUserPrincipal,
			},
			RelationName: perm,
		})
		if err != nil {
			return Role{}, err
		}
	}

	return s.repository.Get(ctx, roleID)
}

func (s Service) Get(ctx context.Context, id string) (Role, error) {
	if utils.IsValidUUID(id) {
		return s.repository.Get(ctx, id)
	}
	// passing empty orgID will return roles created by system
	return s.repository.GetByName(ctx, "", id)
}

func (s Service) List(ctx context.Context, f Filter) ([]Role, error) {
	return s.repository.List(ctx, f)
}

func (s Service) Update(ctx context.Context, toUpdate Role) (Role, error) {
	for idx, permName := range toUpdate.Permissions {
		// verify if perm exists
		if perm, err := s.permissionService.Get(ctx, permName); err != nil {
			return Role{}, fmt.Errorf("%s: %w", permName, err)
		} else {
			toUpdate.Permissions[idx] = perm.Slug
		}
	}

	roleID, err := s.repository.Update(ctx, toUpdate)
	if err != nil {
		return Role{}, err
	}
	return s.repository.Get(ctx, roleID)
}

func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.RoleNamespace,
	}}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
