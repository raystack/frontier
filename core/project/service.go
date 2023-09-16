package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/serviceuser"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
	IsSudo(ctx context.Context, id string) (bool, error)
}

type ServiceuserService interface {
	GetByIDs(ctx context.Context, ids []string) ([]serviceuser.ServiceUser, error)
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) (policy.Policy, error)
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
	suserService    ServiceuserService
	policyService   PolicyService
	authnService    AuthnService
}

func NewService(repository Repository, relationService RelationService, userService UserService,
	policyService PolicyService, authnService AuthnService, suserService ServiceuserService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
		policyService:   policyService,
		authnService:    authnService,
		suserService:    suserService,
	}
}

func (s Service) Get(ctx context.Context, idOrName string) (Project, error) {
	if utils.IsValidUUID(idOrName) {
		return s.repository.GetByID(ctx, idOrName)
	}
	return s.repository.GetByName(ctx, idOrName)
}

func (s Service) GetByIDs(ctx context.Context, ids []string, flt Filter) ([]Project, error) {
	return s.repository.GetByIDs(ctx, ids, flt)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	currentPrincipal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Project{}, err
	}

	newProject, err := s.repository.Create(ctx, prj)
	if err != nil {
		return Project{}, err
	}

	if err = s.addProjectToOrg(ctx, newProject, prj.Organization.ID); err != nil {
		return Project{}, err
	}

	// make user administrator of the project
	if _, err = s.policyService.Create(ctx, policy.Policy{
		RoleID:        OwnerRole,
		ResourceID:    newProject.ID,
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   currentPrincipal.ID,
		PrincipalType: currentPrincipal.Type,
	}); err != nil {
		return Project{}, fmt.Errorf("failed to create owner policy for project %s: %w", newProject.ID, err)
	}
	return newProject, nil
}

func (s Service) List(ctx context.Context, f Filter) ([]Project, error) {
	return s.repository.List(ctx, f)
}

func (s Service) ListByUser(ctx context.Context, userID string, flt Filter) ([]Project, error) {
	requestedUser, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	projIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.UserPrincipal,
			ID:        requestedUser.ID,
		},
		RelationName: MemberPermission,
	})
	if err != nil {
		return nil, err
	}
	if len(projIDs) == 0 {
		return []Project{}, nil
	}
	return s.GetByIDs(ctx, projIDs, flt)
}

func (s Service) Update(ctx context.Context, prj Project) (Project, error) {
	if utils.IsValidUUID(prj.ID) {
		return s.repository.UpdateByID(ctx, prj)
	}
	return s.repository.UpdateByName(ctx, prj)
}

func (s Service) ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error) {
	requestedProject, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.Relation{
		Object: relation.Object{
			ID:        requestedProject.ID,
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.UserPrincipal,
		},
		RelationName: permissionFilter,
	})
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []user.User{}, nil
	}

	// filter superusers from the list of users who have the permission
	// TODO(kushsharma): checking sudo one by one is slow, we need a batch test
	nonSuperUserIDs := make([]string, 0)
	for _, userID := range userIDs {
		isSudo, err := s.userService.IsSudo(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !isSudo {
			nonSuperUserIDs = append(nonSuperUserIDs, userID)
		}
	}

	return s.userService.GetByIDs(ctx, nonSuperUserIDs)
}

func (s Service) ListServiceUsers(ctx context.Context, id string, permissionFilter string) ([]serviceuser.ServiceUser, error) {
	requestedProject, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.Relation{
		Object: relation.Object{
			ID:        requestedProject.ID,
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: permissionFilter,
	})
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []serviceuser.ServiceUser{}, nil
	}
	return s.suserService.GetByIDs(ctx, userIDs)
}

func (s Service) addProjectToOrg(ctx context.Context, prj Project, orgID string) error {
	rel := relation.Relation{
		Object: relation.Object{
			ID:        prj.ID,
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

// DeleteModel doesn't delete the nested resource, only itself
func (s Service) DeleteModel(ctx context.Context, id string) error {
	// delete all relations where resource is an object
	// all relations where project is an subject should already been deleted by now
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.ProjectNamespace,
	}}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return s.repository.Delete(ctx, id)
}
