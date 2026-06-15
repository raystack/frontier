package membership

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/authenticate"
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

type UserPATService interface {
	GetByID(ctx context.Context, id string) (patmodels.PAT, error)
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

// AddOrganizationMember adds a principal (user, service user, or PAT) to an organization
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
	existing = excludePATAllProjects(existing, schema.OrganizationNamespace)
	if len(existing) > 0 {
		return ErrAlreadyMember
	}

	createdPolicy, err := s.createPolicy(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, roleID)
	if err != nil {
		return err
	}

	// PATs don't get relations.
	if principalType == schema.PATPrincipal {
		s.auditOrgMemberAdded(ctx, org, principal, roleID)
		return nil
	}

	relationName := orgRoleToRelation(fetchedRole)
	if relationName != "" {
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
// Supports user, service user, and PAT principals.
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
	// drop the PAT's all-projects policy — only the org role should be replaced here.
	existing = excludePATAllProjects(existing, schema.OrganizationNamespace)
	if len(existing) == 0 && principalType != schema.PATPrincipal {
		return ErrNotMember
	}

	// skip if the user already has exactly this role
	if len(existing) == 1 && existing[0].RoleID == resolvedRoleID {
		return nil
	}

	// only human users can be the last owner — skip for service users and PATs.
	var ownerRoleID string
	if principalType == schema.UserPrincipal {
		ownerRoleID, err = s.validateMinOwnerConstraint(ctx, orgID, resolvedRoleID, existing)
		if err != nil {
			return err
		}
	}

	if err := s.replacePolicy(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, resolvedRoleID, existing, ownerRoleID); err != nil {
		return err
	}

	// PATs don't get relations.
	if principalType == schema.PATPrincipal {
		s.auditOrgMemberRoleChanged(ctx, org, principal, resolvedRoleID)
		return nil
	}

	newRelation := orgRoleToRelation(fetchedRole)
	oldRelations := []string{schema.OwnerRelationName}
	err = s.replaceOrRemoveRelation(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, oldRelations, newRelation)
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

// SetPATAllProjectsRole grants a PAT a project-scoped role across all projects
// in the org via the pat_granted relation. Idempotent — replaces any existing
// all-projects role for this PAT on this org.
func (s *Service) SetPATAllProjectsRole(ctx context.Context, orgID, patID, roleID string) error {
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	principal, err := s.validatePrincipal(ctx, orgID, patID, schema.PATPrincipal)
	if err != nil {
		return err
	}

	fetchedRole, err := s.validateProjectRole(ctx, roleID, orgID)
	if err != nil {
		return err
	}
	resolvedRoleID := fetchedRole.ID

	allPolicies, err := s.policyService.List(ctx, policy.Filter{
		OrgID:         orgID,
		PrincipalID:   patID,
		PrincipalType: schema.PATPrincipal,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}

	var existing []policy.Policy
	for _, p := range allPolicies {
		if p.GrantRelation == schema.PATGrantRelationName {
			existing = append(existing, p)
		}
	}

	if len(existing) == 1 && existing[0].RoleID == resolvedRoleID {
		return nil
	}

	for _, p := range existing {
		if err := s.policyService.Delete(ctx, p.ID); err != nil {
			return fmt.Errorf("delete policy %s: %w", p.ID, err)
		}
	}

	if _, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        resolvedRoleID,
		ResourceID:    orgID,
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   patID,
		PrincipalType: schema.PATPrincipal,
		GrantRelation: schema.PATGrantRelationName,
	}); err != nil {
		s.log.ErrorContext(ctx, "membership state inconsistent: old pat_granted policies deleted but new policy creation failed, needs manual fix",
			"org_id", orgID,
			"pat_id", patID,
			"role_id", resolvedRoleID,
			"error", err,
		)
		return fmt.Errorf("create policy: %w", err)
	}

	s.auditOrgMemberRoleChanged(ctx, org, principal, resolvedRoleID)
	return nil
}

// ListPoliciesByPrincipal returns every policy held by the principal.
func (s *Service) ListPoliciesByPrincipal(ctx context.Context, principalID, principalType string) ([]policy.Policy, error) {
	return s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
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

// RemoveOrganizationMember removes a principal from an organization and cascades
// the removal through all org projects and groups, cleaning up both policies and
// relations. Returns ErrNotMember if the principal has no policies on this org,
// and ErrLastOwnerRole when removing the org's last owner.
func (s *Service) RemoveOrganizationMember(ctx context.Context, orgID, principalID, principalType string) error {
	return s.removeOrganizationMember(ctx, orgID, principalID, principalType, true)
}

// ForceRemoveOrganizationMember is RemoveOrganizationMember without the
// member-and-owner guards: it does not fail when the principal has no
// org-level policies and it removes the org's last owner without complaint.
// It exists for deletion cascades (see core/deleter), where the principal is
// leaving the system entirely and the org may legitimately be left ownerless.
// Anything user-initiated should go through RemoveOrganizationMember instead.
func (s *Service) ForceRemoveOrganizationMember(ctx context.Context, orgID, principalID, principalType string) error {
	return s.removeOrganizationMember(ctx, orgID, principalID, principalType, false)
}

func (s *Service) removeOrganizationMember(ctx context.Context, orgID, principalID, principalType string, guarded bool) error {
	targetAuditType, err := principalTypeToAuditType(principalType)
	if err != nil {
		return err
	}

	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		// deletion cascades must keep working on disabled orgs — the deleter
		// visits them deliberately so user deletion doesn't leave orphan
		// policies behind. Get discards the org payload on ErrDisabled, so
		// reconstruct the minimal org the cascade needs.
		if guarded || !errors.Is(err, organization.ErrDisabled) {
			return err
		}
		org = organization.Organization{ID: orgID}
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

	// only humans can be the last owner — skip for service users and PATs.
	// an empty ownerRoleID makes the cascade delete owner policies unguarded.
	var ownerRoleID string
	if guarded {
		if len(orgPolicies) == 0 {
			return ErrNotMember
		}
		if principalType == schema.UserPrincipal {
			ownerRoleID, err = s.validateMinOwnerConstraint(ctx, orgID, "", orgPolicies)
			if err != nil {
				return err
			}
		}
	}

	if err := s.cascadeRemovePrincipal(ctx, org, principalID, principalType, ownerRoleID); err != nil {
		return err
	}

	s.auditOrgMemberRemoved(ctx, org, principalID, targetAuditType)
	audit.GetAuditor(ctx, org.ID).Log(audit.OrgMemberDeletedEvent, audit.Target{
		ID:   principalID,
		Type: principalType,
	})

	return nil
}

// cascadeRemovePrincipal deletes all policies and SpiceDB relations for a principal
// being removed from an organization, including cascaded project/group sub-resources.
// Owner-role org policies are deleted with the atomic guard first; if the guard rejects
// (last owner), the method returns ErrLastOwnerRole before any other mutation.
func (s *Service) cascadeRemovePrincipal(ctx context.Context, org organization.Organization, principalID, principalType, ownerRoleID string) error {
	orgID := org.ID

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

	allPolicies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list all principal policies: %w", err)
	}

	// classify policies by scope
	var orgPolicies, subResourcePolicies []policy.Policy
	for _, pol := range allPolicies {
		switch pol.ResourceType {
		case schema.OrganizationNamespace:
			if pol.ResourceID == orgID {
				orgPolicies = append(orgPolicies, pol)
			}
		case schema.ProjectNamespace:
			if _, ok := orgProjectIDSet[pol.ResourceID]; ok {
				subResourcePolicies = append(subResourcePolicies, pol)
			}
		case schema.GroupNamespace:
			if _, ok := orgGroupIDSet[pol.ResourceID]; ok {
				subResourcePolicies = append(subResourcePolicies, pol)
			}
		}
	}

	// guarded owner delete first — returns early if this is the last owner
	for _, pol := range orgPolicies {
		if err := s.deletePolicy(ctx, pol, ownerRoleID); err != nil {
			if errors.Is(err, policy.ErrLastRoleGuard) {
				return ErrLastOwnerRole
			}
			return fmt.Errorf("delete org policy %s: %w", pol.ID, err)
		}
	}

	// guard passed — delete sub-resource policies
	var errs error
	for _, pol := range subResourcePolicies {
		if err := s.policyService.Delete(ctx, pol.ID); err != nil {
			errs = errors.Join(errs, fmt.Errorf("delete sub-resource policy %s: %w", pol.ID, err))
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

	// clean up SpiceDB relations
	for _, g := range orgGroups {
		if err := s.removeRelations(ctx, g.ID, schema.GroupNamespace, principalID, principalType, groupRelationNames); err != nil {
			return fmt.Errorf("remove group %s relations: %w", g.ID, err)
		}
	}
	if err := s.removeRelations(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, orgRelationNames); err != nil {
		return fmt.Errorf("remove org relations: %w", err)
	}

	// remove identity link for service users
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

	return nil
}

// removeRelations deletes the given relations for a principal on a resource.
func (s *Service) removeRelations(ctx context.Context, resourceID, resourceType, principalID, principalType string, relationNames []string) error {
	obj := relation.Object{ID: resourceID, Namespace: resourceType}
	sub := relation.Subject{ID: principalID, Namespace: principalType}
	for _, name := range relationNames {
		err := s.relationService.Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: name})
		if err != nil && !errors.Is(err, relation.ErrNotExist) {
			return fmt.Errorf("delete relation %s: %w", name, err)
		}
	}
	return nil
}

var (
	orgRelationNames   = []string{schema.OwnerRelationName}
	groupRelationNames = []string{schema.OwnerRelationName, schema.MemberRelationName}
)

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

// replaceOrRemoveRelation deletes the given old relations and, if
// newRelationName is non-empty, creates the replacement. When
// newRelationName is "" the old relations are simply removed with no
// replacement — used when the target role has no corresponding SpiceDB
// relation (e.g. non-owner org roles after the member relation was removed).
func (s *Service) replaceOrRemoveRelation(ctx context.Context, resourceID, resourceType, principalID, principalType string, oldRelations []string, newRelationName string) error {
	obj := relation.Object{ID: resourceID, Namespace: resourceType}
	sub := relation.Subject{ID: principalID, Namespace: principalType}

	for _, name := range oldRelations {
		err := s.relationService.Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: name})
		if err != nil && !errors.Is(err, relation.ErrNotExist) {
			return fmt.Errorf("delete relation %s: %w", name, err)
		}
	}

	if newRelationName == "" {
		return nil
	}

	_, err := s.relationService.Create(ctx, relation.Relation{
		Object: obj, Subject: sub, RelationName: newRelationName,
	})
	if err != nil {
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
	case schema.PATPrincipal:
		if s.userPATService == nil {
			return principalInfo{}, ErrInvalidPrincipal
		}
		pat, err := s.userPATService.GetByID(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if pat.OrgID != orgID {
			return principalInfo{}, ErrPrincipalNotInOrg
		}
		if !pat.ExpiresAt.After(time.Now()) {
			return principalInfo{}, ErrPrincipalExpired
		}
		return principalInfo{
			ID:   pat.ID,
			Type: schema.PATPrincipal,
			Name: pat.Title,
		}, nil
	default:
		return principalInfo{}, ErrInvalidPrincipal
	}
}

