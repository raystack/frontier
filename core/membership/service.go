package membership

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"log/slog"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/utils"
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

// AddOrganizationMember adds a principal (user or service user) to an organization
// with an explicit role, bypassing the invitation flow.
// Returns ErrAlreadyMember if the principal already has a policy on this org.
func (s *Service) AddOrganizationMember(ctx context.Context, orgID, principalID, principalType, roleID string) error {
	// orgService.Get returns ErrDisabled for disabled orgs
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	principal, err := s.validatePrincipal(ctx, orgID, principalID, principalType)
	if err != nil {
		return err
	}

	fetchedRole, err := s.validateOrgRole(ctx, roleID, orgID)
	if err != nil {
		return err
	}

	// check if principal is already a member
	existing, err := s.policyService.List(ctx, policy.Filter{
		OrgID:         orgID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}
	if len(existing) > 0 {
		return ErrAlreadyMember
	}

	createdPolicy, err := s.createPolicy(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, roleID)
	if err != nil {
		return err
	}

	relationName := orgRoleToRelation(fetchedRole)
	if err := s.createRelation(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, relationName); err != nil {
		// best-effort cleanup to avoid orphaned policy
		if deleteErr := s.policyService.Delete(ctx, createdPolicy.ID); deleteErr != nil {
			s.log.WarnContext(ctx, "orphaned policy: relation creation failed and policy cleanup also failed",
				"policy_id", createdPolicy.ID,
				"org_id", orgID,
				"principal_id", principalID,
				"policy_delete_error", deleteErr,
			)
		}
		return err
	}

	// create identity link for service users (serviceuser#org@organization)
	// used by SpiceDB to resolve the manage permission: manage = org->serviceusermanage
	if principalType == schema.ServiceUserPrincipal {
		if err := s.createRelation(ctx, principalID, schema.ServiceUserPrincipal, orgID, schema.OrganizationNamespace, schema.OrganizationRelationName); err != nil {
			// best-effort cleanup of policy + org relation to avoid orphaned state
			if deleteErr := s.policyService.Delete(ctx, createdPolicy.ID); deleteErr != nil {
				s.log.WarnContext(ctx, "orphaned policy: identity link failed and policy cleanup also failed",
					"policy_id", createdPolicy.ID,
					"principal_id", principalID,
					"error", deleteErr,
				)
			}
			return fmt.Errorf("create serviceuser identity link: %w", err)
		}
	}

	// audit logging
	s.auditOrgMemberAdded(ctx, org, principal, roleID)

	return nil
}

// SetOrganizationMemberRole changes an existing member's role in an organization.
// Supports user and service user principals.
// Skips the write if the member already has exactly the requested role.
func (s *Service) SetOrganizationMemberRole(ctx context.Context, orgID, principalID, principalType, roleID string) error {
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	principal, err := s.validatePrincipal(ctx, orgID, principalID, principalType)
	if err != nil {
		return err
	}

	fetchedRole, err := s.validateOrgRole(ctx, roleID, orgID)
	if err != nil {
		return err
	}

	// use the canonical UUID from the fetched role for all comparisons and writes
	resolvedRoleID := fetchedRole.ID

	existing, err := s.policyService.List(ctx, policy.Filter{
		OrgID:         orgID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}
	if len(existing) == 0 {
		return ErrNotMember
	}

	// skip if the user already has exactly this role
	if len(existing) == 1 && existing[0].RoleID == resolvedRoleID {
		return nil
	}

	ownerRoleID, err := s.validateMinOwnerConstraint(ctx, orgID, resolvedRoleID, existing)
	if err != nil {
		return err
	}

	if err := s.replacePolicyWithOwnerGuard(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, resolvedRoleID, existing, ownerRoleID); err != nil {
		return err
	}

	newRelation := orgRoleToRelation(fetchedRole)
	oldRelations := []string{schema.OwnerRelationName, schema.MemberRelationName}
	err = s.replaceRelation(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, oldRelations, newRelation)
	if err != nil {
		s.log.ErrorContext(ctx, "membership state inconsistent: policy replaced but relation update failed, needs manual fix",
			"org_id", orgID,
			"principal_id", principalID,
			"principal_type", principalType,
			"new_role_id", resolvedRoleID,
			"expected_relation", newRelation,
			"error", err,
		)
		return err
	}

	s.auditOrgMemberRoleChanged(ctx, org, principal, resolvedRoleID)
	return nil
}

// RemoveOrganizationMember removes a principal from an organization and cascades
// the removal through all org projects and groups, cleaning up both policies and
// relations. Returns ErrNotMember if the principal has no policies on this org.
func (s *Service) RemoveOrganizationMember(ctx context.Context, orgID, principalID, principalType string) error {
	targetAuditType, err := principalTypeToAuditType(principalType)
	if err != nil {
		return err
	}

	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	// check if principal is a member at org level
	orgPolicies, err := s.policyService.List(ctx, policy.Filter{
		OrgID:         orgID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}
	if len(orgPolicies) == 0 {
		return ErrNotMember
	}

	if _, err = s.validateMinOwnerConstraint(ctx, orgID, "", orgPolicies); err != nil {
		return err
	}

	// pre-compute org project and group ID sets for O(1) lookups
	orgProjects, err := s.projectService.List(ctx, project.Filter{OrgID: orgID})
	if err != nil {
		return fmt.Errorf("list org projects: %w", err)
	}
	orgProjectIDSet := make(map[string]struct{}, len(orgProjects))
	for _, p := range orgProjects {
		orgProjectIDSet[p.ID] = struct{}{}
	}

	orgGroups, err := s.groupService.List(ctx, group.Filter{OrganizationID: orgID})
	if err != nil {
		return fmt.Errorf("list org groups: %w", err)
	}
	orgGroupIDSet := make(map[string]struct{}, len(orgGroups))
	for _, g := range orgGroups {
		orgGroupIDSet[g.ID] = struct{}{}
	}

	// list all policies for the principal across all resources
	allPolicies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list all principal policies: %w", err)
	}

	// delete sub-resource policies first (projects, groups), then relations,
	// then org policies last — so a retry after partial failure won't hit ErrNotMember
	var orgPolicyIDs []string
	var errs error
	for _, pol := range allPolicies {
		switch pol.ResourceType {
		case schema.OrganizationNamespace:
			if pol.ResourceID == orgID {
				orgPolicyIDs = append(orgPolicyIDs, pol.ID)
			}
		case schema.ProjectNamespace:
			if _, ok := orgProjectIDSet[pol.ResourceID]; ok {
				if err := s.policyService.Delete(ctx, pol.ID); err != nil {
					errs = errors.Join(errs, fmt.Errorf("delete project policy %s: %w", pol.ID, err))
				}
			}
		case schema.GroupNamespace:
			if _, ok := orgGroupIDSet[pol.ResourceID]; ok {
				if err := s.policyService.Delete(ctx, pol.ID); err != nil {
					errs = errors.Join(errs, fmt.Errorf("delete group policy %s: %w", pol.ID, err))
				}
			}
		}
	}
	if errs != nil {
		s.log.Error("partial failure removing member: some policies could not be deleted, manual cleanup may be needed",
			"org_id", orgID,
			"principal_id", principalID,
			"principal_type", principalType,
			"error", errs,
		)
		return errs
	}

	// remove relations at group level
	for _, g := range orgGroups {
		if err := s.removeRelations(ctx, g.ID, schema.GroupNamespace, principalID, principalType); err != nil {
			s.log.Error("partial failure removing member: group relation cleanup failed, manual cleanup may be needed",
				"org_id", orgID,
				"group_id", g.ID,
				"principal_id", principalID,
				"principal_type", principalType,
				"error", err,
			)
			return fmt.Errorf("remove group %s relations: %w", g.ID, err)
		}
	}

	// remove relations at org level
	if err := s.removeRelations(ctx, orgID, schema.OrganizationNamespace, principalID, principalType); err != nil {
		s.log.Error("partial failure removing member: org relation cleanup failed, manual cleanup may be needed",
			"org_id", orgID,
			"principal_id", principalID,
			"principal_type", principalType,
			"error", err,
		)
		return fmt.Errorf("remove org relations: %w", err)
	}

	// remove identity link for service users (serviceuser#org@organization)
	if principalType == schema.ServiceUserPrincipal {
		err := s.relationService.Delete(ctx, relation.Relation{
			Object:       relation.Object{ID: principalID, Namespace: schema.ServiceUserPrincipal},
			Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
			RelationName: schema.OrganizationRelationName,
		})
		if err != nil && !errors.Is(err, relation.ErrNotExist) {
			return fmt.Errorf("remove serviceuser identity link: %w", err)
		}
	}

	// delete org-level policies last
	for _, policyID := range orgPolicyIDs {
		if err := s.policyService.Delete(ctx, policyID); err != nil {
			s.log.Error("partial failure removing member: org policy deletion failed, manual cleanup may be needed",
				"org_id", orgID,
				"policy_id", policyID,
				"principal_id", principalID,
				"principal_type", principalType,
				"error", err,
			)
			return fmt.Errorf("delete org policy %s: %w", policyID, err)
		}
	}

	s.auditOrgMemberRemoved(ctx, org, principalID, targetAuditType)
	audit.GetAuditor(ctx, org.ID).Log(audit.OrgMemberDeletedEvent, audit.Target{
		ID:   principalID,
		Type: principalType,
	})

	return nil
}

// removeRelations deletes owner and member relations for a principal on a resource.
func (s *Service) removeRelations(ctx context.Context, resourceID, resourceType, principalID, principalType string) error {
	obj := relation.Object{ID: resourceID, Namespace: resourceType}
	sub := relation.Subject{ID: principalID, Namespace: principalType}
	for _, name := range []string{schema.OwnerRelationName, schema.MemberRelationName} {
		err := s.relationService.Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: name})
		if err != nil && !errors.Is(err, relation.ErrNotExist) {
			return fmt.Errorf("delete relation %s: %w", name, err)
		}
	}
	return nil
}

