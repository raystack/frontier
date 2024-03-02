package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	CheckPermission(ctx context.Context, rel relation.Relation) (bool, error)
	BatchCheckPermission(ctx context.Context, relations []relation.Relation) ([]relation.CheckPair, error)
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
		ID:            res.ID,
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

func (s Service) CheckAuthz(ctx context.Context, check Check) (bool, error) {
	relSubject, err := s.buildRelationSubject(ctx, check.Subject)
	if err != nil {
		return false, err
	}

	relObject, err := s.buildRelationObject(ctx, check.Object)
	if err != nil {
		return false, err
	}

	return s.relationService.CheckPermission(ctx, relation.Relation{
		Subject:      relSubject,
		Object:       relObject,
		RelationName: check.Permission,
	})
}

func (s Service) buildRelationSubject(ctx context.Context, sub relation.Subject) (relation.Subject, error) {
	// use existing if passed in request
	if sub.ID != "" && sub.Namespace != "" {
		return sub, nil
	}

	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return relation.Subject{}, err
	}
	return relation.Subject{
		ID:        principal.ID,
		Namespace: principal.Type,
	}, nil
}

func (s Service) buildRelationObject(ctx context.Context, obj relation.Object) (relation.Object, error) {
	// a user can pass object name instead of id in the request
	// we should convert name to id based on object namespace
	if !utils.IsValidUUID(obj.ID) {
		if schema.IsSystemNamespace(obj.Namespace) {
			if obj.Namespace == schema.ProjectNamespace {
				// if object is project, then fetch project by name
				project, err := s.projectService.Get(ctx, obj.ID)
				if err != nil {
					return obj, err
				}
				obj.ID = project.ID
			}
			if obj.Namespace == schema.OrganizationNamespace {
				// if object is org, then fetch org by name
				org, err := s.orgService.Get(ctx, obj.ID)
				if err != nil {
					return obj, err
				}
				obj.ID = org.ID
			}
		} else {
			// if not a system namespace it could be a resource, so fetch resource by urn
			resource, err := s.Get(ctx, obj.ID)
			if err != nil {
				return obj, err
			}
			obj.ID = resource.ID
		}
	}
	return obj, nil
}

func (s Service) BatchCheck(ctx context.Context, checks []Check) ([]relation.CheckPair, error) {
	relations := make([]relation.Relation, 0, len(checks))
	for _, check := range checks {
		// we can parallelize this to speed up the process
		relObject, err := s.buildRelationObject(ctx, check.Object)
		if err != nil {
			return nil, err
		}

		relSubject, err := s.buildRelationSubject(ctx, check.Subject)
		if err != nil {
			return nil, err
		}
		relations = append(relations, relation.Relation{
			Subject:      relSubject,
			Object:       relObject,
			RelationName: check.Permission,
		})
	}
	return s.relationService.BatchCheckPermission(ctx, relations)
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
