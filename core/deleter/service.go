package deleter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/raystack/frontier/core/audit"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/billing/checkout"

	"github.com/raystack/frontier/billing/invoice"

	"github.com/raystack/frontier/billing/customer"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/membership"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/core/group"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/serviceuser"
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
	DeleteModel(ctx context.Context, id string) error
}

type MembershipService interface {
	OnGroupDeleted(ctx context.Context, groupID string) error
	ListResourcesByPrincipal(ctx context.Context, principal authenticate.Principal, resourceType string, filter membership.ResourceFilter) ([]string, error)
	ForceRemoveOrganizationMember(ctx context.Context, orgID, principalID, principalType string) error
}

type InvitationService interface {
	List(ctx context.Context, flt invitation.Filter) ([]invitation.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	Delete(ctx context.Context, id string) error
}

type UserPATService interface {
	DeleteAllByUser(ctx context.Context, userID string) error
}

type ServiceUserService interface {
	List(ctx context.Context, flt serviceuser.Filter) ([]serviceuser.ServiceUser, error)
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

type CheckoutService interface {
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
	DeleteByCustomer(ctx context.Context, customerID string) error
}

type CreditService interface {
	DeleteByAccountID(ctx context.Context, accountID string) error
}

type KycService interface {
	DeleteKyc(ctx context.Context, orgID string) error
}

type Service struct {
	projService        ProjectService
	orgService         OrganizationService
	resService         ResourceService
	groupService       GroupService
	membershipService  MembershipService
	policyService      PolicyService
	roleService        RoleService
	invitationService  InvitationService
	userService        UserService
	userPATService     UserPATService
	serviceUserService ServiceUserService
	customerService    CustomerService
	subService         SubscriptionService
	invoiceService     InvoiceService
	checkoutService    CheckoutService
	creditService      CreditService
	kycService         KycService
}

func NewCascadeDeleter(orgService OrganizationService, projService ProjectService,
	resService ResourceService, groupService GroupService,
	membershipService MembershipService,
	policyService PolicyService, roleService RoleService,
	invitationService InvitationService, userService UserService,
	userPATService UserPATService,
	serviceUserService ServiceUserService,
	customerService CustomerService, subService SubscriptionService,
	invoiceService InvoiceService, checkoutService CheckoutService,
	creditService CreditService, kycService KycService) *Service {
	return &Service{
		projService:        projService,
		orgService:         orgService,
		resService:         resService,
		groupService:       groupService,
		membershipService:  membershipService,
		policyService:      policyService,
		roleService:        roleService,
		invitationService:  invitationService,
		userService:        userService,
		userPATService:     userPATService,
		serviceUserService: serviceUserService,
		customerService:    customerService,
		subService:         subService,
		invoiceService:     invoiceService,
		checkoutService:    checkoutService,
		creditService:      creditService,
		kycService:         kycService,
	}
}

func (d Service) DeleteProject(ctx context.Context, id string) error {
	// delete all project-level policies (and their rolebinding relations)
	policies, err := d.policyService.List(ctx, policy.Filter{
		ProjectID: id,
	})
	if err != nil {
		return err
	}
	for _, p := range policies {
		if err = d.policyService.Delete(ctx, p.ID); err != nil {
			return fmt.Errorf("failed to delete project while deleting a policy[%s]: %w", p.ID, err)
		}
	}

	// delete all related resources
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

// DeleteGroup orchestrates teardown of a single group: clears every member's
// policies and relations plus the org<->group hierarchy relations via
// membership, then deletes the group entity itself.
func (d Service) DeleteGroup(ctx context.Context, id string) error {
	if err := d.membershipService.OnGroupDeleted(ctx, id); err != nil {
		return fmt.Errorf("clean up group membership: %w", err)
	}
	return d.groupService.DeleteModel(ctx, id)
}

// DeleteOrganization removes an org and everything belonging to it. Billing is
// torn down first: it talks to the billing provider and is the step most
// likely to fail. Identity (projects, policies, and so on) goes after, with
// org policies (the org owners) near the end. This way a failure at any step
// leaves the org owned and the delete can simply be run again. Every step
// treats already-deleted data as success for the same reason.
func (d Service) DeleteOrganization(ctx context.Context, id string) error {
	// check if delete is allowed
	if err := d.canDelete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrDeleteNotAllowed)
	}

	// delete all billing accounts
	if err := d.DeleteCustomers(ctx, id); err != nil {
		return err
	}

	// delete all related projects
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
		if err = d.DeleteGroup(ctx, g.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a group[%s]: %w", g.Name, err)
		}
	}

	// delete all service users (clears credentials, org membership, and SpiceDB tuples)
	serviceUsers, err := d.serviceUserService.List(ctx, serviceuser.Filter{OrgID: id})
	if err != nil {
		return err
	}
	for _, su := range serviceUsers {
		if err = d.serviceUserService.Delete(ctx, su.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a service user[%s]: %w", su.ID, err)
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

	// delete the org's kyc record; its row references the org row. It goes
	// before the policies so a failure here leaves the org owned
	if err := d.kycService.DeleteKyc(ctx, id); err != nil {
		return fmt.Errorf("failed to delete org while deleting its kyc record: %w", err)
	}

	// delete all policies; this removes the org owners, so it stays after the
	// steps above. Policies must go before roles as they refer to them.
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

	if err := d.orgService.DeleteModel(ctx, id); err != nil {
		return err
	}

	if err := audit.NewLogger(ctx, id).Log(audit.OrgDeletedEvent, audit.OrgTarget(id)); err != nil {
		slog.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.OrgDeletedEvent)
	}
	return nil
}

func (d Service) DeleteCustomers(ctx context.Context, id string) error {
	customers, err := d.customerService.List(ctx, customer.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}
	for _, c := range customers {
		// cancels active subscriptions on the billing provider and removes local records
		if err := d.subService.DeleteByCustomer(ctx, c); err != nil {
			return fmt.Errorf("failed to delete org while deleting a billing account subscriptions[%s]: %w", c.ID, err)
		}
		if err := d.invoiceService.DeleteByCustomer(ctx, c); err != nil {
			return fmt.Errorf("failed to delete org while deleting a billing account invoices[%s]: %w", c.ID, err)
		}
		checkouts, err := d.checkoutService.List(ctx, checkout.Filter{CustomerID: c.ID})
		if err != nil {
			return fmt.Errorf("failed to delete org while listing billing account checkouts[%s]: %w", c.ID, err)
		}
		if err := d.checkoutService.DeleteByCustomer(ctx, c.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a billing account checkouts[%s]: %w", c.ID, err)
		}
		// the checkout rows are gone after this; the audit records keep the
		// provider references so the sessions can still be found on the provider.
		// The records are only written when ctx carries the audit service (the
		// API path seeds it); otherwise audit.NewLogger falls back to a noop
		auditLogger := audit.NewLogger(ctx, id)
		for _, ch := range checkouts {
			attrs := map[string]string{
				"provider_id":    ch.ProviderID,
				"customer_id":    ch.CustomerID,
				"plan_id":        ch.PlanID,
				"product_id":     ch.ProductID,
				"state":          ch.State,
				"payment_status": ch.PaymentStatus,
			}
			for k, v := range attrs {
				if v == "" {
					delete(attrs, k)
				}
			}
			if err := auditLogger.LogWithAttrs(audit.BillingCheckoutDeletedEvent, audit.Target{
				ID:   ch.ID,
				Type: "billing_checkout",
			}, attrs); err != nil {
				slog.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.BillingCheckoutDeletedEvent, "checkout_id", ch.ID)
			}
		}
		if err := d.creditService.DeleteByAccountID(ctx, c.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a billing account transactions[%s]: %w", c.ID, err)
		}

		// delete customer
		if err = d.customerService.Delete(ctx, c.ID); err != nil {
			return fmt.Errorf("failed to delete org while deleting a billing account[%s]: %w", c.ID, err)
		}
	}
	return nil
}