// validateMinOwnerConstraint ensures the org always has at least one owner after a role change.
// Returns the resolved owner role ID for reuse by callers.
func (s *Service) validateMinOwnerConstraint(ctx context.Context, orgID, newRoleID string, existing []policy.Policy) (string, error) {
	ownerRole, err := s.roleService.Get(ctx, schema.RoleOrganizationOwner)
	if err != nil {
		return "", fmt.Errorf("get owner role: %w", err)
	}

	// no constraint if promoting to owner
	if newRoleID == ownerRole.ID {
		return ownerRole.ID, nil
	}

	// no constraint if user is not currently an owner
	isCurrentlyOwner := false
	for _, p := range existing {
		if p.RoleID == ownerRole.ID {
			isCurrentlyOwner = true
			break
		}
	}
	if !isCurrentlyOwner {
		return ownerRole.ID, nil
	}

	// user is owner, being demoted — make sure at least one other owner remains
	ownerPolicies, err := s.policyService.List(ctx, policy.Filter{
		OrgID:  orgID,
		RoleID: ownerRole.ID,
	})
	if err != nil {
		return "", fmt.Errorf("list owner policies: %w", err)
	}
	if len(ownerPolicies) <= 1 {
		return "", ErrLastOwnerRole
	}
	return ownerRole.ID, nil
}

