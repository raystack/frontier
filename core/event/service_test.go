package event

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/core/event/mocks"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleError = errors.New("sample error")

func mockService(t *testing.T) (*billing.Config, *mocks.CheckoutService, *mocks.CustomerService, *mocks.OrganizationService, *mocks.PlanService, *mocks.UserService, *mocks.MembershipService, *mocks.RoleService, *mocks.SubscriptionService, *mocks.CreditService, *mocks.InvoiceService) {
	t.Helper()
	billingConf := &billing.Config{
		StripeKey:            "test_key",
		StripeAutoTax:        false,
		StripeWebhookSecrets: nil,
		PlansPath:            "",
		DefaultCurrency:      "USD",
		AccountConfig:        billing.AccountConfig{AutoCreateWithOrg: true, DefaultPlan: "default_plan", DefaultOffline: false},
		PlanChangeConfig:     billing.PlanChangeConfig{},
		SubscriptionConfig:   billing.SubscriptionConfig{},
		ProductConfig:        billing.ProductConfig{},
		RefreshInterval:      billing.RefreshInterval{},
	}
	checkoutService := mocks.NewCheckoutService(t)
	customerService := mocks.NewCustomerService(t)
	orgService := mocks.NewOrganizationService(t)
	planService := mocks.NewPlanService(t)
	userService := mocks.NewUserService(t)
	membershipService := mocks.NewMembershipService(t)
	roleService := mocks.NewRoleService(t)
	subsService := mocks.NewSubscriptionService(t)
	creditService := mocks.NewCreditService(t)
	invoiceService := mocks.NewInvoiceService(t)

	return billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService
}

// expectAdminLookup sets up mock expectations for fetching the first org admin's email.
func expectAdminLookup(ctx context.Context, roleService *mocks.RoleService, membershipService *mocks.MembershipService, userService *mocks.UserService, orgID string, users []user.User) {
	ownerRoleID := "owner-role-id"
	roleService.On("Get", ctx, organization.AdminRole).Return(role.Role{ID: ownerRoleID}, nil).Once()
	members := make([]membership.Member, 0, len(users))
	for _, u := range users {
		members = append(members, membership.Member{PrincipalID: u.ID, PrincipalType: schema.UserPrincipal})
	}
	membershipService.On("ListPrincipalsByResource", ctx, orgID, schema.OrganizationNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
		RoleIDs:       []string{ownerRoleID},
	}).Return(members, nil).Once()
	if len(users) > 0 {
		userService.On("GetByIDs", ctx, []string{users[0].ID}).Return(users[:1], nil).Once()
	}
}