// RemoveUsersFromOrg removes users from an organization as members. The policy
// and relation cleanup — org, project, group, and custom resource — is
// delegated to membership.ForceRemoveOrganizationMember. It uses the force
// variant because a deletion cascade must succeed even when the user is the
// org's last owner.
func (d Service) RemoveUsersFromOrg(ctx context.Context, orgID string, userIDs []string) error {
	var errs error
	for _, userID := range userIDs {
		if memberErr := d.membershipService.ForceRemoveOrganizationMember(ctx, orgID, userID, schema.UserPrincipal); memberErr != nil {
			errs = errors.Join(errs, memberErr)
		}
	}
	return errs
}

// DeleteUser visits every org the user has a policy on (disabled orgs too),
// otherwise userService.Delete would leave orphan policy rows behind.
func (d Service) DeleteUser(ctx context.Context, userID string) error {
	orgIDs, err := d.membershipService.ListResourcesByPrincipal(ctx, authenticate.Principal{
		ID:   userID,
		Type: schema.UserPrincipal,
	}, schema.OrganizationNamespace, membership.ResourceFilter{})
	if err != nil {
		return err
	}
	for _, orgID := range orgIDs {
		if err = d.RemoveUsersFromOrg(ctx, orgID, []string{userID}); err != nil {
			return fmt.Errorf("failed to delete user from org[%s]: %w", orgID, err)
		}
	}
	if err := d.userPATService.DeleteAllByUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user PATs: %w", err)
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