// replacePolicyWithOwnerGuard deletes existing policies using an atomic SQL guard
// that prevents removing the last owner, then creates the new policy.
func (s *Service) replacePolicyWithOwnerGuard(ctx context.Context, resourceID, resourceType, principalID, principalType, roleID string, existing []policy.Policy, ownerRoleID string) error {
	for _, p := range existing {
		err := s.policyService.DeleteWithMinRoleGuard(ctx, p.ID, ownerRoleID)
		if err != nil {
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

// replacePolicy deletes the given existing policies and creates a new one with the new role.
func (s *Service) replacePolicy(ctx context.Context, resourceID, resourceType, principalID, principalType, roleID string, existing []policy.Policy) error {
	for _, p := range existing {
		err := s.policyService.Delete(ctx, p.ID)
		if err != nil {
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

// replaceRelation deletes the given old relations for the principal on the resource,
// then creates a new relation with the given name.
// Only relation.ErrNotExist is ignored on delete — any other error is returned.
func (s *Service) replaceRelation(ctx context.Context, resourceID, resourceType, principalID, principalType string, oldRelations []string, newRelationName string) error {
	obj := relation.Object{ID: resourceID, Namespace: resourceType}
	sub := relation.Subject{ID: principalID, Namespace: principalType}

	for _, name := range oldRelations {
		err := s.relationService.Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: name})
		if err != nil && !errors.Is(err, relation.ErrNotExist) {
			return fmt.Errorf("delete relation %s: %w", name, err)
		}
	}

	_, err := s.relationService.Create(ctx, relation.Relation{
		Object: obj, Subject: sub, RelationName: newRelationName,
	})
	if err != nil {
		s.log.ErrorContext(ctx, "membership state inconsistent: old relations deleted but new relation creation failed, needs manual fix",
			"resource_id", resourceID,
			"resource_type", resourceType,
			"principal_id", principalID,
			"principal_type", principalType,
			"expected_relation", newRelationName,
			"error", err,
		)
		return fmt.Errorf("create relation: %w", err)
	}
	return nil
}

// validateOrgRole checks that the role is valid for organization scope and returns it.
// A role is valid if it is either:
// - a platform-wide role scoped to organizations, or
// - a custom role created for this specific organization.
func (s *Service) validateOrgRole(ctx context.Context, roleID, orgID string) (role.Role, error) {
	fetchedRole, err := s.roleService.Get(ctx, roleID)
	if err != nil {
		return role.Role{}, err
	}

	// role must be scoped to organization regardless of whether it's platform-wide or org-specific
	if !slices.Contains(fetchedRole.Scopes, schema.OrganizationNamespace) {
		return role.Role{}, ErrInvalidOrgRole
	}

	// custom role belonging to this org
	if fetchedRole.OrgID == orgID {
		return fetchedRole, nil
	}

	// platform-wide role (no org ownership)
	if utils.IsNullUUID(fetchedRole.OrgID) {
		return fetchedRole, nil
	}

	return role.Role{}, ErrInvalidOrgRole
}

// principalInfo holds validated principal details for audit and downstream use.
type principalInfo struct {
	ID    string
	Type  string
	Name  string
	Email string
}

// validatePrincipal checks that the principal exists, is active, and belongs to
// the target org. For users, org membership is checked separately via policies.
// For service users, org ownership is validated here since they have a fixed OrgID.
func (s *Service) validatePrincipal(ctx context.Context, orgID, principalID, principalType string) (principalInfo, error) {
	switch principalType {
	case schema.UserPrincipal:
		usr, err := s.userService.GetByID(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if usr.State == user.Disabled {
			return principalInfo{}, user.ErrDisabled
		}
		return principalInfo{
			ID:    usr.ID,
			Type:  schema.UserPrincipal,
			Name:  usr.Title,
			Email: usr.Email,
		}, nil
	case schema.ServiceUserPrincipal:
		su, err := s.serviceuserService.Get(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if su.OrgID != orgID {
			return principalInfo{}, ErrPrincipalNotInOrg
		}
		if su.State == string(serviceuser.Disabled) {
			return principalInfo{}, serviceuser.ErrDisabled
		}
		return principalInfo{
			ID:   su.ID,
			Type: schema.ServiceUserPrincipal,
			Name: su.Title,
		}, nil
	default:
		return principalInfo{}, ErrInvalidPrincipal
	}
}

// orgRoleToRelation maps an org role to the corresponding SpiceDB relation name.
// Owner role gets "owner" relation, everything else gets "member" relation.
func orgRoleToRelation(r role.Role) string {
	if r.Name == schema.RoleOrganizationOwner {
		return schema.OwnerRelationName
	}
	return schema.MemberRelationName
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

func (s *Service) auditOrgMemberRoleChanged(ctx context.Context, org organization.Organization, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberRoleChangedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})

	audit.GetAuditor(ctx, org.ID).LogWithAttrs(audit.OrgMemberRoleChangedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id": roleID,
	})
}

func (s *Service) auditOrgMemberAdded(ctx context.Context, org organization.Organization, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})

	audit.GetAuditor(ctx, org.ID).LogWithAttrs(audit.OrgMemberCreatedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id": roleID,
	})
}

func (s *Service) auditOrgMemberRemoved(ctx context.Context, org organization.Organization, targetID string, targetType pkgAuditRecord.EntityType) {
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberRemovedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:   targetID,
			Type: targetType,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})
}

