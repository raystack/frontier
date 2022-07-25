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
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
}

func NewService(repository Repository, relationService RelationService, userService UserService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Organization, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) Create(ctx context.Context, org Organization) (Organization, error) {
	user, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Organization{}, err
	}

	newOrg, err := s.repository.Create(ctx, Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})
	if err != nil {
		return Organization{}, err
	}

	if err = s.addAdminToOrg(ctx, user, newOrg); err != nil {
		return Organization{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]Organization, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate Organization) (Organization, error) {
	return s.repository.Update(ctx, toUpdate)
}

func (s Service) AddAdmin(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	org, err := s.repository.Get(ctx, id)
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

	users, err := s.userService.GetByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		if err = s.addAdminToOrg(ctx, usr, org); err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, id)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return s.repository.ListAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, id string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	org, err := s.repository.Get(ctx, id)
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

	usr, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	if err = s.RemoveAdminFromOrg(ctx, usr, org); err != nil {
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

func (s Service) addAdminToOrg(ctx context.Context, user user.User, org Organization) error {
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
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}
