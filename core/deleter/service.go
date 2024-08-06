package deleter

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/billing/invoice"

	"github.com/raystack/frontier/billing/customer"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/invitation"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/core/group"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/resource"
)

const (
	DisableDeleteIfBilled = true
)

type ProjectService interface {
	List(ctx context.Context, flt project.Filter) ([]project.Project, error)
	DeleteModel(ctx context.Context, id string) error
}

type OrganizationService interface {
	Get(ctx context.Context, id string) (organization.Organization, error)
	DeleteModel(ctx context.Context, id string) error
	RemoveUsers(ctx context.Context, orgID string, userIDs []string) error
	ListByUser(ctx context.Context, userID string, f organization.Filter) ([]organization.Organization, error)
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
	Get(ctx context.Context, id string) (resource.Resource, error)
}

type GroupService interface {
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
	Delete(ctx context.Context, id string) error
	RemoveUsers(ctx context.Context, groupID string, userIDs []string) error
}

type InvitationService interface {
	List(ctx context.Context, flt invitation.Filter) ([]invitation.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	Delete(ctx context.Context, id string) error
}

type CustomerService interface {
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
}

type SubscriptionService interface {
	DeleteByCustomer(ctx context.Context, customr customer.Customer) error
}

type InvoiceService interface {
	List(ctx context.Context, flt invoice.Filter) ([]invoice.Invoice, error)
	DeleteByCustomer(ctx context.Context, customr customer.Customer) error
}

type Service struct {
	projService       ProjectService
	orgService        OrganizationService
	resService        ResourceService
	groupService      GroupService
	policyService     PolicyService
	roleService       RoleService
	invitationService InvitationService
	userService       UserService
	customerService   CustomerService
	subService        SubscriptionService
	invoiceService    InvoiceService
}

func NewCascadeDeleter(orgService OrganizationService, projService ProjectService,
	resService ResourceService, groupService GroupService,
	policyService PolicyService, roleService RoleService,
	invitationService InvitationService, userService UserService,
	customerService CustomerService, subService SubscriptionService,
	invoiceService InvoiceService) *Service {
	return &Service{
		projService:       projService,
		orgService:        orgService,
		resService:        resService,
		groupService:      groupService,
		policyService:     policyService,
		roleService:       roleService,
		invitationService: invitationService,
		userService:       userService,
		customerService:   customerService,
		subService:        subService,
		invoiceService:    invoiceService,
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
	// check if delete is allowed
	if err := d.canDelete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrDeleteNotAllowed)
	}

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

	// delete all billing accounts
	if err := d.DeleteCustomers(ctx, id); err != nil {
		return err
	}

	return d.orgService.DeleteModel(ctx, id)
}

func (d Service) DeleteCustomers(ctx context.Context, id string) error {
	customers, err := d.customerService.List(ctx, customer.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}
	for _, c := range customers {
		if c.ProviderID != "" {
			// delete all subscription first
			if err := d.subService.DeleteByCustomer(ctx, c); err != nil {
				return fmt.Errorf("failed to delete org while deleting a billing account subscriptions[%s]: %w", c.ID, err)
			}
			// delete all invoices
			if err := d.invoiceService.DeleteByCustomer(ctx, c); err != nil {
				return fmt.Errorf("failed to delete org while deleting a billing account invoices[%s]: %w", c.ID, err)
			}
		}

		// delete customer
		if err = d.customerService.Delete(ctx, c.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a billing account[%s]: %w", c.ID, err)
		}
	}
	return nil
}

// RemoveUsersFromOrg removes users from an organization as members
func (d Service) RemoveUsersFromOrg(ctx context.Context, orgID string, userIDs []string) error {
	var err error

	orgProjects, err := d.projService.List(ctx, project.Filter{
		OrgID: orgID,
	})
	if err != nil && !errors.Is(err, project.ErrNotExist) {
		return err
	}
	orgProjectIDs := utils.Map(orgProjects, func(p project.Project) string {
		return p.ID
	})
	// it's cheaper to fetch all groups in an org instead of fetching what all groups a user is part of
	orgGroups, err := d.groupService.List(ctx, group.Filter{
		OrganizationID: orgID,
	})
	if err != nil && !errors.Is(err, group.ErrNotExist) {
		return err
	}
	orgGroupIDs := utils.Map(orgGroups, func(g group.Group) string {
		return g.ID
	})

	for _, userID := range userIDs {
		userPolicies, policyErr := d.policyService.List(ctx, policy.Filter{
			PrincipalID:   userID,
			PrincipalType: schema.UserPrincipal,
		})
		if policyErr != nil && !errors.Is(err, policy.ErrNotExist) {
			err = errors.Join(err, policyErr)
			continue
		}

		for _, pol := range userPolicies {
			// delete org level roles
			switch pol.ResourceType {
			case schema.ProjectNamespace:
				// delete project level policies
				if utils.Contains(orgProjectIDs, pol.ResourceID) {
					if policyErr := d.policyService.Delete(ctx, pol.ID); policyErr != nil {
						err = errors.Join(err, policyErr)
					}
				}
			case schema.GroupNamespace:
				// delete group level policies
				if utils.Contains(orgGroupIDs, pol.ResourceID) {
					if groupErr := d.groupService.RemoveUsers(ctx, pol.ResourceID, []string{userID}); groupErr != nil {
						err = errors.Join(err, groupErr)
					}
				}
			case schema.PlatformNamespace, schema.OrganizationNamespace:
				// do nothing
			default:
				// delete resource level policies
				userResource, resErr := d.resService.Get(ctx, pol.ResourceID)
				if !errors.Is(resErr, resource.ErrNotExist) {
					if resErr != nil {
						err = errors.Join(err, resErr)
					} else if userResource.ProjectID != "" && utils.Contains(orgProjectIDs, userResource.ProjectID) {
						// if the resource belong to org project, delete access
						if policyErr := d.policyService.Delete(ctx, pol.ID); policyErr != nil {
							err = errors.Join(err, policyErr)
						}
					}
				}
			} // switch ends
		}
	}
	if err != nil {
		// abort if any error occurred
		return err
	}

	// remove user from org
	return d.orgService.RemoveUsers(ctx, orgID, userIDs)
}

func (d Service) DeleteUser(ctx context.Context, userID string) error {
	userOrgs, err := d.orgService.ListByUser(ctx, userID, organization.Filter{})
	if err != nil {
		return err
	}
	for _, org := range userOrgs {
		if err = d.RemoveUsersFromOrg(ctx, org.ID, []string{userID}); err != nil {
			return fmt.Errorf("failed to delete user from org[%s]: %w", org.Name, err)
		}
	}
	return d.userService.Delete(ctx, userID)
}

func (d Service) canDelete(ctx context.Context, id string) error {
	// check if any invoice is present for customer
	customers, err := d.customerService.List(ctx, customer.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}

	for _, c := range customers {
		if invoices, err := d.invoiceService.List(ctx, invoice.Filter{CustomerID: c.ID}); err != nil {
			return fmt.Errorf("failed to check invoices for billing account[%s]: %w", c.ID, err)
		} else if len(invoices) > 0 {
			if DisableDeleteIfBilled {
				return fmt.Errorf("cannot delete organization with billing account[%s]", c.ID)
			}
		}
	}
	return nil
}