func principalTypeToAuditType(principalType string) (pkgAuditRecord.EntityType, error) {
	switch principalType {
	case schema.ServiceUserPrincipal:
		return pkgAuditRecord.ServiceUserType, nil
	case schema.UserPrincipal:
		return pkgAuditRecord.UserType, nil
	case schema.GroupPrincipal:
		return pkgAuditRecord.GroupType, nil
	case schema.PATPrincipal:
		return pkgAuditRecord.PATType, nil
	default:
		return "", ErrInvalidPrincipalType
	}
}

// SetProjectMemberRole sets or changes a principal's role in a project (upsert).
// It validates the role is project-scoped and the principal is a member of the parent org.
// No explicit SpiceDB relations are managed — projects use policies only.
func (s *Service) SetProjectMemberRole(ctx context.Context, projectID, principalID, principalType, roleID string) error {
	prj, err := s.projectService.Get(ctx, projectID)
	if err != nil {
		return err
	}

	fetchedRole, err := s.validateProjectRole(ctx, roleID, prj.Organization.ID)
	if err != nil {
		return err
	}
	resolvedRoleID := fetchedRole.ID

	if err := s.validateOrgMembership(ctx, prj.Organization.ID, principalID, principalType); err != nil {
		return err
	}

	existing, err := s.policyService.List(ctx, policy.Filter{
		ProjectID:     projectID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}

	// skip if the principal already has exactly this role
	if len(existing) == 1 && existing[0].RoleID == resolvedRoleID {
		return nil
	}

	if err := s.replacePolicy(ctx, projectID, schema.ProjectNamespace, principalID, principalType, resolvedRoleID, existing); err != nil {
		return err
	}

	s.auditProjectMember(ctx, pkgAuditRecord.ProjectMemberRoleChangedEvent, prj, principalID, principalType, map[string]any{"role_id": resolvedRoleID})
	return nil
}

// RemoveProjectMember removes a principal from a project by deleting all their project-level policies.
func (s *Service) RemoveProjectMember(ctx context.Context, projectID, principalID, principalType string) error {
	switch principalType {
	case schema.UserPrincipal, schema.ServiceUserPrincipal, schema.GroupPrincipal:
	default:
		return ErrInvalidPrincipalType
	}

	prj, err := s.projectService.Get(ctx, projectID)
	if err != nil {
		return err
	}

	removed, err := s.removeAllPolicies(ctx, projectID, schema.ProjectNamespace, principalID, principalType)
	if err != nil {
		return err
	}
	if removed == 0 {
		return ErrNotMember
	}

	s.auditProjectMember(ctx, pkgAuditRecord.ProjectMemberRemovedEvent, prj, principalID, principalType, nil)
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

// validateProjectRole checks that the role is valid for project scope:
// - a platform-wide role scoped to projects, or
// - a custom role created for the project's parent organization.
func (s *Service) validateProjectRole(ctx context.Context, roleID, orgID string) (role.Role, error) {
	fetchedRole, err := s.roleService.Get(ctx, roleID)
	if err != nil {
		return role.Role{}, err
	}
	if !slices.Contains(fetchedRole.Scopes, schema.ProjectNamespace) {
		return role.Role{}, ErrInvalidProjectRole
	}

	// custom role belonging to the project's parent org
	if fetchedRole.OrgID == orgID {
		return fetchedRole, nil
	}

	// platform-wide role (no org ownership)
	if utils.IsNullUUID(fetchedRole.OrgID) {
		return fetchedRole, nil
	}

	return role.Role{}, ErrInvalidProjectRole
}

// validateOrgMembership checks that the principal exists and belongs to the given org.
// For users, org membership is verified via org-level policies.
// For service users and groups, org membership is verified via their org ID field.
func (s *Service) validateOrgMembership(ctx context.Context, orgID, principalID, principalType string) error {
	switch principalType {
	case schema.UserPrincipal:
		usr, err := s.userService.GetByID(ctx, principalID)
		if err != nil {
			return err
		}
		if usr.State == user.Disabled {
			return user.ErrDisabled
		}
		orgPolicies, err := s.policyService.List(ctx, policy.Filter{
			OrgID:         orgID,
			PrincipalID:   principalID,
			PrincipalType: principalType,
		})
		if err != nil {
			return err
		}
		if len(orgPolicies) == 0 {
			return ErrNotOrgMember
		}
	case schema.ServiceUserPrincipal:
		su, err := s.serviceuserService.Get(ctx, principalID)
		if err != nil {
			return err
		}
		if su.OrgID != orgID {
			return ErrNotOrgMember
		}
	case schema.GroupPrincipal:
		grp, err := s.groupService.Get(ctx, principalID)
		if err != nil {
			return err
		}
		if grp.OrganizationID != orgID {
			return ErrNotOrgMember
		}
	default:
		return ErrInvalidPrincipalType
	}
	return nil
}

func (s *Service) auditProjectMember(ctx context.Context, event pkgAuditRecord.Event, prj project.Project, principalID, principalType string, meta map[string]any) {
	targetType, _ := principalTypeToAuditType(principalType)
	if meta == nil {
		meta = map[string]any{}
	}
	meta["principal_type"] = principalType
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: event,
		Resource: auditrecord.Resource{
			ID:   prj.ID,
			Type: pkgAuditRecord.ProjectType,
			Name: prj.Title,
		},
		Target: &auditrecord.Target{
			ID:       principalID,
			Type:     targetType,
			Metadata: meta,
		},
		OrgID:      prj.Organization.ID,
		OccurredAt: time.Now(),
	})
}

