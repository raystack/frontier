package membership

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
)

// ListPoliciesByPrincipal returns every policy held by the principal.
func (s *Service) ListPoliciesByPrincipal(ctx context.Context, principalID, principalType string) ([]policy.Policy, error) {
	return s.policyService.List(ctx, policy.Filter{
		PrincipalID:   principalID,
		PrincipalType: principalType,
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
