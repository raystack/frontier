package resource

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	CheckPermission(ctx context.Context, userID string, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
}

type ProjectService interface {
	Get(ctx context.Context, id string) (project.Project, error)
}

type Service struct {
	repository       Repository
	configRepository ConfigRepository
	relationService  RelationService
	userService      UserService
	projectService   ProjectService
}

func NewService(repository Repository, configRepository ConfigRepository, relationService RelationService, userService UserService, projectService ProjectService) *Service {
	return &Service{
		repository:       repository,
		configRepository: configRepository,
		relationService:  relationService,
		userService:      userService,
		projectService:   projectService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Resource, error) {
	return s.repository.GetByID(ctx, id)
}

func (s Service) Create(ctx context.Context, res Resource) (Resource, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Resource{}, err
	}

	fetchedProject, err := s.projectService.Get(ctx, res.ProjectID)
	if err != nil {
		return Resource{}, err
	}

	userId := res.UserID
	if strings.TrimSpace(userId) == "" {
		userId = currentUser.ID
	}

	newResource, err := s.repository.Create(ctx, Resource{
		URN:            res.CreateURN(),
		Name:           res.Name,
		OrganizationID: fetchedProject.Organization.ID,
		ProjectID:      res.ProjectID,
		NamespaceID:    res.NamespaceID,
		UserID:         userId,
	})
	if err != nil {
		return Resource{}, err
	}

	if err = s.relationService.DeleteSubjectRelations(ctx, newResource.NamespaceID, newResource.ID); err != nil {
		return Resource{}, err
	}

	if err = s.AddProjectToResource(ctx, newResource.ProjectID, newResource); err != nil {
		return Resource{}, err
	}

	if err = s.AddOrgToResource(ctx, newResource.OrganizationID, newResource); err != nil {
		return Resource{}, err
	}

	return newResource, nil
}

func (s Service) List(ctx context.Context, flt Filter) ([]Resource, error) {
	return s.repository.List(ctx, flt)
}

func (s Service) Update(ctx context.Context, id string, resource Resource) (Resource, error) {
	// TODO there should be an update logic like create here
	return s.repository.Update(ctx, id, resource)
}

func (s Service) AddProjectToResource(ctx context.Context, projectID string, res Resource) error {
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:        res.ID,
			Namespace: res.NamespaceID,
		},
		Subject: relation.Subject{
			RoleID:    schema.ProjectRelationName,
			ID:        projectID,
			Namespace: schema.ProjectNamespace,
		},
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) AddOrgToResource(ctx context.Context, orgID string, res Resource) error {
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:        res.ID,
			Namespace: res.NamespaceID,
		},
		Subject: relation.Subject{
			RoleID:    schema.OrganizationRelationName,
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) GetAllConfigs(ctx context.Context) ([]YAML, error) {
	return s.configRepository.GetAll(ctx)
}

// TODO(krkvrm): Separate Authz for Resources & System Namespaces
func (s Service) CheckAuthz(ctx context.Context, res Resource, act action.Action) (bool, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return false, err
	}

	isSystemNS := namespace.IsSystemNamespaceID(res.NamespaceID)
	fetchedResource := res

	if isSystemNS {
		fetchedResource.ID = res.Name
	} else {
		fetchedResource, err = s.repository.GetByNamespace(ctx, res.Name, res.NamespaceID)
		if err != nil {
			return false, ErrNotExist
		}
	}

	fetchedResourceNS := namespace.Namespace{ID: fetchedResource.NamespaceID}
	return s.relationService.CheckPermission(ctx, currentUser.ID, fetchedResourceNS, fetchedResource.ID, act)
}

func (s Service) Delete(ctx context.Context, namespaceID, id string) error {
	if err := s.relationService.DeleteSubjectRelations(ctx, namespaceID, id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
