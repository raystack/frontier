package deleter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/core/deleter/mocks"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type deleterMocks struct {
	orgSvc      *mocks.OrganizationService
	projSvc     *mocks.ProjectService
	resSvc      *mocks.ResourceService
	grpSvc      *mocks.GroupService
	mbrSvc      *mocks.MembershipService
	polSvc      *mocks.PolicyService
	roleSvc     *mocks.RoleService
	invSvc      *mocks.InvitationService
	usrSvc      *mocks.UserService
	patSvc      *mocks.UserPATService
	suSvc       *mocks.ServiceUserService
	custSvc     *mocks.CustomerService
	subSvc      *mocks.SubscriptionService
	invocSvc    *mocks.InvoiceService
	checkoutSvc *mocks.CheckoutService
	creditSvc   *mocks.CreditService
	kycSvc      *mocks.KycService
	planSvc     *mocks.PlanService
}

func newMocks(t *testing.T) deleterMocks {
	t.Helper()
	m := deleterMocks{
		orgSvc:      mocks.NewOrganizationService(t),
		projSvc:     mocks.NewProjectService(t),
		resSvc:      mocks.NewResourceService(t),
		grpSvc:      mocks.NewGroupService(t),
		mbrSvc:      mocks.NewMembershipService(t),
		polSvc:      mocks.NewPolicyService(t),
		roleSvc:     mocks.NewRoleService(t),
		invSvc:      mocks.NewInvitationService(t),
		usrSvc:      mocks.NewUserService(t),
		patSvc:      mocks.NewUserPATService(t),
		suSvc:       mocks.NewServiceUserService(t),
		custSvc:     mocks.NewCustomerService(t),
		subSvc:      mocks.NewSubscriptionService(t),
		invocSvc:    mocks.NewInvoiceService(t),
		checkoutSvc: mocks.NewCheckoutService(t),
		creditSvc:   mocks.NewCreditService(t),
		kycSvc:      mocks.NewKycService(t),
		planSvc:     mocks.NewPlanService(t),
	}
	// plans resolve lazily by the subscription's plan id; stub a paid and a
	// free one for every test
	m.planSvc.EXPECT().GetByID(mock.Anything, "plan-paid").
		Return(plan.Plan{ID: "plan-paid", Products: []product.Product{
			{Prices: []product.Price{{Amount: 500}}},
		}}, nil).Maybe()
	m.planSvc.EXPECT().GetByID(mock.Anything, "plan-free").
		Return(plan.Plan{ID: "plan-free"}, nil).Maybe()
	return m
}

func (m deleterMocks) build() *deleter.Service {
	return deleter.NewCascadeDeleter(m.orgSvc, m.projSvc, m.resSvc, m.grpSvc, m.mbrSvc,
		m.polSvc, m.roleSvc, m.invSvc, m.usrSvc, m.patSvc, m.suSvc,
		m.custSvc, m.subSvc, m.invocSvc, m.checkoutSvc, m.creditSvc, m.kycSvc,
		m.planSvc)
}

func TestDeleteProject(t *testing.T) {
	t.Run("deletes policies, resources, then project model", func(t *testing.T) {
		m := newMocks(t)

		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{{ID: "pol-1"}, {ID: "pol-2"}}, nil)
		m.polSvc.EXPECT().Delete(mock.Anything, "pol-1").Return(nil)
		m.polSvc.EXPECT().Delete(mock.Anything, "pol-2").Return(nil)

		m.resSvc.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{{ID: "res-1", NamespaceID: "ns-1", Name: "r1"}}, nil)
		m.resSvc.EXPECT().Delete(mock.Anything, "ns-1", "res-1").Return(nil)

		m.projSvc.EXPECT().DeleteModel(mock.Anything, "proj-1").Return(nil)

		err := m.build().DeleteProject(context.Background(), "proj-1")
		assert.NoError(t, err)
	})

	t.Run("returns error when policy list fails", func(t *testing.T) {
		m := newMocks(t)

		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return(nil, errors.New("db error"))

		err := m.build().DeleteProject(context.Background(), "proj-1")
		assert.ErrorContains(t, err, "db error")
	})

	t.Run("returns error when policy delete fails", func(t *testing.T) {
		m := newMocks(t)

		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{{ID: "pol-fail"}}, nil)
		m.polSvc.EXPECT().Delete(mock.Anything, "pol-fail").Return(errors.New("delete error"))

		err := m.build().DeleteProject(context.Background(), "proj-1")
		assert.ErrorContains(t, err, "pol-fail")
	})

	t.Run("no policies — still deletes resources and project", func(t *testing.T) {
		m := newMocks(t)

		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{}, nil)
		m.resSvc.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{}, nil)
		m.projSvc.EXPECT().DeleteModel(mock.Anything, "proj-1").Return(nil)

		err := m.build().DeleteProject(context.Background(), "proj-1")
		assert.NoError(t, err)
	})
}

