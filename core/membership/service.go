package membership

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

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
	"github.com/raystack/salt/log"
)

type PolicyService interface {
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type RoleService interface {
	Get(ctx context.Context, idOrName string) (role.Role, error)
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
	log                   log.Logger
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
	logger log.Logger,
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

// AddOrganizationMember adds a new user to an organization with an explicit role,
// bypassing the invitation flow. Only user principals are supported — this is a
// direct-add operation for superadmins.
// Returns ErrAlreadyMember if the user already has a policy on this org.
func (s *Service) AddOrganizationMember(ctx context.Context, orgID, principalID, principalType, roleID string) error {
	// orgService.Get returns ErrDisabled for disabled orgs
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	principal, err := s.validatePrincipal(ctx, principalID, principalType)
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
			s.log.Warn("orphaned policy: relation creation failed and policy cleanup also failed",
				"policy_id", createdPolicy.ID,
				"org_id", orgID,
				"principal_id", principalID,
				"policy_delete_error", deleteErr,
			)
		}
		return err
	}

	// audit logging
	s.auditOrgMemberAdded(ctx, org, principal, roleID)

	return nil
}

// SetOrganizationMemberRole changes an existing member's role in an organization.
// SetOrganizationMemberRole skips the write if the member already has exactly the requested role.
// Currently only user principals are supported. May be extended to service users
// in the future to give them org-level roles (see #1544).
func (s *Service) SetOrganizationMemberRole(ctx context.Context, orgID, principalID, principalType, roleID string) error {
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	principal, err := s.validatePrincipal(ctx, principalID, principalType)
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

	if err := s.validateMinOwnerConstraint(ctx, orgID, resolvedRoleID, existing); err != nil {
		return err
	}

	if err := s.replacePolicy(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, resolvedRoleID, existing); err != nil {
		return err
	}

	newRelation := orgRoleToRelation(fetchedRole)
	oldRelations := []string{schema.OwnerRelationName, schema.MemberRelationName}
	err = s.replaceRelation(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, oldRelations, newRelation)
	if err != nil {
		s.log.Error("membership state inconsistent: policy replaced but relation update failed, needs manual fix",
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

	if err = s.validateMinOwnerConstraint(ctx, orgID, "", orgPolicies); err != nil {
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
func (s *Service) validateMinOwnerConstraint(ctx context.Context, orgID, newRoleID string, existing []policy.Policy) error {
	ownerRole, err := s.roleService.Get(ctx, schema.RoleOrganizationOwner)
	if err != nil {
		return fmt.Errorf("get owner role: %w", err)
	}

	// no constraint if promoting to owner
	if newRoleID == ownerRole.ID {
		return nil
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
		return nil
	}

	// user is owner, being demoted — make sure at least one other owner remains
	ownerPolicies, err := s.policyService.List(ctx, policy.Filter{
		OrgID:  orgID,
		RoleID: ownerRole.ID,
	})
	if err != nil {
		return fmt.Errorf("list owner policies: %w", err)
	}
	if len(ownerPolicies) <= 1 {
		return ErrLastOwnerRole
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
		s.log.Error("membership state inconsistent: old policies deleted but new policy creation failed, needs manual fix",
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
		s.log.Error("membership state inconsistent: old relations deleted but new relation creation failed, needs manual fix",
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

// validatePrincipal checks that the principal exists and is active.
// To add support for a new principal type (e.g., service user), add a case here
// and add the corresponding service dependency to the Service struct.
func (s *Service) validatePrincipal(ctx context.Context, principalID, principalType string) (principalInfo, error) {
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
	// To support service users in the future, add:
	// case schema.ServiceUserPrincipal:
	//     su, err := s.serviceUserService.Get(ctx, principalID)
	//     ...
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
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberRoleChangedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:   p.ID,
			Type: pkgAuditRecord.UserType,
			Name: p.Name,
			Metadata: map[string]any{
				"email":   p.Email,
				"role_id": roleID,
			},
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
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:   p.ID,
			Type: pkgAuditRecord.UserType,
			Name: p.Name,
			Metadata: map[string]any{
				"email":   p.Email,
				"role_id": roleID,
			},
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

	s.auditProjectMemberRoleChanged(ctx, prj, principalID, principalType, resolvedRoleID)
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

	s.auditProjectMemberRemoved(ctx, prj, principalID, principalType)
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

func (s *Service) auditProjectMemberRoleChanged(ctx context.Context, prj project.Project, principalID, principalType, roleID string) {
	targetType, _ := principalTypeToAuditType(principalType)
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.ProjectMemberRoleChangedEvent,
		Resource: auditrecord.Resource{
			ID:   prj.ID,
			Type: pkgAuditRecord.ProjectType,
			Name: prj.Title,
		},
		Target: &auditrecord.Target{
			ID:   principalID,
			Type: targetType,
			Metadata: map[string]any{
				"principal_type": principalType,
				"role_id":        roleID,
			},
		},
		OrgID:      prj.Organization.ID,
		OccurredAt: time.Now(),
	})
}

func (s *Service) auditProjectMemberRemoved(ctx context.Context, prj project.Project, principalID, principalType string) {
	targetType, _ := principalTypeToAuditType(principalType)
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.ProjectMemberRemovedEvent,
		Resource: auditrecord.Resource{
			ID:   prj.ID,
			Type: pkgAuditRecord.ProjectType,
			Name: prj.Title,
		},
		Target: &auditrecord.Target{
			ID:   principalID,
			Type: targetType,
			Metadata: map[string]any{
				"principal_type": principalType,
			},
		},
		OrgID:      prj.Organization.ID,
		OccurredAt: time.Now(),
	})
}
