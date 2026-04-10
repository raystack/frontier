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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newMocks(t *testing.T) (
	*mocks.OrganizationService,
	*mocks.ProjectService,
	*mocks.ResourceService,
	*mocks.GroupService,
	*mocks.PolicyService,
	*mocks.RoleService,
	*mocks.InvitationService,
	*mocks.UserService,
	*mocks.CustomerService,
	*mocks.SubscriptionService,
	*mocks.InvoiceService,
) {
	return mocks.NewOrganizationService(t),
		mocks.NewProjectService(t),
		mocks.NewResourceService(t),
		mocks.NewGroupService(t),
		mocks.NewPolicyService(t),
		mocks.NewRoleService(t),
		mocks.NewInvitationService(t),
		mocks.NewUserService(t),
		mocks.NewCustomerService(t),
		mocks.NewSubscriptionService(t),
		mocks.NewInvoiceService(t)
}

func TestDeleteProject(t *testing.T) {
	t.Run("deletes policies, resources, then project model", func(t *testing.T) {
		orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)

		polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{{ID: "pol-1"}, {ID: "pol-2"}}, nil)
		polSvc.EXPECT().Delete(mock.Anything, "pol-1").Return(nil)
		polSvc.EXPECT().Delete(mock.Anything, "pol-2").Return(nil)

		resSvc.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{{ID: "res-1", NamespaceID: "ns-1", Name: "r1"}}, nil)
		resSvc.EXPECT().Delete(mock.Anything, "ns-1", "res-1").Return(nil)

		projSvc.EXPECT().DeleteModel(mock.Anything, "proj-1").Return(nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteProject(context.Background(), "proj-1")
		assert.NoError(t, err)
	})

	t.Run("returns error when policy list fails", func(t *testing.T) {
		_, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)
		orgSvc := mocks.NewOrganizationService(t)

		polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return(nil, errors.New("db error"))

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteProject(context.Background(), "proj-1")
		assert.ErrorContains(t, err, "db error")
	})

	t.Run("returns error when policy delete fails", func(t *testing.T) {
		_, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)
		orgSvc := mocks.NewOrganizationService(t)

		polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{{ID: "pol-fail"}}, nil)
		polSvc.EXPECT().Delete(mock.Anything, "pol-fail").Return(errors.New("delete error"))

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteProject(context.Background(), "proj-1")
		assert.ErrorContains(t, err, "pol-fail")
	})

	t.Run("no policies — still deletes resources and project", func(t *testing.T) {
		orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)

		polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{}, nil)
		resSvc.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{}, nil)
		projSvc.EXPECT().DeleteModel(mock.Anything, "proj-1").Return(nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteProject(context.Background(), "proj-1")
		assert.NoError(t, err)
	})
}

func TestDeleteOrganization(t *testing.T) {
	t.Run("full cascade delete", func(t *testing.T) {
		orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)

		// canDelete: no customers
		custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{}, nil)

		// policies
		polSvc.EXPECT().List(mock.Anything, policy.Filter{OrgID: "org-1"}).
			Return([]policy.Policy{{ID: "org-pol-1"}}, nil)
		polSvc.EXPECT().Delete(mock.Anything, "org-pol-1").Return(nil)

		// projects (each triggers DeleteProject)
		projSvc.EXPECT().List(mock.Anything, project.Filter{OrgID: "org-1"}).
			Return([]project.Project{{ID: "proj-1", Name: "p1"}}, nil)
		// DeleteProject for proj-1
		polSvc.EXPECT().List(mock.Anything, policy.Filter{ProjectID: "proj-1"}).
			Return([]policy.Policy{}, nil)
		resSvc.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{}, nil)
		projSvc.EXPECT().DeleteModel(mock.Anything, "proj-1").Return(nil)

		// groups
		grpSvc.EXPECT().List(mock.Anything, group.Filter{OrganizationID: "org-1"}).
			Return([]group.Group{{ID: "grp-1", Name: "g1"}}, nil)
		grpSvc.EXPECT().Delete(mock.Anything, "grp-1").Return(nil)

		// roles
		roleSvc.EXPECT().List(mock.Anything, role.Filter{OrgID: "org-1"}).
			Return([]role.Role{{ID: "role-1", Name: "r1"}}, nil)
		roleSvc.EXPECT().Delete(mock.Anything, "role-1").Return(nil)

		// invitations
		invID := uuid.New()
		invSvc.EXPECT().List(mock.Anything, invitation.Filter{OrgID: "org-1"}).
			Return([]invitation.Invitation{{ID: invID}}, nil)
		invSvc.EXPECT().Delete(mock.Anything, invID).Return(nil)

		// billing (no customers from canDelete, but DeleteCustomers is called separately)
		custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{}, nil)

		// finally delete org model
		orgSvc.EXPECT().DeleteModel(mock.Anything, "org-1").Return(nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteOrganization(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("blocked when billed customer has invoices", func(t *testing.T) {
		_, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)
		orgSvc := mocks.NewOrganizationService(t)

		custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{{ID: "cust-1", ProviderID: "stripe-1"}}, nil)
		invocSvc.EXPECT().List(mock.Anything, invoice.Filter{CustomerID: "cust-1"}).
			Return([]invoice.Invoice{{ID: "inv-1"}}, nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteOrganization(context.Background(), "org-1")
		assert.ErrorIs(t, err, deleter.ErrDeleteNotAllowed)
	})
}

func TestDeleteCustomers(t *testing.T) {
	t.Run("deletes subscriptions invoices and customer", func(t *testing.T) {
		orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)

		c := customer.Customer{ID: "cust-1", ProviderID: "stripe-1"}
		custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		subSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		invocSvc.EXPECT().DeleteByCustomer(mock.Anything, c).Return(nil)
		custSvc.EXPECT().Delete(mock.Anything, "cust-1").Return(nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteCustomers(context.Background(), "org-1")
		assert.NoError(t, err)
	})

	t.Run("skips subscription and invoice delete when no provider", func(t *testing.T) {
		orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)

		c := customer.Customer{ID: "cust-no-provider", ProviderID: ""}
		custSvc.EXPECT().List(mock.Anything, customer.Filter{OrgID: "org-1"}).
			Return([]customer.Customer{c}, nil)
		// no sub or invoice delete expected
		custSvc.EXPECT().Delete(mock.Anything, "cust-no-provider").Return(nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteCustomers(context.Background(), "org-1")
		assert.NoError(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("removes user from all orgs then deletes", func(t *testing.T) {
		orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc := newMocks(t)

		orgSvc.EXPECT().ListByUser(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, nil)
		usrSvc.EXPECT().Delete(mock.Anything, "user-1").Return(nil)

		svc := deleter.NewCascadeDeleter(orgSvc, projSvc, resSvc, grpSvc, polSvc, roleSvc, invSvc, usrSvc, custSvc, subSvc, invocSvc)
		err := svc.DeleteUser(context.Background(), "user-1")
		assert.NoError(t, err)
	})
}