func TestDeleteOrganization(t *testing.T) {
	t.Run("full cascade delete", func(t *testing.T) {
		m := newMocks(t)

		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)

		// the up-front check and DeleteCustomers both list customers
		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{{ID: "sub-1", State: "canceled"}}, nil)

		// billing teardown
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{{ID: "chk-1", ProviderID: "cs_1", CustomerID: "cust-1"}}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-1").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		// projects (each triggers DeleteProject)
		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{{ID: "proj-1", Name: "p1"}}, nil)
		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{}, nil)
		m.resSvc.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{}, nil)
		m.projSvc.EXPECT().DeleteModel(mock.Anything, "proj-1").Return(nil)

		// groups (DeleteGroup: OnGroupDeleted then DeleteModel)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{{ID: "grp-1", Name: "g1"}}, nil)
		m.mbrSvc.EXPECT().OnGroupDeleted(mock.Anything, "grp-1").Return(nil)
		m.grpSvc.EXPECT().DeleteModel(mock.Anything, "grp-1").Return(nil)

		// service users
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return([]serviceuser.ServiceUser{{ID: "su-1"}}, nil)
		m.suSvc.EXPECT().Delete(mock.Anything, "su-1").Return(nil)

		// invitations
		invID := uuid.New()
		m.invSvc.EXPECT().List(mock.Anything, invitation.Filter{OrgID: "org-1"}).
			Return([]invitation.Invitation{{ID: invID}}, nil)
		m.invSvc.EXPECT().Delete(mock.Anything, invID).Return(nil)

		// kyc goes before the org policies
		m.kycSvc.EXPECT().DeleteKyc(mock.Anything, "org-1").Return(nil)

		// org policies
		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{OrgID: "org-1"}).
			Return([]policy.Policy{{ID: "org-pol-1"}}, nil)
		m.polSvc.EXPECT().Delete(mock.Anything, "org-pol-1").Return(nil)

		// roles
		m.roleSvc.EXPECT().List(mock.Anything, role.Filter{OrgID: "org-1"}).
			Return([]role.Role{{ID: "role-1", Name: "r1"}}, nil)
		m.roleSvc.EXPECT().Delete(mock.Anything, "role-1").Return(nil)

		// org model
		m.orgSvc.EXPECT().DeleteModel(mock.Anything, "org-1").Return(nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("already deleted org returns not found without touching anything", func(t *testing.T) {
		m := newMocks(t)

		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{}, organization.ErrNotExist)
		// strict mocks: no other service may be called

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorIs(t, err, organization.ErrNotExist)
	})

	t.Run("collects all blockers in one error", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{
				{ID: "sub-1", State: "active", PlanID: "plan-paid"},
				{ID: "sub-2", State: "canceled", PlanID: "plan-paid"},
			}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{{ID: "inv-1", State: invoice.OpenState}}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(-50, nil)
		// strict mocks: nothing may be deleted and no subscription may be
		// canceled on a blocked delete

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Equal(t, "org-1", blocked.OrgID)
		types := make([]string, 0, len(blocked.Blockers))
		for _, b := range blocked.Blockers {
			types = append(types, b.Type)
		}
		assert.Equal(t, []string{
			deleter.BlockerActiveSubscription,
			deleter.BlockerUnpaidInvoice,
			deleter.BlockerNegativeTokenBalance,
		}, types)
	})

	t.Run("paid plan subscription blocks until the caller downgrades", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{{ID: "sub-1", State: "active", PlanID: "plan-paid"}}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		// strict mocks: the paid subscription must not be canceled

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerActiveSubscription, blocked.Blockers[0].Type)
		assert.Contains(t, blocked.Blockers[0].Message, "downgrade to the standard plan")
	})

	t.Run("negative token balance blocks the delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(-50, nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerNegativeTokenBalance, blocked.Blockers[0].Type)
		assert.Contains(t, blocked.Blockers[0].Message, "contact support")
	})

	t.Run("unused tokens do not block the delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(100, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{}, nil)

		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-1").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{}, nil)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{}, nil)
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return([]serviceuser.ServiceUser{}, nil)
		m.invSvc.EXPECT().List(mock.Anything, invitation.Filter{OrgID: "org-1"}).
			Return([]invitation.Invitation{}, nil)
		m.kycSvc.EXPECT().DeleteKyc(mock.Anything, "org-1").Return(nil)
		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{OrgID: "org-1"}).
			Return([]policy.Policy{}, nil)
		m.roleSvc.EXPECT().List(mock.Anything, role.Filter{OrgID: "org-1"}).
			Return([]role.Role{}, nil)
		m.orgSvc.EXPECT().DeleteModel(mock.Anything, "org-1").Return(nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("running subscription is canceled by the delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{
				{ID: "sub-1", State: "active", PlanID: "plan-free"},
				{ID: "sub-2", State: "canceled", PlanID: "plan-paid"},
			}, nil)
		// only the running standard-plan subscription is canceled, immediately
		m.subSvc.EXPECT().Cancel(mock.Anything, "sub-1", true).
			Return(subscription.Subscription{ID: "sub-1", State: "canceled"}, nil)

		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-1").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{}, nil)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{}, nil)
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return([]serviceuser.ServiceUser{}, nil)
		m.invSvc.EXPECT().List(mock.Anything, invitation.Filter{OrgID: "org-1"}).
			Return([]invitation.Invitation{}, nil)
		m.kycSvc.EXPECT().DeleteKyc(mock.Anything, "org-1").Return(nil)
		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{OrgID: "org-1"}).
			Return([]policy.Policy{}, nil)
		m.roleSvc.EXPECT().List(mock.Anything, role.Filter{OrgID: "org-1"}).
			Return([]role.Role{}, nil)
		m.orgSvc.EXPECT().DeleteModel(mock.Anything, "org-1").Return(nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("final invoice from the cancellation blocks the delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		// no unpaid invoice before the cancel, one after it
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil).Once()
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{{ID: "sub-1", State: "active", PlanID: "plan-free"}}, nil)
		m.subSvc.EXPECT().Cancel(mock.Anything, "sub-1", true).
			Return(subscription.Subscription{ID: "sub-1", State: "canceled"}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{{ID: "inv-final", State: invoice.OpenState}}, nil).Once()
		// strict mocks: nothing may be deleted

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerUnpaidInvoice, blocked.Blockers[0].Type)
		assert.Equal(t, "inv-final", blocked.Blockers[0].Subject)
	})

	t.Run("draft invoice blocks until the provider finalizes it", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{}, nil)
		// a renewal invoice the provider has not finalized yet, no local row
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{{ProviderID: "in_draft1", State: invoice.DraftState, Amount: 500}}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		// strict mocks: nothing may be deleted

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerUnpaidInvoice, blocked.Blockers[0].Type)
		assert.Equal(t, "in_draft1", blocked.Blockers[0].Subject)
		assert.Contains(t, blocked.Blockers[0].Message, "being prepared")
	})

	t.Run("subscription gone on the provider does not brick the delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		// locally active free-plan sub whose Stripe copy no longer exists
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{{ID: "sub-1", State: "active", PlanID: "plan-free"}}, nil)
		m.subSvc.EXPECT().Cancel(mock.Anything, "sub-1", true).
			Return(subscription.Subscription{}, subscription.ErrSubscriptionOnProviderNotFound)

		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-1").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{}, nil)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{}, nil)
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return([]serviceuser.ServiceUser{}, nil)
		m.invSvc.EXPECT().List(mock.Anything, invitation.Filter{OrgID: "org-1"}).
			Return([]invitation.Invitation{}, nil)
		m.kycSvc.EXPECT().DeleteKyc(mock.Anything, "org-1").Return(nil)
		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{OrgID: "org-1"}).
			Return([]policy.Policy{}, nil)
		m.roleSvc.EXPECT().List(mock.Anything, role.Filter{OrgID: "org-1"}).
			Return([]role.Role{}, nil)
		m.orgSvc.EXPECT().DeleteModel(mock.Anything, "org-1").Return(nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("paid subscription appearing between the passes blocks instead of being canceled", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		// first pass sees nothing, a paid checkout completes in between
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{}, nil).Once()
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{{ID: "sub-new", State: "active", PlanID: "plan-paid"}}, nil).Once()
		// strict mocks: the paid subscription must not be canceled and
		// nothing may be deleted

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerActiveSubscription, blocked.Blockers[0].Type)
		assert.Equal(t, "sub-new", blocked.Blockers[0].Subject)
	})

	t.Run("dangling plan reference blocks instead of bricking the delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{{ID: "sub-1", State: "active", PlanID: "plan-gone"}}, nil)
		m.planSvc.EXPECT().GetByID(mock.Anything, "plan-gone").
			Return(plan.Plan{}, plan.ErrNotFound)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		// strict mocks: nothing may be canceled or deleted

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerActiveSubscription, blocked.Blockers[0].Type)
		assert.Contains(t, blocked.Blockers[0].Message, "cannot be resolved")
	})

	t.Run("offline account only gets token checks", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-offline", ProviderID: ""}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		// strict mocks: no subscription or invoice call may happen
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-offline").Return(-10, nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")

		var blocked *deleter.BlockedError
		assert.ErrorAs(t, err, &blocked)
		assert.Len(t, blocked.Blockers, 1)
		assert.Equal(t, deleter.BlockerNegativeTokenBalance, blocked.Blockers[0].Type)
	})

	t.Run("kyc delete failure keeps owner policies", func(t *testing.T) {
		m := newMocks(t)

		// a disabled org is still deletable
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{}, organization.ErrDisabled)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{}, nil)
		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{}, nil)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{}, nil)
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return([]serviceuser.ServiceUser{}, nil)
		m.invSvc.EXPECT().List(mock.Anything, invitation.Filter{OrgID: "org-1"}).
			Return([]invitation.Invitation{}, nil)
		m.kycSvc.EXPECT().DeleteKyc(mock.Anything, "org-1").
			Return(errors.New("kyc delete failed"))
		// strict mocks: no org policy, role, or org model deletion may happen

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorContains(t, err, "kyc delete failed")
	})

	t.Run("billing failure stops the cascade before identity teardown", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().ListPayableOnProvider(mock.Anything, c).
			Return([]invoice.Invoice{}, nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(0, nil)
		m.subSvc.EXPECT().List(mock.Anything, subscription.Filter{CustomerID: "cust-1"}).
			Return([]subscription.Subscription{}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).
			Return(errors.New("provider is down"))
		// strict mocks: no policy, project, group, or org deletion may happen

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorContains(t, err, "provider is down")
	})

	t.Run("propagates error when service user list fails", func(t *testing.T) {
		m := newMocks(t)

		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{}, nil)
		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{}, nil)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{}, nil)
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return(nil, errors.New("su list failed"))

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorContains(t, err, "su list failed")
	})

	t.Run("propagates error when service user delete fails", func(t *testing.T) {
		m := newMocks(t)

		m.orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1"}, nil)
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{}, nil)
		m.projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{}, nil)
		m.grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{}, nil)
		m.suSvc.EXPECT().List(mock.Anything, serviceuser.Filter{OrgID: "org-1"}).
			Return([]serviceuser.ServiceUser{{ID: "su-1"}}, nil)
		m.suSvc.EXPECT().Delete(mock.Anything, "su-1").Return(errors.New("su delete failed"))

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorContains(t, err, "su delete failed")
		assert.ErrorContains(t, err, "su-1")
	})
}

