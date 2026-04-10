package project

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/core/serviceuser"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
}

type ServiceuserService interface {
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
	GetByIDs(ctx context.Context, ids []string) ([]serviceuser.ServiceUser, error)
	FilterSudos(ctx context.Context, ids []string) ([]string, error)
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
	ProjectMemberCount(ctx context.Context, ids []string) ([]policy.MemberCount, error)
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type GroupService interface {
	Get(ctx context.Context, id string) (group.Group, error)
	GetByIDs(ctx context.Context, ids []string) ([]group.Group, error)
	ListByUser(ctx context.Context, principal authenticate.Principal, flt group.Filter) ([]group.Group, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
	suserService    ServiceuserService
	policyService   PolicyService
	authnService    AuthnService
	groupService    GroupService
	roleService     RoleService
}

func NewService(repository Repository, relationService RelationService, userService UserService,
	policyService PolicyService, authnService AuthnService, suserService ServiceuserService,
	groupService GroupService, roleService RoleService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
		policyService:   policyService,
		authnService:    authnService,
		suserService:    suserService,
		groupService:    groupService,
		roleService:     roleService,
	}
}

func (s Service) Get(ctx context.Context, idOrName string) (Project, error) {
	if utils.IsValidUUID(idOrName) {
		return s.repository.GetByID(ctx, idOrName)
	}
	return s.repository.GetByName(ctx, idOrName)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	currentPrincipal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Project{}, err
	}

	newProject, err := s.repository.Create(ctx, prj)
	if err != nil {
		return Project{}, err
	}

	if err = s.addProjectToOrg(ctx, newProject, prj.Organization.ID); err != nil {
		return Project{}, err
	}

	// make user administrator of the project
	if _, err = s.policyService.Create(ctx, policy.Policy{
		RoleID:        OwnerRole,
		ResourceID:    newProject.ID,
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   currentPrincipal.ID,
		PrincipalType: currentPrincipal.Type,
	}); err != nil {
		return Project{}, fmt.Errorf("failed to create owner policy for project %s: %w", newProject.ID, err)
	}
	return newProject, nil
}

func (s Service) List(ctx context.Context, f Filter) ([]Project, error) {
	projects, err := s.repository.List(ctx, f)
	if err != nil {
		return nil, err
	}

	if f.WithMemberCount && len(projects) > 0 {
		// get member count for each project
		projectIDs := utils.Map(projects, func(p Project) string {
			return p.ID
		})
		memberCounts, err := s.policyService.ProjectMemberCount(ctx, projectIDs)
		if err != nil {
			return nil, err
		}
		for i := range projects {
			for _, count := range memberCounts {
				if projects[i].ID == count.ID {
					projects[i].MemberCount = count.Count
				}
			}
		}
	}

	return projects, nil
}

func (s Service) ListByUser(ctx context.Context, principal authenticate.Principal,
	flt Filter) ([]Project, error) {
	subjectID, subjectType := principal.ResolveSubject()

	var projIDs []string
	var err error
	if flt.NonInherited {
		// direct added users
		projIDs, err = s.listNonInheritedProjectIDs(ctx, subjectID, subjectType)
	} else {
		projIDs, err = s.relationService.LookupResources(ctx, relation.Relation{
			Object:       relation.Object{Namespace: schema.ProjectNamespace},
			Subject:      relation.Subject{Namespace: subjectType, ID: subjectID},
			RelationName: MemberPermission,
		})
	}
	if err != nil {
		return nil, err
	}

	projIDs = utils.Deduplicate(projIDs)
	projIDs, err = s.intersectPATScope(ctx, principal, schema.ProjectNamespace, projIDs)
	if err != nil {
		return nil, err
	}
	if len(projIDs) == 0 {
		return []Project{}, nil
	}

	flt.ProjectIDs = projIDs
	return s.List(ctx, flt)
}

// listNonInheritedProjectIDs returns project IDs where the principal has direct
// role assignments (not inherited through org), including via group memberships.
func (s Service) listNonInheritedProjectIDs(ctx context.Context, principalID, principalType string) ([]string, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalType: principalType,
		PrincipalID:   principalID,
		ResourceType:  schema.ProjectNamespace,
	})
	if err != nil {
		return nil, err
	}
	var projIDs []string
	for _, pol := range policies {
		projIDs = append(projIDs, pol.ResourceID)
	}

	// projects added via group memberships
	groups, err := s.groupService.ListByUser(ctx,
		authenticate.Principal{ID: principalID, Type: principalType}, group.Filter{})
	if err != nil {
		return nil, err
	}
	groupIDs := utils.Map(groups, func(g group.Group) string { return g.ID })
	if len(groupIDs) > 0 {
		policies, err = s.policyService.List(ctx, policy.Filter{
			PrincipalType: schema.GroupPrincipal,
			PrincipalIDs:  groupIDs,
			ResourceType:  schema.ProjectNamespace,
		})
		if err != nil {
			return nil, err
		}
		for _, pol := range policies {
			projIDs = append(projIDs, pol.ResourceID)
		}
	}
	return projIDs, nil
}

// intersectPATScope narrows resource IDs to only those the PAT is scoped to.
func (s Service) intersectPATScope(ctx context.Context, principal authenticate.Principal,
	namespace string, resourceIDs []string) ([]string, error) {
	if principal.PAT == nil || len(resourceIDs) == 0 {
		return resourceIDs, nil
	}
	patIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object:       relation.Object{Namespace: namespace},
		Subject:      relation.Subject{ID: principal.PAT.ID, Namespace: schema.PATPrincipal},
		RelationName: schema.GetPermission,
	})
	if err != nil {
		return nil, err
	}
	return utils.Intersection(resourceIDs, patIDs), nil
}

