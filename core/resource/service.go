package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/core/project"
	patmodels "github.com/raystack/frontier/core/userpat/models"
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

type PATService interface {
	GetByID(ctx context.Context, id string) (patmodels.PAT, error)
}

type Service struct {
	repository       Repository
	configRepository ConfigRepository
	relationService  RelationService
	authnService     AuthnService
	projectService   ProjectService
	orgService       OrgService
	patService       PATService
}

func NewService(repository Repository, configRepository ConfigRepository,
	relationService RelationService, authnService AuthnService,
	projectService ProjectService, orgService OrgService,
	patService PATService) *Service {
	return &Service{
		repository:       repository,
		configRepository: configRepository,
		relationService:  relationService,
		authnService:     authnService,
		projectService:   projectService,
		orgService:       orgService,
		patService:       patService,
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
	relObject, err := s.buildRelationObject(ctx, check.Object)
	if err != nil {
		return false, err
	}

	// PAT scope — early exit if denied
	if allowed, err := s.checkPATScope(ctx, check.Subject, relObject, check.Permission); err != nil || !allowed {
		return false, err
	}

	relSubject, err := s.buildRelationSubject(ctx, check.Subject)
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
		// PAT subject → resolve to underlying user for authorization
		if sub.Namespace == schema.PATPrincipal {
			return s.resolvePATUser(ctx, sub.ID)
		}
		return sub, nil
	}

	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return relation.Subject{}, err
	}
	// PAT principal → use underlying user for authorization
	if principal.PAT != nil {
		return relation.Subject{ID: principal.PAT.UserID, Namespace: schema.UserPrincipal}, nil
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

// resolvePATUser resolves a PAT ID to its owning user subject.
// Tries context first(cached), falls back to DB (for federated checks with explicit subject).
func (s Service) resolvePATUser(ctx context.Context, patID string) (relation.Subject, error) {
	principal, err := s.authnService.GetPrincipal(ctx)
	if err == nil && principal.PAT != nil && principal.PAT.ID == patID {
		return relation.Subject{ID: principal.PAT.UserID, Namespace: schema.UserPrincipal}, nil
	}

	pat, err := s.patService.GetByID(ctx, patID)
	if err != nil {
		return relation.Subject{}, err
	}
	return relation.Subject{ID: pat.UserID, Namespace: schema.UserPrincipal}, nil
}

// resolvePATID returns the PAT ID to scope-check, if any.
// Explicit app/pat subject takes precedence (federated check by admin),
// otherwise falls back to the authenticated principal's PAT.
func (s Service) resolvePATID(ctx context.Context, subject relation.Subject) string {
	if subject.Namespace == schema.PATPrincipal && subject.ID != "" {
		return subject.ID
	}
	principal, _ := s.authnService.GetPrincipal(ctx)
	if principal.PAT != nil {
		return principal.PAT.ID
	}
	return ""
}

// checkPATScope checks if the PAT has scope for the given permission on the object.
// Returns (true, nil) if no PAT is involved.
func (s Service) checkPATScope(ctx context.Context, subject relation.Subject, object relation.Object, permission string) (bool, error) {
	patID := s.resolvePATID(ctx, subject)
	if patID == "" {
		return true, nil
	}
	return s.relationService.CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Object:       object,
		RelationName: permission,
	})
}

// BatchCheck checks permissions for multiple resource checks.
// For PAT requests, it first batch-checks PAT scope, then only runs user permission
// checks for scope-allowed items. Scope-denied items return false directly.
func (s Service) BatchCheck(ctx context.Context, checks []Check) ([]relation.CheckPair, error) {
	relations, patScopeRelations, patScopeIdx, err := s.buildBatchRelations(ctx, checks)
	if err != nil {
		return nil, err
	}

	// no PAT involved — straight to user permission check
	if len(patScopeRelations) == 0 {
		return s.relationService.BatchCheckPermission(ctx, relations)
	}

	// PAT scope gate — check which items the PAT has scope for
	scopeDenied, err := s.batchCheckPATScope(ctx, patScopeRelations, patScopeIdx)
	if err != nil {
		return nil, err
	}

	// run user permission checks only for scope-allowed items, merge results
	return s.batchCheckWithScopeFilter(ctx, relations, scopeDenied)
}

// buildBatchRelations resolves objects/subjects and builds parallel PAT scope relations.
// Returns user relations, PAT scope relations, and index mapping from scope back to user relations.
func (s Service) buildBatchRelations(ctx context.Context, checks []Check) (
	relations, patScopeRelations []relation.Relation, patScopeIdx []int, err error,
) {
	relations = make([]relation.Relation, 0, len(checks))
	for i, check := range checks {
		relObject, err := s.buildRelationObject(ctx, check.Object)
		if err != nil {
			return nil, nil, nil, err
		}
		relSubject, err := s.buildRelationSubject(ctx, check.Subject)
		if err != nil {
			return nil, nil, nil, err
		}
		relations = append(relations, relation.Relation{
			Subject:      relSubject,
			Object:       relObject,
			RelationName: check.Permission,
		})

		if patID := s.resolvePATID(ctx, check.Subject); patID != "" {
			patScopeRelations = append(patScopeRelations, relation.Relation{
				Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
				Object:       relObject,
				RelationName: check.Permission,
			})
			patScopeIdx = append(patScopeIdx, i)
		}
	}
	return relations, patScopeRelations, patScopeIdx, nil
}

// batchCheckPATScope runs a batch scope check and returns the set of denied relation indices.
func (s Service) batchCheckPATScope(ctx context.Context, patScopeRelations []relation.Relation, patScopeIdx []int) (map[int]bool, error) {
	scopeResults, err := s.relationService.BatchCheckPermission(ctx, patScopeRelations)
	if err != nil {
		return nil, err
	}
	denied := make(map[int]bool, len(scopeResults))
	for j, sr := range scopeResults {
		if !sr.Status {
			denied[patScopeIdx[j]] = true
		}
	}
	return denied, nil
}

// batchCheckWithScopeFilter runs user permission checks for scope-allowed items
// and returns merged results where scope-denied items are false.
func (s Service) batchCheckWithScopeFilter(ctx context.Context, relations []relation.Relation, scopeDenied map[int]bool) ([]relation.CheckPair, error) {
	var allowedRelations []relation.Relation
	var allowedIdx []int
	for i, rel := range relations {
		if !scopeDenied[i] {
			allowedRelations = append(allowedRelations, rel)
			allowedIdx = append(allowedIdx, i)
		}
	}

	results := make([]relation.CheckPair, len(relations))
	for i := range results {
		results[i] = relation.CheckPair{Relation: relations[i], Status: false}
	}
	if len(allowedRelations) > 0 {
		userResults, err := s.relationService.BatchCheckPermission(ctx, allowedRelations)
		if err != nil {
			return nil, err
		}
		for j, idx := range allowedIdx {
			results[idx] = userResults[j]
		}
	}
	return results, nil
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