func TestDeleteCustomers(t *testing.T) {
	t.Run("deletes subscriptions invoices checkouts transactions and customer", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{
				{ID: "chk-1", ProviderID: "cs_1", CustomerID: "cust-1", PlanID: "plan-1", State: "complete", PaymentStatus: "paid"},
				{ID: "chk-2", ProviderID: "cs_2", CustomerID: "cust-1", State: "expired"},
			}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").Return(100, nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-1").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("balance check failure stops the customer delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-1").
			Return(0, errors.New("balance check failed"))
		// strict mocks: creditSvc.DeleteByAccountID and custSvc.Delete must not be called

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.ErrorContains(t, err, "balance check failed")
	})

	t.Run("offline account still removes local billing records", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-no-provider", ProviderID: ""}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-no-provider"}).
			Return([]checkout.Checkout{}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-no-provider").Return(nil)
		m.creditSvc.EXPECT().GetBalance(mock.Anything, "cust-no-provider").Return(0, nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-no-provider").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-no-provider").Return(nil)

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("checkout cleanup failure stops the customer delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return([]checkout.Checkout{}, nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").
			Return(errors.New("checkout delete failed"))
		// strict mocks: custSvc.Delete must not be called

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.ErrorContains(t, err, "checkout delete failed")
	})

	t.Run("checkout list failure stops the customer delete", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().List(mock.Anything, checkout.Filter{CustomerID: "cust-1"}).
			Return(nil, errors.New("checkout list failed"))
		// strict mocks: checkoutSvc.DeleteByCustomer and custSvc.Delete must not be called

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.ErrorContains(t, err, "checkout list failed")
	})
}

