package resource

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/str"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
}

type Service struct {
	repository       Repository
	configRepository ConfigRepository
	relationService  RelationService
	userService      UserService
}

func NewService(repository Repository, configRepository ConfigRepository, relationService RelationService, userService UserService) *Service {
	return &Service{
		repository:       repository,
		configRepository: configRepository,
		relationService:  relationService,
		userService:      userService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Resource, error) {
	return s.repository.GetByID(ctx, id)
}

func (s Service) Create(ctx context.Context, res Resource) (Resource, error) {
	urn := res.CreateURN()

	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Resource{}, err
	}

	userId := res.UserID
	if strings.TrimSpace(userId) == "" {
		userId = currentUser.ID
	}

	newResource, err := s.repository.Create(ctx, Resource{
		URN:            urn,
		Name:           res.Name,
		OrganizationID: res.OrganizationID,
		ProjectID:      res.ProjectID,
		GroupID:        res.GroupID,
		NamespaceID:    res.NamespaceID,
		UserID:         userId,
	})
	if err != nil {
		return Resource{}, err
	}

	if err = s.relationService.DeleteSubjectRelations(ctx, newResource.NamespaceID, newResource.Idxa); err != nil {
		return Resource{}, err
	}

	if newResource.GroupID != "" {
		if err = s.AddTeamToResource(ctx, group.Group{ID: res.GroupID}, newResource); err != nil {
			return Resource{}, err
		}
	}

	if userId != "" {
		if err = s.AddOwnerToResource(ctx, user.User{ID: userId}, newResource); err != nil {
			return Resource{}, err
		}
	}

	if err = s.AddProjectToResource(ctx, project.Project{ID: res.ProjectID}, newResource); err != nil {
		return Resource{}, err
	}

	if err = s.AddOrgToResource(ctx, organization.Organization{ID: res.OrganizationID}, newResource); err != nil {
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

func (s Service) AddProjectToResource(ctx context.Context, project project.Project, res Resource) error {
	resourceNS := namespace.Namespace{
		ID: res.NamespaceID,
	}

	rel := relation.Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         res.Idxa,
		SubjectID:        project.ID,
		SubjectNamespace: namespace.DefinitionProject,
		Role: role.Role{
			ID:        namespace.DefinitionProject.ID,
			Namespace: resourceNS,
		},
		RelationType: relation.RelationTypes.Namespace,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) AddOrgToResource(ctx context.Context, org organization.Organization, res Resource) error {
	resourceNS := namespace.Namespace{
		ID: res.NamespaceID,
	}

	rel := relation.Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         res.Idxa,
		SubjectID:        org.ID,
		SubjectNamespace: namespace.DefinitionOrg,
		Role: role.Role{
			ID:        namespace.DefinitionOrg.ID,
			Namespace: resourceNS,
		},
		RelationType: relation.RelationTypes.Namespace,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) AddTeamToResource(ctx context.Context, team group.Group, res Resource) error {
	resourceNS := namespace.Namespace{
		ID: res.NamespaceID,
	}

	rel := relation.Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         res.Idxa,
		SubjectID:        team.ID,
		SubjectNamespace: namespace.DefinitionTeam,
		Role: role.Role{
			ID:        namespace.DefinitionTeam.ID,
			Namespace: resourceNS,
		},
		RelationType: relation.RelationTypes.Namespace,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) AddOwnerToResource(ctx context.Context, user user.User, res Resource) error {
	nsId := str.DefaultStringIfEmpty(res.NamespaceID, res.Namespace.ID)

	resourceNS := namespace.Namespace{
		ID: nsId,
	}

	relationSet, err := s.configRepository.GetRelationsForNamespace(ctx, nsId)
	if err != nil {
		return err
	}

	rl := role.GetOwnerRole(resourceNS)

	if !relationSet[rl.ID] {
		return nil
	}

	rel := relation.Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         res.Idxa,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role:             rl,
	}

	if _, err = s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) GetAllConfigs(ctx context.Context) ([]YAML, error) {
	return s.configRepository.GetAll(ctx)
}

func (s Service) CheckAuthz(ctx context.Context, res Resource, act action.Action) (bool, error) {
	user, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return false, err
	}

	res.URN = res.CreateURN()

	isSystemNS := namespace.IsSystemNamespaceID(res.NamespaceID)
	fetchedResource := res

	if isSystemNS {
		fetchedResource.Idxa = res.URN
	} else {
		fetchedResource, err = s.repository.GetByURN(ctx, res.URN)
		if err != nil {
			return false, err
		}
	}

	fetchedResourceNS := namespace.Namespace{
		ID: str.DefaultStringIfEmpty(fetchedResource.NamespaceID, fetchedResource.Namespace.ID),
	}
	return s.relationService.CheckPermission(ctx, user, fetchedResourceNS, fetchedResource.Idxa, act)
}