func (s Service) Update(ctx context.Context, prj Project) (Project, error) {
	if utils.IsValidUUID(prj.ID) {
		return s.repository.UpdateByID(ctx, prj)
	}
	return s.repository.UpdateByName(ctx, prj)
}

func (s Service) ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error) {
	requestedProject, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	policies, err := s.policyService.List(ctx, policy.Filter{
		ProjectID: requestedProject.ID,
	})
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0)
	for _, pol := range policies {
		// get all users with the permission
		if pol.PrincipalType == schema.UserPrincipal {
			userIDs = append(userIDs, pol.PrincipalID)
		}
	}

	if len(userIDs) == 0 {
		// no users
		return []user.User{}, nil
	}

	return s.userService.GetByIDs(ctx, userIDs)
}

func (s Service) ListServiceUsers(ctx context.Context, id string, permissionFilter string) ([]serviceuser.ServiceUser, error) {
	requestedProject, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.Relation{
		Object: relation.Object{
			ID:        requestedProject.ID,
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: permissionFilter,
	})
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []serviceuser.ServiceUser{}, nil
	}

	// filter service users which got access because of SU permission
	// even if they are from same org, I think it's ideal to not list them
	userIDs, err = s.suserService.FilterSudos(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []serviceuser.ServiceUser{}, nil
	}
	return s.suserService.GetByIDs(ctx, userIDs)
}

func (s Service) ListGroups(ctx context.Context, id string) ([]group.Group, error) {
	requestedProject, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Note(kushsharma): we don't need relation service here as we don't care about inheritance for now
	// if we do ever need it, we will have to use relation service
	groupPolicies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalType: schema.GroupPrincipal,
		ProjectID:     requestedProject.ID,
	})
	if err != nil {
		return nil, err
	}
	if len(groupPolicies) == 0 {
		// no groups
		return []group.Group{}, nil
	}
	groupIDs := utils.Map(groupPolicies, func(p policy.Policy) string {
		return p.PrincipalID
	})
	return s.groupService.GetByIDs(ctx, groupIDs)
}

func (s Service) addProjectToOrg(ctx context.Context, prj Project, orgID string) error {
	rel := relation.Relation{
		Object: relation.Object{
			ID:        prj.ID,
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

// SetMemberRole sets a principal's role in a project.
// It deletes any existing project-level policies for the principal and creates a new one.
// Supported principal types: user, service user, group.
func (s Service) SetMemberRole(ctx context.Context, projectID, principalID, principalType, newRoleID string) error {
	prj, err := s.Get(ctx, projectID)
	if err != nil {
		return err
	}

	if err := s.validatePrincipal(ctx, prj.Organization.ID, principalID, principalType); err != nil {
		return err
	}

	if err := s.validateProjectRole(ctx, newRoleID); err != nil {
		return err
	}

	existingPolicies, err := s.policyService.List(ctx, policy.Filter{
		ProjectID:     projectID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return err
	}

	// skip if the principal already has exactly this role
	if len(existingPolicies) == 1 && existingPolicies[0].RoleID == newRoleID {
		return nil
	}

	for _, p := range existingPolicies {
		if err := s.policyService.Delete(ctx, p.ID); err != nil {
			return err
		}
	}

	_, err = s.policyService.Create(ctx, policy.Policy{
		RoleID:        newRoleID,
		ResourceID:    projectID,
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	return err
}

// RemoveMember removes a principal from a project by deleting all their project-level policies.
// Supported principal types: user, service user, group.
func (s Service) RemoveMember(ctx context.Context, projectID, principalID, principalType string) error {
	_, err := s.Get(ctx, projectID)
	if err != nil {
		return err
	}

	switch principalType {
	case schema.UserPrincipal, schema.ServiceUserPrincipal, schema.GroupPrincipal:
	default:
		return ErrInvalidPrincipalType
	}

	existingPolicies, err := s.policyService.List(ctx, policy.Filter{
		ProjectID:     projectID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return err
	}

	if len(existingPolicies) == 0 {
		return ErrNotMember
	}

	for _, p := range existingPolicies {
		if err := s.policyService.Delete(ctx, p.ID); err != nil {
			return err
		}
	}

	return nil
}

// validatePrincipal checks that the principal exists and belongs to the org.
// For users, org membership is checked via org-level policies.
// For service users and groups, org membership is checked via their org ID field.
func (s Service) validatePrincipal(ctx context.Context, orgID, principalID, principalType string) error {
	switch principalType {
	case schema.UserPrincipal:
		if _, err := s.userService.GetByID(ctx, principalID); err != nil {
			return err
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
		su, err := s.suserService.Get(ctx, principalID)
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

func (s Service) validateProjectRole(ctx context.Context, roleID string) error {
	fetchedRole, err := s.roleService.Get(ctx, roleID)
	if err != nil {
		return err
	}

	if !slices.Contains(fetchedRole.Scopes, schema.ProjectNamespace) {
		return ErrInvalidProjectRole
	}

	return nil
}

// DeleteModel doesn't delete the nested resource, only itself
func (s Service) DeleteModel(ctx context.Context, id string) error {
	// delete all relations where resource is an object
	// all relations where project is an subject should already been deleted by now
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.ProjectNamespace,
	}}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return s.repository.Delete(ctx, id)
}