// orgRoleToRelation maps an org role to the corresponding SpiceDB relation name.
// Owner role gets the "owner" relation; all other roles return "" because the
// org member relation has been removed from the schema — non-owner access is
// resolved entirely through rolebindings (granted->).
func orgRoleToRelation(r role.Role) string {
	if r.Name == schema.RoleOrganizationOwner {
		return schema.OwnerRelationName
	}
	return ""
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

	if err := s.replacePolicy(ctx, projectID, schema.ProjectNamespace, principalID, principalType, resolvedRoleID, existing, ""); err != nil {
		return err
	}

	s.auditProjectMember(ctx, pkgAuditRecord.ProjectMemberRoleChangedEvent, prj, principalID, principalType, map[string]any{"role_id": resolvedRoleID})
	return nil
}

// RemoveProjectMember removes a principal from a project by deleting all their project-level policies.
func (s *Service) RemoveProjectMember(ctx context.Context, projectID, principalID, principalType string) error {
	switch principalType {
	case schema.UserPrincipal, schema.ServiceUserPrincipal, schema.GroupPrincipal, schema.PATPrincipal:
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
	case schema.PATPrincipal:
		if s.userPATService == nil {
			return ErrInvalidPrincipal
		}
		pat, err := s.userPATService.GetByID(ctx, principalID)
		if err != nil {
			return err
		}
		if pat.OrgID != orgID {
			return ErrNotOrgMember
		}
		if !pat.ExpiresAt.After(time.Now()) {
			return ErrPrincipalExpired
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

// resourcePolicyFilter builds the policy filter that scopes a listing to the
// given resource. Returns ErrInvalidResourceType for unsupported namespaces.
func resourcePolicyFilter(resourceID, resourceType string, filter MemberFilter) (policy.Filter, error) {
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
		return policy.Filter{}, ErrInvalidResourceType
	}
	return flt, nil
}

// ListPrincipalsByResource returns the principals (users, service users, groups)
// that have at least one policy on the given resource, optionally filtered by
// principal type and/or role, and optionally enriched with the full list of
// roles each principal holds on the resource.
func (s *Service) ListPrincipalsByResource(ctx context.Context, resourceID, resourceType string, filter MemberFilter) ([]Member, error) {
	flt, err := resourcePolicyFilter(resourceID, resourceType, filter)
	if err != nil {
		return nil, err
	}

	policies, err := s.policyService.List(ctx, flt)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	policies = excludePATAllProjects(policies, resourceType)

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
	allPolicies = excludePATAllProjects(allPolicies, resourceType)

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

// ListPrincipalIDsByResource returns the IDs of principals of the given type
// that have at least one policy on the resource. It is a primitive-typed,
// ID-only variant of ListPrincipalsByResource: it skips role enrichment
// entirely (a single policy query) and exists for consumer packages that
// cannot import membership types without creating an import cycle
// (e.g. core/serviceuser, which this package itself imports).
func (s *Service) ListPrincipalIDsByResource(ctx context.Context, resourceID, resourceType, principalType string) ([]string, error) {
	flt, err := resourcePolicyFilter(resourceID, resourceType, MemberFilter{PrincipalType: principalType})
	if err != nil {
		return nil, err
	}

	policies, err := s.policyService.List(ctx, flt)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	policies = excludePATAllProjects(policies, resourceType)

	ids := make([]string, 0, len(policies))
	seen := make(map[string]struct{}, len(policies))
	for _, pol := range policies {
		if _, ok := seen[pol.PrincipalID]; ok {
			continue
		}
		seen[pol.PrincipalID] = struct{}{}
		ids = append(ids, pol.PrincipalID)
	}
	return ids, nil
}

// SetGroupMemberRole upserts the role assignment for a principal in a group:
// if the principal has no existing group policy, they are added with the
// requested role; otherwise their existing role is replaced with the
// requested role. New adds require the principal to be a member of the
// group's parent organization. Demoting the last owner returns
// ErrLastGroupOwnerRole.
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

	// add path: principal has no existing group policy
	if len(existing) == 0 {
		if err := s.validateOrgMembership(ctx, grp.OrganizationID, principalID, principalType); err != nil {
			return err
		}
		createdPolicy, err := s.createPolicy(ctx, groupID, schema.GroupNamespace, principalID, principalType, resolvedRoleID)
		if err != nil {
			return err
		}
		if err := s.createRelation(ctx, groupID, schema.GroupNamespace, principalID, principalType, groupRoleToRelation(fetchedRole)); err != nil {
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
		s.auditGroupMemberAdded(ctx, grp, principal, resolvedRoleID)
		return nil
	}

	// change path: skip if the principal already has exactly this role
	if len(existing) == 1 && existing[0].RoleID == resolvedRoleID {
		return nil
	}

	ownerRoleID, err := s.validateMinGroupOwnerConstraint(ctx, groupID, resolvedRoleID, existing)
	if err != nil {
		return err
	}

	if err := s.replacePolicy(ctx, groupID, schema.GroupNamespace, principalID, principalType, resolvedRoleID, existing, ownerRoleID); err != nil {
		// replacePolicy returns ErrLastOwnerRole for any namespace; surface the
		// group-specific variant for callers/error mappers.
		if errors.Is(err, ErrLastOwnerRole) {
			return ErrLastGroupOwnerRole
		}
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

// RemoveGroupMember removes a principal from a group, cleaning up both their
// group policies and the matching SpiceDB relations. Returns ErrNotMember if
// the principal has no policies on this group; ErrLastGroupOwnerRole if they
// are the sole remaining owner (enforced atomically via the policy guard).
func (s *Service) RemoveGroupMember(ctx context.Context, groupID, principalID, principalType string) error {
	grp, err := s.groupService.Get(ctx, groupID)
	if err != nil {
		return err
	}

	principal, err := s.validateGroupPrincipal(ctx, principalID, principalType)
	if err != nil {
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
	if len(existing) == 0 {
		return ErrNotMember
	}

	// Pass empty newRoleID — removal, not role change. The function still
	// returns the owner role ID for the atomic guard on the delete path.
	ownerRoleID, err := s.validateMinGroupOwnerConstraint(ctx, groupID, "", existing)
	if err != nil {
		return err
	}

	for _, p := range existing {
		if err := s.deletePolicy(ctx, p, ownerRoleID); err != nil {
			if errors.Is(err, policy.ErrLastRoleGuard) {
				return ErrLastGroupOwnerRole
			}
			return fmt.Errorf("delete policy %s: %w", p.ID, err)
		}
	}

	if err := s.removeRelations(ctx, groupID, schema.GroupNamespace, principalID, principalType, groupRelationNames); err != nil {
		s.log.ErrorContext(ctx, "membership state inconsistent: group policies removed but relation cleanup failed, needs manual fix",
			"group_id", groupID,
			"principal_id", principalID,
			"principal_type", principalType,
			"error", err,
		)
		return err
	}

	s.auditGroupMemberRemoved(ctx, grp, principal)
	return nil
}

// RemoveAllGroupMembers tears down membership for a group that is being
// destroyed: deletes every policy on the group and every owner/member
// relation per principal. No min-owner check — the group itself is going
// away, so the invariant doesn't apply. Errors are joined; partial failures
// are logged so a retry can complete the cleanup.
func (s *Service) RemoveAllGroupMembers(ctx context.Context, groupID string) error {
	policies, err := s.policyService.List(ctx, policy.Filter{GroupID: groupID})
	if err != nil {
		return fmt.Errorf("list group policies: %w", err)
	}

	// First pass: delete every policy. Track which principals had any
	// delete failure so we don't strip their SpiceDB relations while a
	// surviving policy still references them.
	principals := make(map[string]policy.Policy, len(policies))
	failed := make(map[string]struct{}, len(policies))
	var errs error
	for _, p := range policies {
		key := p.PrincipalType + "\x00" + p.PrincipalID
		principals[key] = p
		if delErr := s.policyService.Delete(ctx, p.ID); delErr != nil {
			failed[key] = struct{}{}
			errs = errors.Join(errs, fmt.Errorf("delete policy %s: %w", p.ID, delErr))
		}
	}

	// Second pass: clean up direct relations only for principals whose
	// policies were all deleted successfully. The rest get retried on the
	// next attempt once their lingering policies are removed.
	for key, p := range principals {
		if _, hadFailure := failed[key]; hadFailure {
			continue
		}
		if relErr := s.removeRelations(ctx, groupID, schema.GroupNamespace, p.PrincipalID, p.PrincipalType, groupRelationNames); relErr != nil {
			errs = errors.Join(errs, fmt.Errorf("remove relations for %s:%s: %w", p.PrincipalType, p.PrincipalID, relErr))
		}
	}

	if errs != nil {
		s.log.ErrorContext(ctx, "partial failure cleaning up group members during group deletion; retry may be required",
			"group_id", groupID,
			"error", errs,
		)
	}
	return errs
}

// OnGroupDeleted tears down all SpiceDB state created during the group's
// lifetime: per-member policies and owner/member relations, policies where
// the group itself is the principal on other resources (e.g. group granted
// a role on a project), and the two org<->group hierarchy relations. The
// group entity itself is left for the caller (group.Service.DeleteModel)
// to remove.
//
// Errors are joined; partial failures are logged so a retry can complete
// the cleanup.
func (s *Service) OnGroupDeleted(ctx context.Context, groupID string) error {
	grp, err := s.groupService.Get(ctx, groupID)
	if err != nil {
		return err
	}

	var errs error
	if err := s.RemoveAllGroupMembers(ctx, groupID); err != nil {
		errs = errors.Join(errs, fmt.Errorf("remove group members: %w", err))
	}
	if err := s.removeGroupAsPrincipalPolicies(ctx, groupID); err != nil {
		errs = errors.Join(errs, fmt.Errorf("remove group-as-principal policies: %w", err))
	}
	if err := s.unlinkGroupFromOrg(ctx, groupID, grp.OrganizationID); err != nil {
		errs = errors.Join(errs, fmt.Errorf("unlink group from org: %w", err))
	}
	return errs
}

// removeGroupAsPrincipalPolicies deletes every policy where the given group
// is the principal — e.g. policies created by `CreatePolicy(principal=group:X,
// resource=project:Y, role=viewer)` that grant the group access to other
// resources. policyService.Delete is expected to also remove the matching
// rolebinding relation in SpiceDB.
func (s *Service) removeGroupAsPrincipalPolicies(ctx context.Context, groupID string) error {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalType: schema.GroupPrincipal,
		PrincipalID:   groupID,
	})
	if err != nil {
		return fmt.Errorf("list group-as-principal policies: %w", err)
	}
	var errs error
	for _, p := range policies {
		if delErr := s.policyService.Delete(ctx, p.ID); delErr != nil {
			errs = errors.Join(errs, fmt.Errorf("delete policy %s: %w", p.ID, delErr))
		}
	}
	return errs
}

// OnGroupCreated wires up SpiceDB relations for a newly-created group:
// links the group to its parent organization (both directions) and adds the
// creator as owner via SetGroupMemberRole. If the owner add fails, hierarchy
// relations are best-effort rolled back to avoid an unowned, half-linked group.
func (s *Service) OnGroupCreated(ctx context.Context, groupID, orgID, creatorID, creatorType string) error {
	if err := s.linkGroupToOrg(ctx, groupID, orgID); err != nil {
		return err
	}
	if err := s.SetGroupMemberRole(ctx, groupID, creatorID, creatorType, schema.GroupOwnerRole); err != nil {
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
// linkGroupToOrg creates the group→org hierarchy relation in SpiceDB.
// This is the only relation needed: group.org is used by the group
// permission definitions (e.g. group.delete = org->group_delete + ...).
func (s *Service) linkGroupToOrg(ctx context.Context, groupID, orgID string) error {
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}); err != nil {
		return fmt.Errorf("link group to org: %w", err)
	}
	return nil
}

// unlinkGroupFromOrg removes the group→org hierarchy relation.
// Used as best-effort cleanup when group-create wiring fails partway.
func (s *Service) unlinkGroupFromOrg(ctx context.Context, groupID, orgID string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
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
// after the role change. Returns the resolved group owner role ID so the
// caller can hand it to replacePolicy as a min-role guard, closing the TOCTOU
// race between this pre-check and the policy delete.
func (s *Service) validateMinGroupOwnerConstraint(ctx context.Context, groupID, newRoleID string, existing []policy.Policy) (string, error) {
	ownerRole, err := s.roleService.Get(ctx, schema.GroupOwnerRole)
	if err != nil {
		return "", fmt.Errorf("get group owner role: %w", err)
	}

	if newRoleID == ownerRole.ID {
		return ownerRole.ID, nil
	}

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

	ownerPolicies, err := s.policyService.List(ctx, policy.Filter{
		GroupID: groupID,
		RoleID:  ownerRole.ID,
	})
	if err != nil {
		return "", fmt.Errorf("list group owner policies: %w", err)
	}
	if len(ownerPolicies) <= 1 {
		return "", ErrLastGroupOwnerRole
	}
	return ownerRole.ID, nil
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

func (s *Service) auditGroupMemberRemoved(ctx context.Context, grp group.Group, p principalInfo) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.GroupMemberRemovedEvent,
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

	audit.GetAuditor(ctx, grp.OrganizationID).LogWithAttrs(audit.GroupMemberRemovedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"group_id": grp.ID,
	})
}

// ResourceFilter narrows the results of ListResourcesByPrincipal.
type ResourceFilter struct {
	// OrgID restricts project/group results to one org. No-op for orgs.
	OrgID string

	// NonInherited suppresses org-inheritance expansion for projects (direct
	// + group-expanded only). No-op for orgs and groups.
	NonInherited bool
}

// ListOrgsByPrincipal lets the organization package consume this without
// importing membership — that direction would be a cycle since membership
// already imports organization.
func (s *Service) ListOrgsByPrincipal(ctx context.Context, principal authenticate.Principal) ([]string, error) {
	return s.ListResourcesByPrincipal(ctx, principal, schema.OrganizationNamespace, ResourceFilter{})
}

// ListGroupsByPrincipal Shim for the group package (group → membership would cycle). PATs scope
// orgs and projects, not groups, so a PAT sees exactly its user's groups — resolve the PAT.
func (s *Service) ListGroupsByPrincipal(ctx context.Context, principal authenticate.Principal, orgID string) ([]string, error) {
	subjectID, subjectType := principal.ResolveSubject()
	return s.listResourcesForPrincipal(ctx, subjectID, subjectType, schema.GroupNamespace, ResourceFilter{OrgID: orgID})
}

// ListProjectsByPrincipal Shim for the project package (project → membership
// would cycle). Delegates to ListResourcesByPrincipal so PAT scope is intersected.
func (s *Service) ListProjectsByPrincipal(ctx context.Context, principal authenticate.Principal, orgID string, nonInherited bool) ([]string, error) {
	return s.ListResourcesByPrincipal(ctx, principal, schema.ProjectNamespace, ResourceFilter{OrgID: orgID, NonInherited: nonInherited})
}

// ListResourcesByPrincipal returns the resource IDs of the given type on which
// the principal has at least one policy. Reads Postgres policies — no SpiceDB.
// With a PAT, runs the algorithm twice (user, then PAT-as-principal) and
// intersects, so the PAT can narrow but never widen the user's visibility.
func (s *Service) ListResourcesByPrincipal(ctx context.Context, principal authenticate.Principal, resourceType string, filter ResourceFilter) ([]string, error) {
	subjectID, subjectType := principal.ResolveSubject()
	subjectResourceIDs, err := s.listResourcesForPrincipal(ctx, subjectID, subjectType, resourceType, filter)
	if err != nil {
		return nil, err
	}
	if principal.PAT == nil {
		return subjectResourceIDs, nil
	}

	patResourceIDs, err := s.listResourcesForPrincipal(ctx, principal.PAT.ID, schema.PATPrincipal, resourceType, filter)
	if err != nil {
		return nil, err
	}
	return utils.Intersection(subjectResourceIDs, patResourceIDs), nil
}

// listResourcesForPrincipal is the per-principal core; no PAT awareness.
func (s *Service) listResourcesForPrincipal(ctx context.Context, principalID, principalType, resourceType string, filter ResourceFilter) ([]string, error) {
	switch resourceType {
	case schema.OrganizationNamespace:
		return s.listOrgsForPrincipal(ctx, principalID, principalType)
	case schema.GroupNamespace:
		return s.listGroupsForPrincipal(ctx, principalID, principalType, filter)
	case schema.ProjectNamespace:
		return s.listProjectsForPrincipal(ctx, principalID, principalType, filter)
	default:
		return nil, ErrInvalidResourceType
	}
}

// listOrgsForPrincipal returns every org the principal has a policy on.
// Any policy is enough — we don't look at what the role grants. (Project
// listing does check role permissions; orgs and groups don't.)
func (s *Service) listOrgsForPrincipal(ctx context.Context, principalID, principalType string) ([]string, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
		ResourceType:  schema.OrganizationNamespace,
	})
	if err != nil {
		return nil, fmt.Errorf("list org policies: %w", err)
	}
	ids := make([]string, 0, len(policies))
	for _, pol := range policies {
		ids = append(ids, pol.ResourceID)
	}
	return utils.Deduplicate(ids), nil
}

// listGroupsForPrincipal returns every group the principal has a policy on.
// Same rule as orgs — any policy is enough, role permissions aren't checked.
func (s *Service) listGroupsForPrincipal(ctx context.Context, principalID, principalType string, filter ResourceFilter) ([]string, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
		ResourceType:  schema.GroupNamespace,
	})
	if err != nil {
		return nil, fmt.Errorf("list group policies: %w", err)
	}
	ids := make([]string, 0, len(policies))
	for _, pol := range policies {
		ids = append(ids, pol.ResourceID)
	}
	ids = utils.Deduplicate(ids)

	if filter.OrgID != "" && len(ids) > 0 {
		ids, err = s.narrowGroupsByOrg(ctx, ids, filter.OrgID)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// narrowGroupsByOrg keeps only group IDs whose org_id matches the given org.
// Performed by re-issuing groupService.List({OrganizationID, GroupIDs: ids}).
func (s *Service) narrowGroupsByOrg(ctx context.Context, ids []string, orgID string) ([]string, error) {
	groups, err := s.groupService.List(ctx, group.Filter{
		OrganizationID: orgID,
		GroupIDs:       ids,
	})
	if err != nil {
		return nil, fmt.Errorf("narrow groups by org: %w", err)
	}
	out := make([]string, 0, len(groups))
	for _, g := range groups {
		out = append(out, g.ID)
	}
	return out, nil
}

// listProjectsForPrincipal unions three sources, dedups, then narrows by
// filter.OrgID if set:
//
//  1. Direct project policies — gated by schema.ProjectDirectVisibilityPerms.
//  2. Group-expanded projects — same gate as direct. Runs even with
//     NonInherited=true; a user can be a project member via group.
//  3. Org inheritance (skipped if NonInherited=true) — gated by
//     schema.OrganizationProjectInheritPerms so only org roles that grant
//     project visibility (Owner, Manager, etc.) expand. Batched via
//     project.Filter.OrgIDs to avoid N+1 across multi-org users.
func (s *Service) listProjectsForPrincipal(ctx context.Context, principalID, principalType string, filter ResourceFilter) ([]string, error) {
	directIDs, err := s.listDirectProjectIDs(ctx, principalID, principalType)
	if err != nil {
		return nil, err
	}

	groupExpandedIDs, err := s.listGroupExpandedProjectIDs(ctx, principalID, principalType)
	if err != nil {
		return nil, err
	}

	var inheritedIDs []string
	if !filter.NonInherited {
		inheritedIDs, err = s.listOrgInheritedProjectIDs(ctx, principalID, principalType)
		if err != nil {
			return nil, err
		}
	}

	all := make([]string, 0, len(directIDs)+len(groupExpandedIDs)+len(inheritedIDs))
	all = append(all, directIDs...)
	all = append(all, groupExpandedIDs...)
	all = append(all, inheritedIDs...)
	ids := utils.Deduplicate(all)

	if filter.OrgID != "" && len(ids) > 0 {
		ids, err = s.narrowProjectsByOrg(ctx, ids, filter.OrgID)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// listDirectProjectIDs returns projects the principal has a direct policy on,
// kept only if the role grants any of the permissions that imply project
// visibility.
func (s *Service) listDirectProjectIDs(ctx context.Context, principalID, principalType string) ([]string, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:     principalID,
		PrincipalType:   principalType,
		ResourceType:    schema.ProjectNamespace,
		RolePermissions: schema.ProjectDirectVisibilityPerms,
	})
	if err != nil {
		return nil, fmt.Errorf("list direct project policies: %w", err)
	}
	return policyResourceIDs(policies), nil
}

// listGroupExpandedProjectIDs walks: principal → groups → project policies on
// those groups → kept only if the role grants project visibility.
func (s *Service) listGroupExpandedProjectIDs(ctx context.Context, principalID, principalType string) ([]string, error) {
	// Use the per-principal helper (not ListResourcesByPrincipal) so the PAT
	// pass doesn't trigger another PAT recursion on itself.
	groupIDs, err := s.listResourcesForPrincipal(ctx, principalID, principalType, schema.GroupNamespace, ResourceFilter{NonInherited: true})
	if err != nil {
		return nil, fmt.Errorf("list principal groups for project expansion: %w", err)
	}
	if len(groupIDs) == 0 {
		return nil, nil
	}
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalType:   schema.GroupPrincipal,
		PrincipalIDs:    groupIDs,
		ResourceType:    schema.ProjectNamespace,
		RolePermissions: schema.ProjectDirectVisibilityPerms,
	})
	if err != nil {
		return nil, fmt.Errorf("list project policies for principal groups: %w", err)
	}
	return policyResourceIDs(policies), nil
}

