package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgauditrecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/utils"
	"go.uber.org/zap"
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
	List(ctx context.Context, f project.Filter) ([]project.Project, error)
}

type OrgService interface {
	Get(ctx context.Context, idOrName string) (organization.Organization, error)
	GetRaw(ctx context.Context, idOrName string) (organization.Organization, error)
}

type PATService interface {
	GetByID(ctx context.Context, id string) (patmodels.PAT, error)
}

type PolicyService interface {
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repository            Repository
	relationService       RelationService
	authnService          AuthnService
	projectService        ProjectService
	orgService            OrgService
	patService            PATService
	policyService         PolicyService
	auditRecordRepository AuditRecordRepository
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord auditmodels.AuditRecord) (auditmodels.AuditRecord, error)
}

func NewService(repository Repository,
	relationService RelationService, authnService AuthnService,
	projectService ProjectService, orgService OrgService,
	patService PATService, policyService PolicyService,
	auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		repository:            repository,
		relationService:       relationService,
		authnService:          authnService,
		projectService:        projectService,
		orgService:            orgService,
		patService:            patService,
		policyService:         policyService,
		auditRecordRepository: auditRecordRepository,
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
	// PAT → resolve to underlying user
	if principalType == schema.PATPrincipal {
		sub, err := s.resolvePATUser(ctx, principalID)
		if err != nil {
			return Resource{}, fmt.Errorf("resolving PAT principal: %w", err)
		}
		principalID = sub.ID
		principalType = sub.Namespace
	}

	p, err := s.getParent(ctx, res)
	if err != nil {
		return Resource{}, err
	}

	newResource, err := s.repository.Create(ctx, Resource{
		ID:            res.ID,
		URN:           res.CreateURN(p.urnScope),
		Name:          res.Name,
		Title:         res.Title,
		ProjectID:     p.projectID,
		OrgID:         p.org.ID,
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

	if err = s.addParentToResource(ctx, p, newResource); err != nil {
		return Resource{}, err
	}
	if err = s.AddResourceOwner(ctx, newResource); err != nil {
		return Resource{}, err
	}

	s.createAuditRecord(ctx, pkgauditrecord.ResourceCreatedEvent, newResource, p)
	return newResource, nil
}

type parent struct {
	projectID string
	org       organization.Organization
	urnScope  string
	link      relation.Subject
	linkName  string
	audit     auditmodels.Resource
}

func (s Service) getParent(ctx context.Context, res Resource) (parent, error) {
	if res.ProjectID == "" && res.OrgID != "" {
		org, err := s.orgService.GetRaw(ctx, res.OrgID)
		if err != nil {
			return parent{}, fmt.Errorf("failed to get organization: %w", err)
		}
		return parent{
			org:      org,
			urnScope: org.ID,
			link:     relation.Subject{ID: org.ID, Namespace: schema.OrganizationNamespace},
			linkName: schema.OrganizationRelationName,
			audit:    auditmodels.Resource{ID: org.ID, Type: pkgauditrecord.OrganizationType, Name: org.Title},
		}, nil
	}

	proj, err := s.projectService.Get(ctx, res.ProjectID)
	if err != nil {
		return parent{}, fmt.Errorf("failed to get project: %w", err)
	}
	return projectParent(proj), nil
}

func (s Service) getParentToDelete(ctx context.Context, res Resource) (parent, error) {
	if res.ProjectID == "" {
		return s.getParent(ctx, res)
	}
	projects, err := s.projectService.List(ctx, project.Filter{ProjectIDs: []string{res.ProjectID}, IncludeDisabled: true})
	if err != nil {
		return parent{}, fmt.Errorf("failed to get project: %w", err)
	}
	if len(projects) == 0 {
		return parent{}, fmt.Errorf("failed to get project: %w", project.ErrNotExist)
	}
	return projectParent(projects[0]), nil
}

func projectParent(proj project.Project) parent {
	return parent{
		projectID: proj.ID,
		org:       proj.Organization,
		urnScope:  proj.Name,
		link:      relation.Subject{ID: proj.ID, Namespace: schema.ProjectNamespace},
		linkName:  schema.ProjectRelationName,
		audit:     auditmodels.Resource{ID: proj.ID, Type: pkgauditrecord.ProjectType, Name: proj.Title},
	}
}

func (s Service) createAuditRecord(ctx context.Context, event pkgauditrecord.Event, res Resource, p parent) {
	targetName := res.Title
	if targetName == "" {
		targetName = res.Name
	}
	if _, err := s.auditRecordRepository.Create(ctx, auditmodels.AuditRecord{
		Event:    event,
		Resource: p.audit,
		Target: &auditmodels.Target{
			ID:   res.ID,
			Type: pkgauditrecord.EntityType(res.NamespaceID),
			Name: targetName,
		},
		OrgID:      p.org.ID,
		OrgName:    p.org.Title,
		OccurredAt: time.Now(),
	}); err != nil {
		grpczap.Extract(ctx).Error("failed to create resource audit record",
			zap.String("resource_id", res.ID), zap.Error(err))
	}
}

func (s Service) List(ctx context.Context, flt Filter) ([]Resource, error) {
	return s.repository.List(ctx, flt)
}

func (s Service) Update(ctx context.Context, resource Resource) (Resource, error) {
	return s.repository.Update(ctx, resource)
}

func (s Service) addParentToResource(ctx context.Context, p parent, res Resource) error {
	_, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        res.ID,
			Namespace: res.NamespaceID,
		},
		Subject:      p.link,
		RelationName: p.linkName,
	})
	return err
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
				// if object is org, then fetch org by name. The lookup is
				// state blind on purpose: the permission check must run
				// before org state matters, so a disabled org resolves here
				// and the authorization layer decides the outcome
				org, err := s.orgService.GetRaw(ctx, obj.ID)
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
	res, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	p, err := s.getParentToDelete(ctx, res)
	if err != nil {
		return err
	}

	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id,
			Namespace: namespaceID,
		},
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}

	s.createAuditRecord(ctx, pkgauditrecord.ResourceDeletedEvent, res, p)
	return nil
}

// Purge removes the resource and its SpiceDB tuples for good. Only the project
// delete cascade uses it; the API uses Delete, which keeps the row.
// TODO(fix): remove once the project delete cascade soft-deletes resources
func (s Service) Purge(ctx context.Context, namespaceID, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id,
			Namespace: namespaceID,
		},
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return s.repository.Purge(ctx, id)
}

func (s Service) RemovePrincipalAccess(ctx context.Context, principalID, principalType, orgID string, projectIDs []string) error {
	var scopes []Filter
	if orgID != "" {
		scopes = append(scopes, Filter{OrgID: orgID})
	}
	for _, projectID := range projectIDs {
		scopes = append(scopes, Filter{ProjectID: projectID})
	}
	if len(scopes) == 0 {
		return nil
	}

	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list principal policies: %w", err)
	}
	if len(policies) == 0 {
		return nil
	}

	resourceIDs := make(map[string]struct{})
	for _, scope := range scopes {
		resources, err := s.repository.List(ctx, scope)
		if err != nil {
			return fmt.Errorf("list resources: %w", err)
		}
		for _, res := range resources {
			resourceIDs[res.ID] = struct{}{}
		}
	}

	var errs error
	for _, pol := range policies {
		if _, ok := resourceIDs[pol.ResourceID]; ok {
			if err := s.policyService.Delete(ctx, pol.ID); err != nil {
				errs = errors.Join(errs, fmt.Errorf("delete resource policy %s: %w", pol.ID, err))
			}
		}
	}
	return errs
}
