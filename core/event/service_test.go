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
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleError = errors.New("sample error")

func mockService(t *testing.T) (*billing.Config, *mocks.CheckoutService, *mocks.CustomerService, *mocks.OrganizationService, *mocks.PlanService, *mocks.UserService, *mocks.SubscriptionService, *mocks.CreditService, *mocks.InvoiceService) {
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
	subsService := mocks.NewSubscriptionService(t)
	creditService := mocks.NewCreditService(t)
	invoiceService := mocks.NewInvoiceService(t)

	return billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return([]customer.Customer{}, nil).Once()
				orgService.On("GetRaw", ctx, "").Return(organization.Organization{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return error if userService.ListByOrg returns error",
			args: args{
				ctx:   ctx,
				orgID: "",
			},
			wantErr: sampleError,
			setup: func() *Service {
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				orgService.On("GetRaw", ctx, "").Return(organization.Organization{ID: "org_1"}, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return(nil, sampleError).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				const onboardingAmount = 10.0
				billingConf.AccountConfig.OnboardCreditsWithOrg = onboardingAmount
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				const onboardingAmount = 10.0
				billingConf.AccountConfig.OnboardCreditsWithOrg = onboardingAmount
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
				billingConf, checkoutService, customerService, orgService, planService, userService, subsService, creditService, invoiceService := mockService(t)
				const onboardingAmount = 10.0
				billingConf.AccountConfig.OnboardCreditsWithOrg = onboardingAmount
				service := NewService(*billingConf, orgService, checkoutService, customerService, planService, userService, subsService, creditService, invoiceService)

				customerService.On("List", ctx, customer.Filter{}).Return(nil, nil).Once()
				org := organization.Organization{ID: "org_1", Title: "org_title"}
				orgService.On("GetRaw", ctx, "").Return(org, nil).Once()
				userService.On("ListByOrg", ctx, "org_1", organization.AdminRole).Return([]user.User{{Email: "email@example.com"}}, nil).Once()
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