func TestEnsureDefaultPlan(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx   context.Context
		orgID string
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		setup   func() *Service
	}{
		{
			name: "return error if customerService.List returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, sampleError).Once()
				return service
			},
		},
		{
			name: "short circuit if customer account exists",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: nil,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return([]customer.Customer{{ID: "1"}}, nil).Once()
				return service
			},
		},
		{
			name: "return error if orgService.GetRaw returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return([]customer.Customer{}, nil).Once()
				orgService.On("GetRaw", ctx, "").Return(organization.Organization{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return error if org admin lookup returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				orgService.On("GetRaw", ctx, "").Return(organization.Organization{ID: "org_1"}, nil).Once()
				roleService.On("Get", ctx, organization.AdminRole).Return(role.Role{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return error if customerService.Create returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return error if planService.GetByID returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{ID: "cid_1"}, nil).Once()
				planService.On("GetByID", ctx, "default_plan").Return(plan.Plan{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return error if defaultPlan is not free",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: DefaultPlanNotFree,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{ID: "cid_1"}, nil).Once()
				planService.On("GetByID", ctx, "default_plan").
					Return(plan.Plan{
						Products: []product.Product{
							{
								Prices: []product.Price{
									{
										Amount: 100.0,
									},
								},
							},
						},
					}, nil).Once()
				return service
			},
		},
		{
			name: "return error if checkoutService.Apply returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{ID: "cid_1"}, nil).Once()
				planService.On("GetByID", ctx, "default_plan").
					Return(plan.Plan{
						ID: "plan_1",
						Products: []product.Product{
							{
								Prices: []product.Price{
									{
										Amount: 0.0,
									},
								},
							},
						},
					}, nil).Once()
				checkoutService.On("Apply", ctx, checkout.Checkout{
					CustomerID: "cid_1",
					PlanID:     "plan_1",
					SkipTrial:  true,
				}).Return(nil, nil, sampleError).Once()
				return service
			},
		},
		{
			name: "return no error if creditService.Add returns no error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: nil,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				const onboardingAmount = 10.0
				billingConf.AccountConfig.OnboardCreditsWithOrg = onboardingAmount
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{ID: "cid_1"}, nil).Once()
				planService.On("GetByID", ctx, "default_plan").
					Return(plan.Plan{
						ID: "plan_1",
						Products: []product.Product{
							{
								Prices: []product.Price{
									{
										Amount: 0.0,
									},
								},
							},
						},
					}, nil).Once()
				checkoutService.On("Apply", ctx, checkout.Checkout{
					CustomerID: "cid_1",
					PlanID:     "plan_1",
					SkipTrial:  true,
				}).Return(nil, nil, nil).Once()
				creditService.On("Add", ctx, credit.Credit{
					ID:          "43b9d78f-ccd7-5011-88a7-27791d9baeb2",
					CustomerID:  "cid_1",
					Amount:      onboardingAmount,
					UserID:      "",
					Source:      "system.awarded",
					Description: "Awarded 10 credits for onboarding",
					Metadata:    metadata.Metadata{"auto_created": "true"},
				}).Return(nil).Once()
				return service
			},
		},
		{
			name: "return error if creditService.Add returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				const onboardingAmount = 10.0
				billingConf.AccountConfig.OnboardCreditsWithOrg = onboardingAmount
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{ID: "cid_1"}, nil).Once()
				planService.On("GetByID", ctx, "default_plan").
					Return(plan.Plan{
						ID: "plan_1",
						Products: []product.Product{
							{
								Prices: []product.Price{
									{
										Amount: 0.0,
									},
								},
							},
						},
					}, nil).Once()
				checkoutService.On("Apply", ctx, checkout.Checkout{
					CustomerID: "cid_1",
					PlanID:     "plan_1",
					SkipTrial:  true,
				}).Return(nil, nil, nil).Once()
				creditService.On("Add", ctx, credit.Credit{
					ID:          "43b9d78f-ccd7-5011-88a7-27791d9baeb2",
					CustomerID:  "cid_1",
					Amount:      onboardingAmount,
					UserID:      "",
					Source:      "system.awarded",
					Description: "Awarded 10 credits for onboarding",
					Metadata:    metadata.Metadata{"auto_created": "true"},
				}).Return(sampleError).Once()
				return service
			},
		},
		{
			name: "return no error if creditService.Add returns credit.ErrAlreadyApplied error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: nil,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService := mockService(t)
				const onboardingAmount = 10.0
				billingConf.AccountConfig.OnboardCreditsWithOrg = onboardingAmount
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, membershipService, roleService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				expectAdminLookup(ctx, roleService, membershipService, userService, "org_1", []user.User{{ID: "admin-1", Email: "email@example.com"}})
				customerService.On("Create", ctx, customer.Customer{
					OrgID:    "org_1",
					Name:     getCustomerName(org),
					Email:    "email@example.com",
					Currency: "USD",
					Metadata: map[string]any{
						"auto_created": "true",
					}}, false).Return(customer.Customer{ID: "cid_1"}, nil).Once()
				planService.On("GetByID", ctx, "default_plan").
					Return(plan.Plan{
						ID: "plan_1",
						Products: []product.Product{
							{
								Prices: []product.Price{
									{
										Amount: 0.0,
									},
								},
							},
						},
					}, nil).Once()
				checkoutService.On("Apply", ctx, checkout.Checkout{
					CustomerID: "cid_1",
					PlanID:     "plan_1",
					SkipTrial:  true,
				}).Return(nil, nil, nil).Once()
				creditService.On("Add", ctx, credit.Credit{
					ID:          "43b9d78f-ccd7-5011-88a7-27791d9baeb2",
					CustomerID:  "cid_1",
					Amount:      onboardingAmount,
					UserID:      "",
					Source:      "system.awarded",
					Description: "Awarded 10 credits for onboarding",
					Metadata:    metadata.Metadata{"auto_created": "true"},
				}).Return(credit.ErrAlreadyApplied).Once()
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			err := s.EnsureDefaultPlan(ctx, tt.args.orgID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
