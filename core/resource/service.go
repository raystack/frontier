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
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
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
		NamespaceID:    res.NamespaceID,
		UserID:         userId,
	})
	if err != nil {
		return Resource{}, err
	}

	if err = s.relationService.DeleteSubjectRelations(ctx, newResource.NamespaceID, newResource.Idxa); err != nil {
		return Resource{}, err
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
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:          res.Idxa,
			NamespaceID: res.NamespaceID,
		},
		Subject: relation.Subject{
			RoleID:    schema.ProjectRelationName,
			ID:        project.ID,
			Namespace: schema.ProjectNamespace,
		},
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) AddOrgToResource(ctx context.Context, org organization.Organization, res Resource) error {
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:          res.Idxa,
			NamespaceID: res.NamespaceID,
		},
		Subject: relation.Subject{
			RoleID:    schema.OrganizationRelationName,
			ID:        org.ID,
			Namespace: schema.OrganizationNamespace,
		},
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) AddTeamToResource(ctx context.Context, team group.Group, res Resource) error {
	//resourceNS := namespace.Namespace{
	//	ID: res.NamespaceID,
	//}
	//
	//rel := relation.Relation{
	//	ObjectNamespace:  resourceNS,
	//	ObjectID:         res.Idxa,
	//	SubjectID:        team.ID,
	//	SubjectNamespace: namespace.DefinitionTeam,
	//	Role: role.Role{
	//		ID:          namespace.DefinitionTeam.ID,
	//		NamespaceID: resourceNS.ID,
	//	},
	//	RelationType: relation.RelationTypes.Namespace,
	//}
	//if _, err := s.relationService.Create(ctx, rel); err != nil {
	//	return err
	//}

	return nil
}

func (s Service) AddOwnerToResource(ctx context.Context, user user.User, res Resource) error {
	//nsId := str.DefaultStringIfEmpty(res.NamespaceID, res.Namespace.ID)
	//
	//resourceNS := namespace.Namespace{
	//	ID: nsId,
	//}
	//
	//relationSet, err := s.configRepository.GetRelationsForNamespace(ctx, nsId)
	//if err != nil {
	//	return err
	//}
	//
	//rl := role.GetOwnerRole(resourceNS)
	//
	//if !relationSet[rl.ID] {
	//	return nil
	//}
	//
	//rel := relation.Relation{
	//	ObjectNamespace:  resourceNS,
	//	ObjectID:         res.Idxa,
	//	SubjectID:        user.ID,
	//	SubjectNamespace: namespace.DefinitionUser,
	//	Role:             rl,
	//}
	//
	//if _, err := s.relationService.Create(ctx, rel); err != nil {
	//	return err
	//}

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
		fetchedResource.Idxa = res.Name
	} else {
		fetchedResource, err = s.repository.GetByNamespace(ctx, res.Name, res.NamespaceID)
		if err != nil {
			return false, err
		}
	}

	fetchedResourceNS := namespace.Namespace{ID: fetchedResource.NamespaceID}
	return s.relationService.CheckPermission(ctx, currentUser, fetchedResourceNS, fetchedResource.Idxa, act)
}
