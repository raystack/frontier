package deleter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raystack/shield/core/invitation"

	"github.com/raystack/shield/core/policy"
	"github.com/raystack/shield/core/role"

	"github.com/raystack/shield/core/group"

	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/core/resource"
)

type ProjectService interface {
	List(ctx context.Context, flt project.Filter) ([]project.Project, error)
	DeleteModel(ctx context.Context, id string) error
}

type OrganizationService interface {
	Get(ctx context.Context, id string) (organization.Organization, error)
	DeleteModel(ctx context.Context, id string) error
}

type RoleService interface {
	List(ctx context.Context, flt role.Filter) ([]role.Role, error)
	Delete(ctx context.Context, id string) error
}

type PolicyService interface {
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
}

type ResourceService interface {
	List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error)
	Delete(ctx context.Context, namespaceID, id string) error
}

type GroupService interface {
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
	Delete(ctx context.Context, id string) error
}

type InvitationService interface {
	List(ctx context.Context, flt invitation.Filter) ([]invitation.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	projService       ProjectService
	orgService        OrganizationService
	resService        ResourceService
	groupService      GroupService
	policyService     PolicyService
	roleService       RoleService
	invitationService InvitationService
}

func NewCascadeDeleter(orgService OrganizationService, projService ProjectService,
	resService ResourceService, groupService GroupService,
	policyService PolicyService, roleService RoleService,
	invitationService InvitationService) *Service {
	return &Service{
		projService:       projService,
		orgService:        orgService,
		resService:        resService,
		groupService:      groupService,
		policyService:     policyService,
		roleService:       roleService,
		invitationService: invitationService,
	}
}

func (d Service) DeleteProject(ctx context.Context, id string) error {
	// delete all related resources first
	resources, err := d.resService.List(ctx, resource.Filter{
		ProjectID: id,
	})
	if err != nil {
		return err
	}
	for _, r := range resources {
		if err = d.resService.Delete(ctx, r.NamespaceID, r.ID); err != nil {
			return fmt.Errorf("failed to delete project while deleting a resource[%s]: %w", r.Name, err)
		}
	}

	return d.projService.DeleteModel(ctx, id)
}

func (d Service) DeleteOrganization(ctx context.Context, id string) error {
	// delete all policies
	policies, err := d.policyService.List(ctx, policy.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}
	for _, p := range policies {
		if err = d.policyService.Delete(ctx, p.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a policy[%s]: %w", p.ID, err)
		}
	}

	// delete all related projects first
	projects, err := d.projService.List(ctx, project.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}
	for _, p := range projects {
		if err = d.DeleteProject(ctx, p.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a project[%s]: %w", p.Name, err)
		}
	}

	// delete all related groups
	groups, err := d.groupService.List(ctx, group.Filter{OrganizationID: id})
	if err != nil {
		return err
	}
	for _, g := range groups {
		if err = d.groupService.Delete(ctx, g.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a group[%s]: %w", g.Name, err)
		}
	}

	// delete all roles
	roles, err := d.roleService.List(ctx, role.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}
	for _, p := range roles {
		if err = d.roleService.Delete(ctx, p.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a role[%s]: %w", p.Name, err)
		}
	}

	// delete all invitations
	invitations, err := d.invitationService.List(ctx, invitation.Filter{OrgID: id})
	if err != nil {
		return err
	}
	for _, i := range invitations {
		if err = d.invitationService.Delete(ctx, i.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a invitation[%s]: %w", i.ID, err)
		}
	}

	return d.orgService.DeleteModel(ctx, id)
}
