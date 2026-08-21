package deleter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/raystack/frontier/core/audit"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/billing"

	"github.com/raystack/frontier/billing/checkout"

	"github.com/raystack/frontier/billing/credit"

	"github.com/raystack/frontier/billing/invoice"

	"github.com/raystack/frontier/billing/customer"

	"github.com/raystack/frontier/billing/plan"

	"github.com/raystack/frontier/billing/subscription"

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
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/mailer"
)

type ProjectService interface {
	List(ctx context.Context, flt project.Filter) ([]project.Project, error)
	DeleteModel(ctx context.Context, id string) error
}

type OrganizationService interface {
	GetRaw(ctx context.Context, id string) (organization.Organization, error)
	DeleteModel(ctx context.Context, id string) error
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
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
	ListPrincipalsByResource(ctx context.Context, resourceID, resourceType string, filter membership.MemberFilter) ([]membership.Member, error)
	ForceRemoveOrganizationMember(ctx context.Context, orgID, principalID, principalType string) error
}

type InvitationService interface {
	List(ctx context.Context, flt invitation.Filter) ([]invitation.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
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
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
	Cancel(ctx context.Context, id string, immediate bool) (subscription.Subscription, error)
	DeleteByCustomer(ctx context.Context, customr customer.Customer) error
}

type InvoiceService interface {
	List(ctx context.Context, flt invoice.Filter) ([]invoice.Invoice, error)
	ListPayableOnProvider(ctx context.Context, customr customer.Customer) ([]invoice.Invoice, error)
	DeleteByCustomer(ctx context.Context, customr customer.Customer) error
}

type CheckoutService interface {
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
	DeleteByCustomer(ctx context.Context, customerID string) error
}

type CreditService interface {
	GetBalance(ctx context.Context, accountID string) (int64, error)
	List(ctx context.Context, flt credit.Filter) ([]credit.Transaction, error)
	DeleteByAccountID(ctx context.Context, accountID string) error
}

type KycService interface {
	DeleteKyc(ctx context.Context, orgID string) error
}

type PlanService interface {
	GetByID(ctx context.Context, id string) (plan.Plan, error)
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
	planService        PlanService
	// mailDialer and forfeitNoticeConfig drive the email that tells the org
	// owners about tokens forfeited by the delete
	mailDialer          mailer.Dialer
	forfeitNoticeConfig billing.TokenForfeitNoticeConfig
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
	creditService CreditService, kycService KycService,
	planService PlanService,
	mailDialer mailer.Dialer, forfeitNoticeConfig billing.TokenForfeitNoticeConfig) *Service {
	return &Service{
		projService:         projService,
		orgService:          orgService,
		resService:          resService,
		groupService:        groupService,
		membershipService:   membershipService,
		policyService:       policyService,
		roleService:         roleService,
		invitationService:   invitationService,
		userService:         userService,
		userPATService:      userPATService,
		serviceUserService:  serviceUserService,
		customerService:     customerService,
		subService:          subService,
		invoiceService:      invoiceService,
		checkoutService:     checkoutService,
		creditService:       creditService,
		kycService:          kycService,
		planService:         planService,
		mailDialer:          mailDialer,
		forfeitNoticeConfig: forfeitNoticeConfig,
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
// treats already-deleted data as success for the same reason. That applies
// to data inside a half-deleted org; an org whose row is already fully gone
// answers not found instead, so a caller can tell a finished delete from a
// repeatable one (and a mistyped org id from a success).
func (d Service) DeleteOrganization(ctx context.Context, id string) error {
	// an org that is already gone has nothing left to check or tear down;
	// GetRaw keeps disabled orgs deletable
	org, err := d.orgService.GetRaw(ctx, id)
	if err != nil {
		return err
	}

	customers, err := d.customerService.List(ctx, customer.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}

	// the token forfeit notice reads owners and balances, so it has to be
	// collected while they still exist; its balance reads are reused by the
	// blocker check and the teardown audit below
	notice, err := d.collectForfeitNotice(ctx, org, customers)
	if err != nil {
		return err
	}
	// a retry whose earlier attempt already tore down some accounts finds
	// their balances gone; their forfeits are recovered per account from the
	// audit records that attempt wrote
	d.recoverForfeitFromAudit(ctx, id, &notice)

	// clear what we can and collect what still blocks the delete, before
	// touching any data
	if err := d.ensureDeletable(ctx, id, customers, notice.Balances); err != nil {
		return err
	}

	// delete all billing accounts
	if err := d.deleteCustomers(ctx, id, customers, notice.Accounts); err != nil {
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

	// the org is gone; tell the owners about any tokens the delete forfeited
	if notice.Amount > 0 {
		d.sendForfeitNotices(ctx, org, notice)
	}
	return nil
}

// DeleteCustomers lists the accounts and reads the balances itself;
// DeleteOrganization goes through deleteCustomers with what it already
// collected.
func (d Service) DeleteCustomers(ctx context.Context, id string) error {
	customers, err := d.customerService.List(ctx, customer.Filter{
		OrgID: id,
	})
	if err != nil {
		return err
	}
	return d.deleteCustomers(ctx, id, customers, nil)
}

// deleteCustomers tears down the org's billing accounts. amounts carries the
// per-account token numbers the caller already read; nil means read them
// here.
func (d Service) deleteCustomers(ctx context.Context, id string, customers []customer.Customer, amounts map[string]accountTokens) error {
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
		// tokens still on the account are forfeited by this delete, so
		// record the amount before the transactions are removed. The
		// purchased share goes on the record too: the transaction rows are
		// deleted right after, and support settles a transfer from this
		// number later
		account, err := d.resolveAccountTokens(ctx, c.ID, amounts)
		if err != nil {
			return fmt.Errorf("failed to delete org while checking balance of billing account[%s]: %w", c.ID, err)
		}
		if account.Balance > 0 {
			if err := auditLogger.LogWithAttrs(audit.BillingTokensForfeitedEvent, audit.Target{
				ID:   c.ID,
				Type: "billing_account",
			}, map[string]string{
				"amount":    strconv.FormatInt(account.Balance, 10),
				"purchased": strconv.FormatInt(account.Purchased, 10),
			}); err != nil {
				slog.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.BillingTokensForfeitedEvent, "customer_id", c.ID)
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

// ensureDeletable collects everything that blocks deleting the organization and
// returns it all as one BlockedError, so the caller gets a full checklist
// instead of discovering blockers one retry at a time.
//
// A running subscription on a paid plan blocks the delete: the caller has
// to downgrade it to the standard plan first. A running subscription on a
// free plan is not a blocker — once no blockers are left, ensureDeletable
// cancels it itself, before any deletion starts. The cancel is immediate
// and bills unbilled usage on the spot, so the invoice check runs once more
// after it: a final invoice created by the cancel still blocks the delete
// until it is paid or its automatic charge settles.
//
// Unused tokens do not block either: the delete forfeits them. The client
// gets the caller's confirmation before sending the delete, and the
// forfeited amount is written to an audit record during teardown.
//
// Accounts without a billing provider are only checked for token balances:
// their subscription and invoice rows have nothing behind them the caller
// could cancel or pay.
func (d Service) ensureDeletable(ctx context.Context, id string, customers []customer.Customer, balances map[string]int64) error {
	// each plan resolves at most once per call, and only when a running
	// subscription actually references it
	paidPlans := map[string]bool{}

	var blockers []Blocker
	for _, c := range customers {
		if !c.IsOffline() {
			bs, err := d.subscriptionBlockers(ctx, c, paidPlans)
			if err != nil {
				return err
			}
			blockers = append(blockers, bs...)

			bs, err = d.invoiceBlockers(ctx, c)
			if err != nil {
				return err
			}
			blockers = append(blockers, bs...)
		}

		balance := balances[c.ID]
		// the balance goes below zero when the account has an overdraft
		// floor (credit_min under zero, the postpaid setup) and tokens were
		// spent on credit. That debt is money owed, so it must be settled
		// before the org can go
		if balance < 0 {
			blockers = append(blockers, Blocker{
				Type:    BlockerNegativeTokenBalance,
				Subject: c.ID,
				Message: fmt.Sprintf("billing account[%s] owes %d tokens: contact support to settle the balance, then retry the delete", c.ID, -balance),
			})
		}
	}
	if len(blockers) > 0 {
		return &BlockedError{OrgID: id, Blockers: blockers}
	}

	// no blockers were found, so the delete will happen. Only free-plan
	// subscriptions can still be running: cancel them, then check the
	// invoices again because the cancel may have created a final one. Every
	// account is judged again before anything is canceled — a paid
	// subscription created since the first pass, on any account, must block
	// the delete without costing another account its subscription first
	type cancelTarget struct {
		customer customer.Customer
		subID    string
	}
	var toCancel []cancelTarget
	for _, c := range customers {
		if c.IsOffline() {
			continue
		}
		subs, err := d.subService.List(ctx, subscription.Filter{CustomerID: c.ID})
		if err != nil {
			return fmt.Errorf("failed to check subscriptions for billing account[%s]: %w", c.ID, err)
		}
		for _, sub := range subs {
			if !sub.IsActive() {
				continue
			}
			if sub.PlanID == "" {
				blockers = append(blockers, unresolvablePlanBlocker(sub))
				continue
			}
			paid, err := d.isPaidPlan(ctx, sub.PlanID, paidPlans)
			if errors.Is(err, plan.ErrNotFound) {
				blockers = append(blockers, unresolvablePlanBlocker(sub))
				continue
			}
			if err != nil {
				return err
			}
			if paid {
				blockers = append(blockers, paidSubscriptionBlocker(sub))
				continue
			}
			toCancel = append(toCancel, cancelTarget{customer: c, subID: sub.ID})
		}
	}
	if len(blockers) > 0 {
		return &BlockedError{OrgID: id, Blockers: blockers}
	}

	canceledFor := map[string]customer.Customer{}
	for _, t := range toCancel {
		// a subscription whose provider copy is already gone has nothing
		// to cancel; the teardown removes its local rows
		if _, err := d.subService.Cancel(ctx, t.subID, true); err != nil && !errors.Is(err, subscription.ErrSubscriptionOnProviderNotFound) {
			return fmt.Errorf("failed to cancel subscription[%s] of billing account[%s]: %w", t.subID, t.customer.ID, err)
		}
		canceledFor[t.customer.ID] = t.customer
	}
	for _, c := range canceledFor {
		bs, err := d.invoiceBlockers(ctx, c)
		if err != nil {
			return err
		}
		blockers = append(blockers, bs...)
	}
	if len(blockers) > 0 {
		return &BlockedError{OrgID: id, Blockers: blockers}
	}
	return nil
}

// subscriptionBlockers returns a blocker for every running subscription on a
// paid plan; the caller downgrades those to the standard plan. Running
// free-plan subscriptions are not blockers, the delete cancels them itself.
func (d Service) subscriptionBlockers(ctx context.Context, c customer.Customer, paidPlans map[string]bool) ([]Blocker, error) {
	subs, err := d.subService.List(ctx, subscription.Filter{CustomerID: c.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to check subscriptions for billing account[%s]: %w", c.ID, err)
	}
	var blockers []Blocker
	for _, sub := range subs {
		if !sub.IsActive() {
			continue
		}
		// a missing or dangling plan reference must not make the org
		// undeletable, and "downgrade" is no advice for it — the caller
		// can still cancel the subscription themselves
		if sub.PlanID == "" {
			blockers = append(blockers, unresolvablePlanBlocker(sub))
			continue
		}
		paid, err := d.isPaidPlan(ctx, sub.PlanID, paidPlans)
		if errors.Is(err, plan.ErrNotFound) {
			blockers = append(blockers, unresolvablePlanBlocker(sub))
			continue
		}
		if err != nil {
			return nil, err
		}
		if paid {
			blockers = append(blockers, paidSubscriptionBlocker(sub))
		}
	}
	return blockers, nil
}

func unresolvablePlanBlocker(sub subscription.Subscription) Blocker {
	return Blocker{
		Type:    BlockerActiveSubscription,
		Subject: sub.ID,
		Message: fmt.Sprintf("subscription[%s] is %s on a plan that cannot be resolved: cancel the subscription, then retry the delete", sub.ID, sub.State),
	}
}

func paidSubscriptionBlocker(sub subscription.Subscription) Blocker {
	return Blocker{
		Type:    BlockerActiveSubscription,
		Subject: sub.ID,
		Message: fmt.Sprintf("subscription[%s] is %s on a paid plan: downgrade to the standard plan, then retry the delete", sub.ID, sub.State),
	}
}

// isPaidPlan reports whether the plan charges money, caching each plan for
// the length of one delete. Callers handle a subscription without a plan
// before coming here.
func (d Service) isPaidPlan(ctx context.Context, planID string, cache map[string]bool) (bool, error) {
	if paid, ok := cache[planID]; ok {
		return paid, nil
	}
	pln, err := d.planService.GetByID(ctx, planID)
	if err != nil {
		return false, fmt.Errorf("failed to resolve plan[%s]: %w", planID, err)
	}
	cache[planID] = !pln.IsFree()
	return cache[planID], nil
}

// invoiceBlockers returns a blocker for every invoice of the account that
// still asks for money. Open and uncollectible invoices the caller can pay.
// A draft is money the provider is still preparing to charge — the provider
// finalizes it shortly — and deleting before that would silently lose the
// charge, so it blocks too. The answer comes straight from the billing
// provider, so a just-paid invoice does not block and a just-created one
// does.
func (d Service) invoiceBlockers(ctx context.Context, c customer.Customer) ([]Blocker, error) {
	invoices, err := d.invoiceService.ListPayableOnProvider(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to check invoices for billing account[%s]: %w", c.ID, err)
	}
	var blockers []Blocker
	for _, inv := range invoices {
		// an invoice the sync has not stored yet carries no local id
		subject := inv.ID
		if subject == "" {
			subject = inv.ProviderID
		}
		message := fmt.Sprintf("invoice[%s] is unpaid: pay it via its hosted payment page, then retry the delete", subject)
		if inv.State == invoice.DraftState {
			message = fmt.Sprintf("invoice[%s] is still being prepared by the billing provider: retry the delete once it finalizes, then pay it", subject)
		}
		blockers = append(blockers, Blocker{
			Type:    BlockerUnpaidInvoice,
			Subject: subject,
			Message: message,
		})
	}
	return blockers, nil
}
