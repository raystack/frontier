package membership

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/utils"
)

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

// OnProjectCreated runs right after a project is created. It links the project
// to its parent org in SpiceDB and gives the creator the project owner role.
// There is no org-membership check on the creator here: platform superusers
// can create projects in orgs they are not members of.
func (s *Service) OnProjectCreated(ctx context.Context, projectID, orgID, creatorID, creatorType string) error {
	if err := s.linkProjectToOrg(ctx, projectID, orgID); err != nil {
		return err
	}
	if _, err := s.createPolicy(ctx, projectID, schema.ProjectNamespace, creatorID, creatorType, schema.RoleProjectOwner); err != nil {
		if cleanupErr := s.unlinkProjectFromOrg(ctx, projectID, orgID); cleanupErr != nil {
			s.log.WarnContext(ctx, "project hierarchy cleanup failed after owner add failure",
				"project_id", projectID,
				"org_id", orgID,
				"error", cleanupErr,
			)
		}
		return err
	}
	return nil
}

// linkProjectToOrg creates the project->org relation in SpiceDB. Org-level
// project permissions (e.g. project delete via the org) resolve through it.
func (s *Service) linkProjectToOrg(ctx context.Context, projectID, orgID string) error {
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object:       relation.Object{ID: projectID, Namespace: schema.ProjectNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}); err != nil {
		return fmt.Errorf("link project to org: %w", err)
	}
	return nil
}

// unlinkProjectFromOrg removes the project->org relation. Used to clean up
// when project creation fails partway.
func (s *Service) unlinkProjectFromOrg(ctx context.Context, projectID, orgID string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object:       relation.Object{ID: projectID, Namespace: schema.ProjectNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return nil
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