// MemberFilter narrows the results of ListPrincipalsByResource.
type MemberFilter struct {
	// PrincipalType restricts the result to a single principal type
	// (e.g. schema.UserPrincipal, schema.ServiceUserPrincipal, schema.GroupPrincipal).
	// Empty means no restriction.
	PrincipalType string
	// RoleIDs includes principals that have at least one of these roles on the resource.
	// Empty means no role filtering.
	RoleIDs []string
}

// Member is a principal that has one or more policies on a resource.
type Member struct {
	PrincipalID   string
	PrincipalType string
	Roles         []role.Role
}

// ListPrincipalsByResource returns the principals (users, service users, groups)
// that have at least one policy on the given resource, optionally filtered by
// principal type and/or role, and optionally enriched with the full list of
// roles each principal holds on the resource.
func (s *Service) ListPrincipalsByResource(ctx context.Context, resourceID, resourceType string, filter MemberFilter) ([]Member, error) {
	flt := policy.Filter{
		PrincipalType: filter.PrincipalType,
		RoleIDs:       filter.RoleIDs,
		ResourceType:  resourceType,
	}
	switch resourceType {
	case schema.OrganizationNamespace:
		flt.OrgID = resourceID
	case schema.ProjectNamespace:
		flt.ProjectID = resourceID
	case schema.GroupNamespace:
		flt.GroupID = resourceID
	default:
		return nil, ErrInvalidResourceType
	}

	policies, err := s.policyService.List(ctx, flt)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	// deduplicate by (principalID, principalType) preserving order
	memberIndex := make(map[string]int, len(policies))
	members := make([]Member, 0, len(policies))
	for _, pol := range policies {
		key := pol.PrincipalType + "\x00" + pol.PrincipalID
		if _, ok := memberIndex[key]; ok {
			continue
		}
		memberIndex[key] = len(members)
		members = append(members, Member{
			PrincipalID:   pol.PrincipalID,
			PrincipalType: pol.PrincipalType,
		})
	}

	// fetch all policies for the resource (without role filtering) to get
	// the complete set of roles per principal in a single query
	roleFlt := flt
	roleFlt.RoleIDs = nil
	allPolicies, err := s.policyService.List(ctx, roleFlt)
	if err != nil {
		return nil, fmt.Errorf("list policies for role enrichment: %w", err)
	}

	principalRoleIDs := make(map[string][]string, len(members))
	roleSeen := make(map[string]map[string]struct{}, len(members))
	uniqueRoleIDs := make(map[string]struct{})
	for _, pol := range allPolicies {
		if pol.RoleID == "" {
			continue
		}
		key := pol.PrincipalType + "\x00" + pol.PrincipalID
		if _, ok := memberIndex[key]; !ok {
			continue
		}
		if roleSeen[key] == nil {
			roleSeen[key] = make(map[string]struct{})
		}
		if _, ok := roleSeen[key][pol.RoleID]; ok {
			continue
		}
		roleSeen[key][pol.RoleID] = struct{}{}
		principalRoleIDs[key] = append(principalRoleIDs[key], pol.RoleID)
		uniqueRoleIDs[pol.RoleID] = struct{}{}
	}

	if len(uniqueRoleIDs) > 0 {
		ids := make([]string, 0, len(uniqueRoleIDs))
		for id := range uniqueRoleIDs {
			ids = append(ids, id)
		}
		roles, err := s.roleService.List(ctx, role.Filter{IDs: ids})
		if err != nil {
			return nil, fmt.Errorf("list roles: %w", err)
		}
		roleByID := make(map[string]role.Role, len(roles))
		for _, r := range roles {
			roleByID[r.ID] = r
		}
		for key, idx := range memberIndex {
			memberRoles := make([]role.Role, 0, len(principalRoleIDs[key]))
			for _, rid := range principalRoleIDs[key] {
				if r, ok := roleByID[rid]; ok {
					memberRoles = append(memberRoles, r)
				}
			}
			members[idx].Roles = memberRoles
		}
	}

	return members, nil
}