// listOrgInheritedProjectIDs finds projects a principal can see by virtue of
// holding a strong-enough role on the project's org (e.g. Org Owner sees all
// projects in their org; Org Viewer doesn't). Steps:
//   - get the principal's policies on orgs, kept only if the role grants any
//     permission that implies org→all-projects inheritance
//   - fetch all projects in those orgs in a single batched query
func (s *Service) listOrgInheritedProjectIDs(ctx context.Context, principalID, principalType string) ([]string, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:     principalID,
		PrincipalType:   principalType,
		ResourceType:    schema.OrganizationNamespace,
		RolePermissions: schema.OrganizationProjectInheritPerms,
	})
	if err != nil {
		return nil, fmt.Errorf("list org policies for inheritance: %w", err)
	}
	inheritingOrgIDs := policyResourceIDs(policies)
	if len(inheritingOrgIDs) == 0 {
		return nil, nil
	}
	projects, err := s.projectService.List(ctx, project.Filter{OrgIDs: inheritingOrgIDs})
	if err != nil {
		return nil, fmt.Errorf("list inherited projects: %w", err)
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	return ids, nil
}

// policyResourceIDs plucks the deduped resource IDs from a policy slice.
func policyResourceIDs(policies []policy.Policy) []string {
	ids := make([]string, 0, len(policies))
	for _, pol := range policies {
		ids = append(ids, pol.ResourceID)
	}
	return utils.Deduplicate(ids)
}

// narrowProjectsByOrg keeps only IDs whose org_id matches orgID (single query).
func (s *Service) narrowProjectsByOrg(ctx context.Context, ids []string, orgID string) ([]string, error) {
	projects, err := s.projectService.List(ctx, project.Filter{
		OrgID:      orgID,
		ProjectIDs: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("narrow projects by org: %w", err)
	}
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, p.ID)
	}
	return out, nil
}
