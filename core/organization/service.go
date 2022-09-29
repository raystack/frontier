package organization

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/uuid"
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

func (s Service) Get(ctx context.Context, idOrSlug string) (Organization, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.GetByID(ctx, idOrSlug)
	}
	return s.repository.GetBySlug(ctx, idOrSlug)
}

func (s Service) Create(ctx context.Context, org Organization) (Organization, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	}

	newOrg, err := s.repository.Create(ctx, Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})
	if err != nil {
		return Organization{}, err
	}

	if err = s.addAdminToOrg(ctx, currentUser, newOrg); err != nil {
		return Organization{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]Organization, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, org Organization) (Organization, error) {
	if org.ID != "" {
		return s.repository.UpdateByID(ctx, org)
	}
	return s.repository.UpdateBySlug(ctx, org)
}

func (s Service) AddAdmins(ctx context.Context, idOrSlug string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	} else if len(userIds) < 1 {
		return nil, user.ErrInvalidID
	}

	var org Organization
	if uuid.IsValid(idOrSlug) {
		org, err = s.repository.GetByID(ctx, idOrSlug)
	} else {
		org, err = s.repository.GetBySlug(ctx, idOrSlug)
	}
	if err != nil {
		return []user.User{}, err
	}

	isAllowed, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionOrg, org.ID, action.DefinitionManageOrganization)
	if err != nil {
		return []user.User{}, err
	} else if !isAllowed {
		return []user.User{}, errors.ErrForbidden
	}

	users, err := s.userService.GetByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	//TODO might need to check len users < 1

	for _, usr := range users {
		if err = s.addAdminToOrg(ctx, usr, org); err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, org.ID)
}

func (s Service) ListAdmins(ctx context.Context, idOrSlug string) ([]user.User, error) {
	var org Organization
	var err error
	if uuid.IsValid(idOrSlug) {
		return s.repository.ListAdminsByOrgID(ctx, idOrSlug)
	}
	org, err = s.repository.GetBySlug(ctx, idOrSlug)
	if err != nil {
		return []user.User{}, err
	}
	return s.repository.ListAdminsByOrgID(ctx, org.ID)
}

func (s Service) RemoveAdmin(ctx context.Context, idOrSlug string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	}

	var org Organization
	if uuid.IsValid(idOrSlug) {
		org, err = s.repository.GetByID(ctx, idOrSlug)
	} else {
		org, err = s.repository.GetBySlug(ctx, idOrSlug)
	}
	if err != nil {
		return []user.User{}, err
	}

	isAllowed, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionOrg, org.ID, action.DefinitionManageOrganization)
	if err != nil {
		return []user.User{}, err
	} else if !isAllowed {
		return []user.User{}, errors.ErrForbidden
	}

	removedUser, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	if err = s.removeAdminFromOrg(ctx, removedUser, org); err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, org.ID)
}

func (s Service) removeAdminFromOrg(ctx context.Context, user user.User, org Organization) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionOrg,
		ObjectID:         org.ID,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:          role.DefinitionOrganizationAdmin.ID,
			NamespaceID: namespace.DefinitionOrg.ID,
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
			ID:          role.DefinitionOrganizationAdmin.ID,
			NamespaceID: namespace.DefinitionOrg.ID,
		},
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}