// AddGroupMember adds a principal as a member of a group with an explicit role.
// Returns ErrAlreadyMember if the principal already has a policy on this group.
// The principal must be a member of the group's parent organization.
func (s *Service) AddGroupMember(ctx context.Context, groupID, principalID, principalType, roleID string) error {
	grp, err := s.groupService.Get(ctx, groupID)
	if err != nil {
		return err
	}

	principal, err := s.validateGroupPrincipal(ctx, principalID, principalType)
	if err != nil {
		return err
	}

	fetchedRole, err := s.validateGroupRole(ctx, roleID, grp.OrganizationID)
	if err != nil {
		return err
	}

	if err := s.validateOrgMembership(ctx, grp.OrganizationID, principalID, principalType); err != nil {
		return err
	}

	existing, err := s.policyService.List(ctx, policy.Filter{
		GroupID:       groupID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}
	if len(existing) > 0 {
		return ErrAlreadyMember
	}

	createdPolicy, err := s.createPolicy(ctx, groupID, schema.GroupNamespace, principalID, principalType, fetchedRole.ID)
	if err != nil {
		return err
	}

	relationName := groupRoleToRelation(fetchedRole)
	if err := s.createRelation(ctx, groupID, schema.GroupNamespace, principalID, principalType, relationName); err != nil {
		if deleteErr := s.policyService.Delete(ctx, createdPolicy.ID); deleteErr != nil {
			s.log.WarnContext(ctx, "orphaned policy: relation creation failed and policy cleanup also failed",
				"policy_id", createdPolicy.ID,
				"group_id", groupID,
				"principal_id", principalID,
				"policy_delete_error", deleteErr,
			)
		}
		return err
	}

	s.auditGroupMemberAdded(ctx, grp, principal, fetchedRole.ID)
	return nil
}

// SetGroupMemberRole changes an existing member's role in a group.
// Returns ErrNotMember if the principal has no existing policy on the group.
// Enforces the min-owner constraint: demoting the last owner returns ErrLastGroupOwnerRole.
func (s *Service) SetGroupMemberRole(ctx context.Context, groupID, principalID, principalType, roleID string) error {
	grp, err := s.groupService.Get(ctx, groupID)
	if err != nil {
		return err
	}

	principal, err := s.validateGroupPrincipal(ctx, principalID, principalType)
	if err != nil {
		return err
	}

	fetchedRole, err := s.validateGroupRole(ctx, roleID, grp.OrganizationID)
	if err != nil {
		return err
	}
	resolvedRoleID := fetchedRole.ID

	existing, err := s.policyService.List(ctx, policy.Filter{
		GroupID:       groupID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}
	if len(existing) == 0 {
		return ErrNotMember
	}

	// skip if the user already has exactly this role
	if len(existing) == 1 && existing[0].RoleID == resolvedRoleID {
		return nil
	}

	if err := s.validateMinGroupOwnerConstraint(ctx, groupID, resolvedRoleID, existing); err != nil {
		return err
	}

	if err := s.replacePolicy(ctx, groupID, schema.GroupNamespace, principalID, principalType, resolvedRoleID, existing); err != nil {
		return err
	}

	newRelation := groupRoleToRelation(fetchedRole)
	oldRelations := []string{schema.OwnerRelationName, schema.MemberRelationName}
	if err := s.replaceRelation(ctx, groupID, schema.GroupNamespace, principalID, principalType, oldRelations, newRelation); err != nil {
		s.log.ErrorContext(ctx, "membership state inconsistent: policy replaced but group relation update failed, needs manual fix",
			"group_id", groupID,
			"principal_id", principalID,
			"principal_type", principalType,
			"new_role_id", resolvedRoleID,
			"expected_relation", newRelation,
			"error", err,
		)
		return err
	}

	s.auditGroupMemberRoleChanged(ctx, grp, principal, resolvedRoleID)
	return nil
}

// OnGroupCreated wires up SpiceDB relations for a newly-created group:
// links the group to its parent organization (both directions) and adds the
// creator as owner via AddGroupMember. If the owner add fails, hierarchy
// relations are best-effort rolled back to avoid an unowned, half-linked group.
func (s *Service) OnGroupCreated(ctx context.Context, groupID, orgID, creatorID, creatorType string) error {
	if err := s.linkGroupToOrg(ctx, groupID, orgID); err != nil {
		return err
	}
	if err := s.AddGroupMember(ctx, groupID, creatorID, creatorType, schema.GroupOwnerRole); err != nil {
		if cleanupErr := s.unlinkGroupFromOrg(ctx, groupID, orgID); cleanupErr != nil {
			s.log.WarnContext(ctx, "group hierarchy cleanup failed after owner add failure",
				"group_id", groupID,
				"org_id", orgID,
				"error", cleanupErr,
			)
		}
		return err
	}
	return nil
}

// linkGroupToOrg creates the two hierarchy relations between a group and its org:
//   - group#org@organization      (identity link from group to org)
//   - organization#member@group#member (lets org#member traverse to group members)
//
// If the second relation fails, the first is best-effort rolled back so we
// don't leave a one-way link.
func (s *Service) linkGroupToOrg(ctx context.Context, groupID, orgID string) error {
	groupOrg := relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}
	if _, err := s.relationService.Create(ctx, groupOrg); err != nil {
		return fmt.Errorf("link group to org: %w", err)
	}

	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Subject: relation.Subject{
			ID:              groupID,
			Namespace:       schema.GroupNamespace,
			SubRelationName: schema.MemberRelationName,
		},
		RelationName: schema.MemberRelationName,
	}); err != nil {
		if delErr := s.relationService.Delete(ctx, groupOrg); delErr != nil && !errors.Is(delErr, relation.ErrNotExist) {
			s.log.WarnContext(ctx, "group->org rollback failed after org member relation failure",
				"group_id", groupID,
				"org_id", orgID,
				"error", delErr,
			)
		}
		return fmt.Errorf("add group as org member: %w", err)
	}

	return nil
}

