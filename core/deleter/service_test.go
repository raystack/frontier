package deleter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/core/deleter/mocks"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
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
}

func newMocks(t *testing.T) deleterMocks {
	t.Helper()
	return deleterMocks{
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
	}
}

func (m deleterMocks) build() *deleter.Service {
	return deleter.NewCascadeDeleter(m.orgSvc, m.projSvc, m.resSvc, m.grpSvc, m.mbrSvc,
		m.polSvc, m.roleSvc, m.invSvc, m.usrSvc, m.patSvc, m.suSvc,
		m.custSvc, m.subSvc, m.invocSvc, m.checkoutSvc, m.creditSvc, m.kycSvc)
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

		// canDelete and DeleteCustomers both list customers
		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().List(mock.Anything, invoice.Filter{CustomerID: "cust-1"}).
			Return([]invoice.Invoice{}, nil)

		// billing teardown
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
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

		// org policies
		m.polSvc.EXPECT().List(mock.Anything, policy.Filter{OrgID: "org-1"}).
			Return([]policy.Policy{{ID: "org-pol-1"}}, nil)
		m.polSvc.EXPECT().Delete(mock.Anything, "org-pol-1").Return(nil)

		// roles
		m.roleSvc.EXPECT().List(mock.Anything, role.Filter{OrgID: "org-1"}).
			Return([]role.Role{{ID: "role-1", Name: "r1"}}, nil)
		m.roleSvc.EXPECT().Delete(mock.Anything, "role-1").Return(nil)

		// kyc, then the org model
		m.kycSvc.EXPECT().DeleteKyc(mock.Anything, "org-1").Return(nil)
		m.orgSvc.EXPECT().DeleteModel(mock.Anything, "org-1").Return(nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("blocked when billed customer has invoices", func(t *testing.T) {
		m := newMocks(t)

		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{{ID: "cust-1", ProviderID: "stripe-1"}}, nil)
		m.invocSvc.EXPECT().List(mock.Anything, invoice.Filter{CustomerID: "cust-1"}).
			Return([]invoice.Invoice{{ID: "inv-1"}}, nil)

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorIs(t, err, deleter.ErrDeleteNotAllowed)
	})

	t.Run("billing failure stops the cascade before identity teardown", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.invocSvc.EXPECT().List(mock.Anything, invoice.Filter{CustomerID: "cust-1"}).
			Return([]invoice.Invoice{}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).
			Return(errors.New("provider is down"))
		// strict mocks: no policy, project, group, or org deletion may happen

		err := m.build().DeleteOrganization(context.Background(), "org-1")
		assert.ErrorContains(t, err, "provider is down")
	})

	t.Run("propagates error when service user list fails", func(t *testing.T) {
		m := newMocks(t)

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
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").Return(nil)
		m.creditSvc.EXPECT().DeleteByAccountID(mock.Anything, "cust-1").Return(nil)
		m.custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("offline account still removes local billing records", func(t *testing.T) {
		m := newMocks(t)

		c := customer.Customer{ID: "cust-no-provider", ProviderID: ""}
		m.custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		m.subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-no-provider").Return(nil)
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
		m.checkoutSvc.EXPECT().DeleteByCustomer(mock.Anything, "cust-1").
			Return(errors.New("checkout delete failed"))
		// strict mocks: custSvc.Delete must not be called

		err := m.build().DeleteCustomers(context.Background(), "org-1")
		assert.ErrorContains(t, err, "checkout delete failed")
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
