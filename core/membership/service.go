package membership

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type PolicyService interface {
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
	DeleteWithMinRoleGuard(ctx context.Context, id string, guardRoleID string) error
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type RoleService interface {
	Get(ctx context.Context, idOrName string) (role.Role, error)
	List(ctx context.Context, flt role.Filter) ([]role.Role, error)
}

type OrgService interface {
	Get(ctx context.Context, idOrName string) (organization.Organization, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

type ProjectService interface {
	Get(ctx context.Context, idOrName string) (project.Project, error)
	List(ctx context.Context, flt project.Filter) ([]project.Project, error)
}

type GroupService interface {
	Get(ctx context.Context, idOrName string) (group.Group, error)
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
}

type ServiceuserService interface {
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
}

type UserPATService interface {
	GetByID(ctx context.Context, id string) (patmodels.PAT, error)
}

type ResourceService interface {
	RemovePrincipalAccess(ctx context.Context, principalID, principalType string, projectIDs []string) error
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord auditrecord.AuditRecord) (auditrecord.AuditRecord, error)
}

type Service struct {
	log                   *slog.Logger
	policyService         PolicyService
	relationService       RelationService
	roleService           RoleService
	orgService            OrgService
	userService           UserService
	projectService        ProjectService
	groupService          GroupService
	serviceuserService    ServiceuserService
	userPATService        UserPATService
	resourceService       ResourceService
	auditRecordRepository AuditRecordRepository
}

func NewService(
	logger *slog.Logger,
	policyService PolicyService,
	relationService RelationService,
	roleService RoleService,
	orgService OrgService,
	userService UserService,
	projectService ProjectService,
	groupService GroupService,
	serviceuserService ServiceuserService,
	auditRecordRepository AuditRecordRepository,
) *Service {
	return &Service{
		log:                   logger,
		policyService:         policyService,
		relationService:       relationService,
		roleService:           roleService,
		orgService:            orgService,
		userService:           userService,
		projectService:        projectService,
		groupService:          groupService,
		serviceuserService:    serviceuserService,
		auditRecordRepository: auditRecordRepository,
	}
}

// SetUserPATService sets the PAT dependency after construction to break the
// circular init order between userpat and membership services.
func (s *Service) SetUserPATService(ups UserPATService) {
	s.userPATService = ups
}

// SetResourceService sets the resource dependency after construction. The
// resource service is built after membership because it reaches membership
// through the PAT service.
func (s *Service) SetResourceService(rs ResourceService) {
	s.resourceService = rs
}

// RemoveAllPATPolicies deletes every policy held by a PAT.
func (s *Service) RemoveAllPATPolicies(ctx context.Context, patID string) error {
	_, err := s.removePoliciesByFilter(ctx, policy.Filter{
		PrincipalID:   patID,
		PrincipalType: schema.PATPrincipal,
	})
	return err
}

// removePoliciesByFilter lists policies matching the filter and deletes them.
func (s *Service) removePoliciesByFilter(ctx context.Context, filter policy.Filter) (int, error) {
	policies, err := s.policyService.List(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("list policies: %w", err)
	}
	for _, p := range policies {
		if err := s.policyService.Delete(ctx, p.ID); err != nil {
			return 0, fmt.Errorf("delete policy %s: %w", p.ID, err)
		}
	}
	return len(policies), nil
}

// replacePolicy deletes the given existing policies and creates a new one with the new role.
// When ownerRoleID is non-empty, owner-role policies are deleted atomically via
// DeleteWithMinRoleGuard to prevent the last-owner TOCTOU race.
func (s *Service) replacePolicy(ctx context.Context, resourceID, resourceType, principalID, principalType, roleID string, existing []policy.Policy, ownerRoleID string) error {
	for _, p := range existing {
		if err := s.deletePolicy(ctx, p, ownerRoleID); err != nil {
			if errors.Is(err, policy.ErrLastRoleGuard) {
				return ErrLastOwnerRole
			}
			return fmt.Errorf("delete policy %s: %w", p.ID, err)
		}
	}

	_, err := s.createPolicy(ctx, resourceID, resourceType, principalID, principalType, roleID)
	if err != nil {
		s.log.ErrorContext(ctx, "membership state inconsistent: old policies deleted but new policy creation failed, needs manual fix",
			"resource_id", resourceID,
			"resource_type", resourceType,
			"principal_id", principalID,
			"principal_type", principalType,
			"role_id", roleID,
			"error", err,
		)
		return err
	}
	return nil
}

func (s *Service) deletePolicy(ctx context.Context, pol policy.Policy, ownerRoleID string) error {
	if ownerRoleID != "" && pol.RoleID == ownerRoleID {
		return s.policyService.DeleteWithMinRoleGuard(ctx, pol.ID, ownerRoleID)
	}
	return s.policyService.Delete(ctx, pol.ID)
}

func (s *Service) createPolicy(ctx context.Context, resourceID, resourceType, principalID, principalType, roleID string) (policy.Policy, error) {
	created, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        roleID,
		ResourceID:    resourceID,
		ResourceType:  resourceType,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return policy.Policy{}, fmt.Errorf("create policy: %w", err)
	}
	return created, nil
}

func (s *Service) createRelation(ctx context.Context, resourceID, resourceType, principalID, principalType, relationName string) error {
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object:       relation.Object{ID: resourceID, Namespace: resourceType},
		Subject:      relation.Subject{ID: principalID, Namespace: principalType},
		RelationName: relationName,
	}); err != nil {
		return fmt.Errorf("create relation: %w", err)
	}
	return nil
}

// removeAllPolicies finds and deletes all policies for a principal on a resource.
// Returns the number of policies deleted.
func (s *Service) removeAllPolicies(ctx context.Context, resourceID, resourceType, principalID, principalType string) (int, error) {
	f := policyFilterForResource(resourceID, resourceType, principalID, principalType)
	existing, err := s.policyService.List(ctx, f)
	if err != nil {
		return 0, fmt.Errorf("list policies: %w", err)
	}
	for _, pol := range existing {
		if err := s.policyService.Delete(ctx, pol.ID); err != nil {
			return 0, fmt.Errorf("delete policy %s: %w", pol.ID, err)
		}
	}
	return len(existing), nil
}

// policyFilterForResource builds a policy.Filter with the correct resource-type field set.
func policyFilterForResource(resourceID, resourceType, principalID, principalType string) policy.Filter {
	f := policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
	}
	switch resourceType {
	case schema.OrganizationNamespace:
		f.OrgID = resourceID
	case schema.ProjectNamespace:
		f.ProjectID = resourceID
	case schema.GroupNamespace:
		f.GroupID = resourceID
	}
	return f
}

// excludePATAllProjects hides a PAT's all-projects grant from org member
// listings — that policy lives on the org but grants project access, not
// org membership.
func excludePATAllProjects(policies []policy.Policy, resourceType string) []policy.Policy {
	if resourceType != schema.OrganizationNamespace {
		return policies
	}
	filtered := policies[:0]
	for _, p := range policies {
		if p.GrantRelation != schema.PATGrantRelationName {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