func TestRemoveUsersFromOrg(t *testing.T) {
	const orgID = "org-1"
	const userID = "user-1"

	t.Run("delegates the whole removal to the membership cascade", func(t *testing.T) {
		m := newMocks(t)

		// org, project, group, and custom-resource cleanup all happen inside the
		// cascade — strict mocks fail if the deleter touches them itself
		m.mbrSvc.EXPECT().ForceRemoveOrganizationMember(mock.Anything, orgID, userID, schema.UserPrincipal).Return(nil)

		err := m.build().RemoveUsersFromOrg(context.Background(), orgID, []string{userID})
		assert.NoError(t, err)
	})

	t.Run("propagates membership cascade failure", func(t *testing.T) {
		m := newMocks(t)

		m.mbrSvc.EXPECT().ForceRemoveOrganizationMember(mock.Anything, orgID, userID, schema.UserPrincipal).
			Return(errors.New("cascade boom"))

		err := m.build().RemoveUsersFromOrg(context.Background(), orgID, []string{userID})
		assert.ErrorContains(t, err, "cascade boom")
	})

	t.Run("keeps going to the next user after one fails", func(t *testing.T) {
		m := newMocks(t)

		m.mbrSvc.EXPECT().ForceRemoveOrganizationMember(mock.Anything, orgID, userID, schema.UserPrincipal).
			Return(errors.New("cascade boom"))
		m.mbrSvc.EXPECT().ForceRemoveOrganizationMember(mock.Anything, orgID, "user-2", schema.UserPrincipal).Return(nil)

		err := m.build().RemoveUsersFromOrg(context.Background(), orgID, []string{userID, "user-2"})
		assert.ErrorContains(t, err, "cascade boom")
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("removes user from all orgs, cleans PATs, then deletes user", func(t *testing.T) {
		m := newMocks(t)

		m.mbrSvc.EXPECT().ListResourcesByPrincipal(mock.Anything, mock.Anything, schema.OrganizationNamespace, mock.Anything).
			Return(nil, nil)
		m.patSvc.EXPECT().DeleteAllByUser(mock.Anything, "user-1").Return(nil)
		m.usrSvc.EXPECT().Delete(mock.Anything, "user-1").Return(nil)

		err := m.build().DeleteUser(context.Background(), "user-1")
		assert.NoError(t, err)
	})

	t.Run("aborts before userService.Delete when PAT cleanup fails", func(t *testing.T) {
		m := newMocks(t)

		m.mbrSvc.EXPECT().ListResourcesByPrincipal(mock.Anything, mock.Anything, schema.OrganizationNamespace, mock.Anything).
			Return(nil, nil)
		m.patSvc.EXPECT().DeleteAllByUser(mock.Anything, "user-1").Return(errors.New("pat cleanup boom"))
		// usrSvc.Delete must NOT be called — strict mock fails on unexpected call.

		err := m.build().DeleteUser(context.Background(), "user-1")
		assert.ErrorContains(t, err, "pat cleanup boom")
	})
}
