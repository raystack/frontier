package membership

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// removeGroupMemberRelation deletes the member relation for a principal on a group.
func (s *Service) removeGroupMemberRelation(ctx context.Context, groupID, principalID, principalType string) error {
	err := s.relationService.Delete(ctx, relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: principalID, Namespace: principalType},
		RelationName: schema.MemberRelationName,
	})
	if err != nil && !errors.Is(err, relation.ErrNotExist) {
		return fmt.Errorf("delete relation %s: %w", schema.MemberRelationName, err)
	}
	return nil
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
		if err := s.createRelation(ctx, groupID, schema.GroupNamespace, principalID, principalType, schema.MemberRelationName); err != nil {
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
	if hasExactlyRole(existing, resolvedRoleID) {
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

	if err := s.removeGroupMemberRelation(ctx, groupID, principalID, principalType); err != nil {
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
// destroyed: deletes every policy on the group and every member
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
	principals := make(map[principalKey]policy.Policy, len(policies))
	failed := make(map[principalKey]struct{}, len(policies))
	var errs error
	for _, p := range policies {
		key := policyPrincipalKey(p)
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
		if relErr := s.removeGroupMemberRelation(ctx, groupID, p.PrincipalID, p.PrincipalType); relErr != nil {
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

// linkGroupToOrg creates the group#org@organization identity link. It is what
// resolves org-level synthetic group permissions (e.g. group delete = org->group_delete).
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

// unlinkGroupFromOrg removes the group#org identity link. Used as best-effort
// cleanup when group-create wiring fails partway.
// relation.ErrNotExist is ignored; any other error is returned.
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
	return s.validateRoleForScope(ctx, roleID, orgID, schema.GroupNamespace, ErrInvalidGroupRole)
}

// validateMinGroupOwnerConstraint ensures the group keeps at least one owner
// after the role change. Returns the resolved group owner role ID so the
// caller can hand it to replacePolicy as a min-role guard, closing the TOCTOU
// race between this pre-check and the policy delete.
func (s *Service) validateMinGroupOwnerConstraint(ctx context.Context, groupID, newRoleID string, existing []policy.Policy) (string, error) {
	return s.validateMinRoleConstraint(ctx, schema.GroupOwnerRole, policy.Filter{GroupID: groupID}, newRoleID, existing, ErrLastGroupOwnerRole)
}
