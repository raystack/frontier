package organization

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
}

type Service struct {
	store           Store
	relationService RelationService
	userService     UserService
}

func NewService(store Store, relationService RelationService, userService UserService) *Service {
	return &Service{
		store:           store,
		relationService: relationService,
		userService:     userService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Organization, error) {
	return s.store.GetOrg(ctx, id)
}

func (s Service) Create(ctx context.Context, org Organization) (Organization, error) {
	user, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Organization{}, err
	}

	newOrg, err := s.store.CreateOrg(ctx, Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})
	if err != nil {
		return Organization{}, err
	}

	err = s.AddAdminToOrg(ctx, user, newOrg)
	if err != nil {
		return Organization{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]Organization, error) {
	return s.store.ListOrg(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate Organization) (Organization, error) {
	return s.store.UpdateOrg(ctx, toUpdate)
}

func (s Service) AddAdmin(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	org, err := s.store.GetOrg(ctx, id)
	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionOrg, org.ID, action.DefinitionManageOrganization)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	users, err := s.store.GetUsersByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		err = s.AddAdminToOrg(ctx, usr, org)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, id)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return s.store.ListOrgAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, id string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	org, err := s.store.GetOrg(ctx, id)

	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionOrg, org.ID, action.DefinitionManageOrganization)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	usr, err := s.store.GetUser(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	err = s.RemoveAdminFromOrg(ctx, usr, org)
	if err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, id)
}

func (s Service) RemoveAdminFromOrg(ctx context.Context, user user.User, org Organization) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionOrg,
		ObjectID:         org.ID,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionOrganizationAdmin.ID,
			Namespace: namespace.DefinitionOrg,
		},
	}
	return s.relationService.Delete(ctx, rel)
}

func (s Service) AddAdminToOrg(ctx context.Context, user user.User, org Organization) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionOrg,
		ObjectID:         org.ID,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionOrganizationAdmin.ID,
			Namespace: namespace.DefinitionOrg,
		},
	}
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}