// unlinkGroupFromOrg removes both hierarchy relations between a group and its
// org. Used as best-effort cleanup when group-create wiring fails partway.
// relation.ErrNotExist is ignored; any other error is returned.
func (s *Service) unlinkGroupFromOrg(ctx context.Context, groupID, orgID string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}

	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Subject: relation.Subject{
			ID:              groupID,
			Namespace:       schema.GroupNamespace,
			SubRelationName: schema.MemberRelationName,
		},
		RelationName: schema.MemberRelationName,
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return nil
}

// validateGroupRole checks that the role is valid for group scope:
//   - a platform-wide role scoped to groups, or
//   - a custom role created for the group's parent organization.
func (s *Service) validateGroupRole(ctx context.Context, roleID, orgID string) (role.Role, error) {
	fetchedRole, err := s.roleService.Get(ctx, roleID)
	if err != nil {
		return role.Role{}, err
	}
	if !slices.Contains(fetchedRole.Scopes, schema.GroupNamespace) {
		return role.Role{}, ErrInvalidGroupRole
	}
	if fetchedRole.OrgID == orgID {
		return fetchedRole, nil
	}
	if utils.IsNullUUID(fetchedRole.OrgID) {
		return fetchedRole, nil
	}
	return role.Role{}, ErrInvalidGroupRole
}

