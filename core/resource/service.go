package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/shield/core/authenticate"

	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/pkg/utils"

	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	CheckPermission(ctx context.Context, subject relation.Subject, object relation.Object, permName string) (bool, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type ProjectService interface {
	Get(ctx context.Context, idOrName string) (project.Project, error)
}

type OrgService interface {
	Get(ctx context.Context, idOrName string) (organization.Organization, error)
}

type Service struct {
	repository       Repository
	configRepository ConfigRepository
	relationService  RelationService
	authnService     AuthnService
	projectService   ProjectService
	orgService       OrgService
}

func NewService(repository Repository, configRepository ConfigRepository,
	relationService RelationService, authnService AuthnService,
	projectService ProjectService, orgService OrgService) *Service {
	return &Service{
		repository:       repository,
		configRepository: configRepository,
		relationService:  relationService,
		authnService:     authnService,
		projectService:   projectService,
		orgService:       orgService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Resource, error) {
	if utils.IsValidUUID(id) {
		return s.repository.GetByID(ctx, id)
	}
	return s.repository.GetByURN(ctx, id)
}

func (s Service) Create(ctx context.Context, res Resource) (Resource, error) {
	// TODO(kushsharma): currently we allow users to pass a principal in request which allow
	// them to create resource on behalf of other users. Should we only allow this for admins?
	principalID := res.PrincipalID
	principalType := res.PrincipalType
	if strings.TrimSpace(principalID) == "" {
		principal, err := s.authnService.GetPrincipal(ctx)
		if err != nil {
			return Resource{}, err
		}
		principalID = principal.ID
		principalType = principal.Type
	}

	resourceProject, err := s.projectService.Get(ctx, res.ProjectID)
	if err != nil {
		return Resource{}, fmt.Errorf("failed to get project: %w", err)
	}

	newResource, err := s.repository.Create(ctx, Resource{
		URN:           res.CreateURN(resourceProject.Name),
		Name:          res.Name,
		Title:         res.Title,
		ProjectID:     resourceProject.ID,
		NamespaceID:   res.NamespaceID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
		Metadata:      res.Metadata,
	})
	if err != nil {
		return Resource{}, err
	}

	if err = s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        newResource.ID,
			Namespace: newResource.NamespaceID,
		},
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return Resource{}, err
	}

	if err = s.AddProjectToResource(ctx, newResource.ProjectID, newResource); err != nil {
		return Resource{}, err
	}
	if err = s.AddResourceOwner(ctx, newResource); err != nil {
		return Resource{}, err
	}

	return newResource, nil
}

func (s Service) List(ctx context.Context, flt Filter) ([]Resource, error) {
	return s.repository.List(ctx, flt)
}

func (s Service) Update(ctx context.Context, resource Resource) (Resource, error) {
	return s.repository.Update(ctx, resource)
}

func (s Service) AddProjectToResource(ctx context.Context, projectID string, res Resource) error {
	rel := relation.Relation{
		Object: relation.Object{
			ID:        res.ID,
			Namespace: res.NamespaceID,
		},
		Subject: relation.Subject{
			ID:        projectID,
			Namespace: schema.ProjectNamespace,
		},
		RelationName: schema.ProjectRelationName,
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) AddResourceOwner(ctx context.Context, res Resource) error {
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        res.ID,
			Namespace: res.NamespaceID,
		},
		Subject: relation.Subject{
			ID:        res.PrincipalID,
			Namespace: res.PrincipalType,
		},
		RelationName: schema.OwnerRelationName,
	}); err != nil {
		return err
	}
	return nil
}

func (s Service) CheckAuthz(ctx context.Context, rel relation.Object, permissionName string) (bool, error) {
	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return false, err
	}

	// a user can pass object name instead of id in the request
	// we should convert name to id based on object namespace
	if !utils.IsValidUUID(rel.ID) {
		if schema.IsSystemNamespace(rel.Namespace) {
			if rel.Namespace == schema.ProjectNamespace {
				// if object is project, then fetch project by name
				project, err := s.projectService.Get(ctx, rel.ID)
				if err != nil {
					return false, err
				}
				rel.ID = project.ID
			}
			if rel.Namespace == schema.OrganizationNamespace {
				// if object is org, then fetch org by name
				org, err := s.orgService.Get(ctx, rel.ID)
				if err != nil {
					return false, err
				}
				rel.ID = org.ID
			}
		} else {
			// if not a system namespace it could be a resource, so fetch resource by urn
			resource, err := s.Get(ctx, rel.ID)
			if err != nil {
				return false, err
			}
			rel.ID = resource.ID
		}
	}

	return s.relationService.CheckPermission(ctx, relation.Subject{
		ID:        principal.ID,
		Namespace: principal.Type,
	}, rel, permissionName)
}

func (s Service) Delete(ctx context.Context, namespaceID, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id,
			Namespace: namespaceID,
		},
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return s.repository.Delete(ctx, id)
}
