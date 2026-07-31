package membership

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

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

	if _, err := s.validateOrgRole(ctx, roleID, orgID); err != nil {
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
	if hasExactlyRole(existing, resolvedRoleID) {
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

	if hasExactlyRole(existing, resolvedRoleID) {
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
	if err := audit.GetAuditor(ctx, org.ID).Log(audit.OrgMemberDeletedEvent, audit.Target{
		ID:   principalID,
		Type: principalType,
	}); err != nil {
		s.log.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.OrgMemberDeletedEvent)
	}

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
	orgProjectIDs := make([]string, 0, len(orgProjects))
	orgProjectIDSet := make(map[string]struct{}, len(orgProjects))
	for _, p := range orgProjects {
		orgProjectIDs = append(orgProjectIDs, p.ID)
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
	if err := s.removeCustomResourceAccess(ctx, principalID, principalType, orgProjectIDs); err != nil {
		errs = errors.Join(errs, err)
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
		if err := s.removeGroupMemberRelation(ctx, g.ID, principalID, principalType); err != nil {
			return fmt.Errorf("remove group %s relations: %w", g.ID, err)
		}
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

// removeCustomResourceAccess deletes the principal's policies on custom
// resources in the org's projects. A policy only names a resource, so the
// resource service is needed to tell which project that resource belongs to.
func (s *Service) removeCustomResourceAccess(ctx context.Context, principalID, principalType string, orgProjectIDs []string) error {
	if s.resourceService == nil {
		s.log.ErrorContext(ctx, "resource service not set: custom-resource policies left behind",
			"principal_id", principalID,
			"principal_type", principalType,
		)
		return nil
	}
	if err := s.resourceService.RemovePrincipalAccess(ctx, principalID, principalType, orgProjectIDs); err != nil {
		return fmt.Errorf("remove custom-resource policies: %w", err)
	}
	return nil
}

// validateMinOwnerConstraint ensures the org always has at least one owner after a role change.
// Returns the resolved owner role ID for reuse by callers.
func (s *Service) validateMinOwnerConstraint(ctx context.Context, orgID, newRoleID string, existing []policy.Policy) (string, error) {
	return s.validateMinRoleConstraint(ctx, schema.RoleOrganizationOwner, policy.Filter{OrgID: orgID}, newRoleID, existing, ErrLastOwnerRole)
}

// validateOrgRole checks that the role is valid for organization scope and returns it.
// A role is valid if it is either:
// - a platform-wide role scoped to organizations, or
// - a custom role created for this specific organization.
func (s *Service) validateOrgRole(ctx context.Context, roleID, orgID string) (role.Role, error) {
	return s.validateRoleForScope(ctx, roleID, orgID, schema.OrganizationNamespace, ErrInvalidOrgRole)
}