// validateGroupPrincipal fetches and validates the principal for group operations.
// Currently only app/user is supported; the switch is structured so future principal
// types (e.g. serviceuser) can be enabled here without touching call sites.
func (s *Service) validateGroupPrincipal(ctx context.Context, principalID, principalType string) (principalInfo, error) {
	switch principalType {
	case schema.UserPrincipal:
		usr, err := s.userService.GetByID(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if usr.State == user.Disabled {
			return principalInfo{}, user.ErrDisabled
		}
		return principalInfo{
			ID:    usr.ID,
			Type:  schema.UserPrincipal,
			Name:  usr.Title,
			Email: usr.Email,
		}, nil
	default:
		return principalInfo{}, ErrInvalidPrincipalType
	}
}

// validateMinGroupOwnerConstraint ensures the group keeps at least one owner
// after the role change. Mirrors the org-level constraint.
func (s *Service) validateMinGroupOwnerConstraint(ctx context.Context, groupID, newRoleID string, existing []policy.Policy) error {
	ownerRole, err := s.roleService.Get(ctx, schema.GroupOwnerRole)
	if err != nil {
		return fmt.Errorf("get group owner role: %w", err)
	}

	if newRoleID == ownerRole.ID {
		return nil
	}

	isCurrentlyOwner := false
	for _, p := range existing {
		if p.RoleID == ownerRole.ID {
			isCurrentlyOwner = true
			break
		}
	}
	if !isCurrentlyOwner {
		return nil
	}

	ownerPolicies, err := s.policyService.List(ctx, policy.Filter{
		GroupID: groupID,
		RoleID:  ownerRole.ID,
	})
	if err != nil {
		return fmt.Errorf("list group owner policies: %w", err)
	}
	if len(ownerPolicies) <= 1 {
		return ErrLastGroupOwnerRole
	}
	return nil
}

// groupRoleToRelation maps a group role to the matching SpiceDB relation name.
func groupRoleToRelation(r role.Role) string {
	if r.Name == schema.GroupOwnerRole {
		return schema.OwnerRelationName
	}
	return schema.MemberRelationName
}

func (s *Service) auditGroupMemberAdded(ctx context.Context, grp group.Group, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.GroupMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   grp.ID,
			Type: pkgAuditRecord.GroupType,
			Name: grp.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      grp.OrganizationID,
		OccurredAt: time.Now(),
	})

	audit.GetAuditor(ctx, grp.OrganizationID).LogWithAttrs(audit.GroupMemberCreatedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id":  roleID,
		"group_id": grp.ID,
	})
}

func (s *Service) auditGroupMemberRoleChanged(ctx context.Context, grp group.Group, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.GroupMemberRoleChangedEvent,
		Resource: auditrecord.Resource{
			ID:   grp.ID,
			Type: pkgAuditRecord.GroupType,
			Name: grp.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      grp.OrganizationID,
		OccurredAt: time.Now(),
	})

	audit.GetAuditor(ctx, grp.OrganizationID).LogWithAttrs(audit.GroupMemberRoleChangedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id":  roleID,
		"group_id": grp.ID,
	})
}
