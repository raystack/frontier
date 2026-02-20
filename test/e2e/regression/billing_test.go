package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"testing"
	"time"

	"github.com/raystack/frontier/billing/credit"

	"github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/billing/usage"

	"github.com/google/uuid"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/core/authenticate"
	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/pkg/server"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type BillingRegressionTestSuite struct {
	suite.Suite
	testBench   *testbench.TestBench
	adminCookie string
}

func (s *BillingRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	connectPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	appConfig := &config.Frontier{
		Log: logger.Config{
			Level:       "error",
			AuditEvents: "db",
		},
		App: server.Config{
			Host:    "localhost",
			Port:    apiPort,
			Connect: server.ConnectConfig{Port: connectPort},
			GRPC: server.GRPCConfig{
				Port:           grpcPort,
				MaxRecvMsgSize: 2 << 10,
				MaxSendMsgSize: 2 << 10,
			},
			ResourcesConfigPath: path.Join(testDataPath, "resource"),
			Authentication: authenticate.Config{
				Session: authenticate.SessionConfig{
					HashSecretKey:  "hash-secret-should-be-32-chars--",
					BlockSecretKey: "hash-secret-should-be-32-chars--",
					Validity:       time.Hour,
				},
				MailOTP: authenticate.MailOTPConfig{
					Subject:  "{{.Otp}}",
					Body:     "{{.Otp}}",
					Validity: 10 * time.Minute,
				},
				TestUsers: testusers.Config{Enabled: true, Domain: "raystack.org", OTP: testbench.TestOTP},
			},
		},
		Billing: billing.Config{
			StripeKey:       "sk_test_mock",
			PlansPath:       path.Join(testDataPath, "plans"),
			DefaultCurrency: "usd",
			AccountConfig: billing.AccountConfig{
				AutoCreateWithOrg:                true,
				OnboardCreditsWithOrg:            200,
				CreditOverdraftProduct:           "support_credits",
				CreditOverdraftInvoiceDay:        1,
				CreditOverdraftInvoiceRangeShift: 1, // shift to future to process current month
			},
		},
	}

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	ctx := context.Background()

	adminCookie, err := testbench.AuthenticateUser(ctx, s.testBench.Client, testbench.OrgAdminEmail)
	s.Require().NoError(err)
	s.adminCookie = adminCookie

	s.Require().NoError(testbench.BootstrapUsers(ctx, s.testBench.Client, adminCookie))
	s.Require().NoError(testbench.BootstrapOrganizations(ctx, s.testBench.Client, adminCookie))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.testBench.Client, adminCookie))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.testBench.Client, adminCookie))
}

func (s *BillingRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *BillingRegressionTestSuite) TestBillingCustomerAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	s.Run("1. creating multiple active billing account shouldn't be allowed", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name:  "org-billing-customer-1",
				Title: "Org Billing Customer 1",
			},
		}))
		s.Assert().NoError(err)

		// creating an org should have already created one billing account
		var billingAccounts []*frontierv1beta1.BillingAccount
		s.Assert().Eventually(func() bool {
			// wait for billing account to be created
			listCustomersResp, err := s.testBench.Client.ListBillingAccounts(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListBillingAccountsRequest{
				OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			}))
			s.Assert().NoError(err)
			billingAccounts = listCustomersResp.Msg.GetBillingAccounts()
			return len(billingAccounts) > 0
		}, 2*time.Second, time.Millisecond*20)

		// creating another billing account shouldn't be allowed
		_, err = s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Body: &frontierv1beta1.BillingAccountRequestBody{
				Email:    "test@example.com",
				Currency: "usd",
				Phone:    "1234567890",
				Name:     "Test Customer",
				Address: &frontierv1beta1.BillingAccount_Address{
					Line1: "123 Main St",
					City:  "San Francisco",
					State: "CA",
				},
			},
		}))
		s.Assert().ErrorContains(err, "active account already exists")
	})
	s.Run("2. create and fetch billing customers successfully", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name:  "org-billing-customer-2",
				Title: "Org Billing Customer 2",
			},
		}))
		s.Assert().NoError(err)
		s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

		createCustomerResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Body: &frontierv1beta1.BillingAccountRequestBody{
				Email:    "test@example.com",
				Currency: "usd",
				Phone:    "1234567890",
				Name:     "Test Customer",
				Address: &frontierv1beta1.BillingAccount_Address{
					Line1: "123 Main St",
					City:  "San Francisco",
					State: "CA",
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createCustomerResp)

		getCustomerResp, err := s.testBench.Client.GetBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
			OrgId:  createOrgResp.Msg.GetOrganization().GetId(),
			Id:     createCustomerResp.Msg.GetBillingAccount().GetId(),
			Expand: []string{"organization"},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getCustomerResp)
		s.Assert().Equal(createCustomerResp.Msg.GetBillingAccount().GetId(), getCustomerResp.Msg.GetBillingAccount().GetId())
		s.Assert().Equal(createCustomerResp.Msg.GetBillingAccount().GetEmail(), getCustomerResp.Msg.GetBillingAccount().GetEmail())
		s.Assert().Equal(createOrgResp.Msg.GetOrganization().GetId(), getCustomerResp.Msg.GetBillingAccount().GetOrganization().GetId())
	})
	s.Run("3. update billing customer successfully", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name:  "org-billing-customer-3",
				Title: "Org Billing Customer 3",
			},
		}))
		s.Assert().NoError(err)
		s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

		createCustomerResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Body: &frontierv1beta1.BillingAccountRequestBody{
				Email:    "test@example2.com",
				Currency: "usd",
				Name:     "Test Customer",
				Address: &frontierv1beta1.BillingAccount_Address{
					State: "CA",
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createCustomerResp)

		// update customer
		updateCustomerResp, err := s.testBench.Client.UpdateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.UpdateBillingAccountRequest{
			Id:    createCustomerResp.Msg.GetBillingAccount().GetId(),
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Body: &frontierv1beta1.BillingAccountRequestBody{
				Email:    "test@example2.com",
				Currency: "usd",
				Phone:    "1234567890",
				Name:     "Test Customer 2",
				Address: &frontierv1beta1.BillingAccount_Address{
					Line1: "123 Main St",
					City:  "San Francisco",
					State: "CA",
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(updateCustomerResp)
		s.Assert().Equal("1234567890", updateCustomerResp.Msg.GetBillingAccount().GetPhone())
		s.Assert().Equal("123 Main St", updateCustomerResp.Msg.GetBillingAccount().GetAddress().GetLine1())
		s.Assert().Equal("San Francisco", updateCustomerResp.Msg.GetBillingAccount().GetAddress().GetCity())
	})
	s.Run("4. create and fetch billing customers successfully with tax data", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name:  "org-billing-customer-4",
				Title: "Org Billing Customer 4",
			},
		}))
		s.Assert().NoError(err)
		s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

		createCustomerResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Body: &frontierv1beta1.BillingAccountRequestBody{
				Email:    "test@example.com",
				Currency: "usd",
				Phone:    "1234567890",
				Name:     "Test Customer",
				Address: &frontierv1beta1.BillingAccount_Address{
					Line1: "123 Main St",
					City:  "San Francisco",
					State: "CA",
				},
				TaxData: []*frontierv1beta1.BillingAccount_Tax{
					{
						Type: "us_ein",
						Id:   "1234567890",
					},
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createCustomerResp)

		getCustomerResp, err := s.testBench.Client.GetBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
			OrgId:  createOrgResp.Msg.GetOrganization().GetId(),
			Id:     createCustomerResp.Msg.GetBillingAccount().GetId(),
			Expand: []string{"organization"},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getCustomerResp)
		s.Assert().Equal(createCustomerResp.Msg.GetBillingAccount().GetId(), getCustomerResp.Msg.GetBillingAccount().GetId())
		s.Assert().Equal(createCustomerResp.Msg.GetBillingAccount().GetEmail(), getCustomerResp.Msg.GetBillingAccount().GetEmail())
		s.Assert().Equal(createOrgResp.Msg.GetOrganization().GetId(), getCustomerResp.Msg.GetBillingAccount().GetOrganization().GetId())
		s.Assert().Equal(1, len(getCustomerResp.Msg.GetBillingAccount().GetTaxData()))
		s.Assert().Equal("us_ein", getCustomerResp.Msg.GetBillingAccount().GetTaxData()[0].GetType())
		s.Assert().Equal("1234567890", getCustomerResp.Msg.GetBillingAccount().GetTaxData()[0].GetId())
	})
	s.Run("5. onboarding credits should be auto credited in org billing account", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name:  "org-billing-customer-5",
				Title: "Org Billing Customer 5",
			},
		}))
		s.Assert().NoError(err)
		s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

		var customerID string
		s.Assert().Eventually(func() bool {
			listCustomerAccountResp, err := s.testBench.Client.ListBillingAccounts(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListBillingAccountsRequest{
				OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			}))
			s.Assert().NoError(err)
			s.Assert().NotNil(listCustomerAccountResp)
			if len(listCustomerAccountResp.Msg.GetBillingAccounts()) > 0 {
				customerID = listCustomerAccountResp.Msg.GetBillingAccounts()[0].GetId()
				return true
			}
			return false
		}, time.Second*2, time.Millisecond*50)

		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			Id:    customerID,
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getBalanceResp)
		s.Assert().Equal(int64(200), getBalanceResp.Msg.GetBalance().GetAmount())
	})
}

func (s *BillingRegressionTestSuite) TestPlansAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	s.Run("1. fetch existing plans successfully", func() {
		listPlansResp, err := s.testBench.Client.ListPlans(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListPlansRequest{}))
		s.Assert().NoError(err)
		s.Assert().NotNil(listPlansResp)
		s.Assert().NotEmpty(listPlansResp.Msg.GetPlans())
	})
	s.Run("2. create a plan successfully", func() {
		createPlanResp, err := s.testBench.Client.CreatePlan(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
			Body: &frontierv1beta1.PlanRequestBody{
				Name:        "test-plan-2",
				Title:       "Test Plan 2",
				Description: "Test Plan 2",
				Interval:    "month",
				Products: []*frontierv1beta1.Product{
					{
						Name:        "test-plan-product-2",
						Title:       "Test Product 2",
						Description: "Test Product 2",
						Prices: []*frontierv1beta1.Price{
							{
								Currency: "usd",
								Amount:   100,
								Interval: "month",
							},
						},
					},
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createPlanResp)
		s.Assert().NotNil(createPlanResp.Msg.GetPlan().GetProducts())

		getPlanResp, err := s.testBench.Client.GetPlan(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetPlanRequest{
			Id: createPlanResp.Msg.GetPlan().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getPlanResp)
		s.Assert().Equal(createPlanResp.Msg.GetPlan().GetId(), getPlanResp.Msg.GetPlan().GetId())
		s.Assert().Equal(createPlanResp.Msg.GetPlan().GetProducts(), getPlanResp.Msg.GetPlan().GetProducts())
	})
}

func (s *BillingRegressionTestSuite) TestProductsAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	s.Run("1. create a credit buying product successfully", func() {
		createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "test-product",
				Title:       "Test Product",
				Description: "Test Product",
				PlanId:      "",
				Prices: []*frontierv1beta1.Price{
					{
						Currency: "usd",
						Amount:   100,
						Interval: "month",
					},
				},
				Features: []*frontierv1beta1.Feature{
					{
						Name: "test-feature",
					},
				},
				BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
					CreditAmount: 400,
					MinQuantity:  2,
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createProductResp)
		s.Assert().NotNil(createProductResp.Msg.GetProduct().GetPrices())

		getProductResp, err := s.testBench.Client.GetProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetProductRequest{
			Id: createProductResp.Msg.GetProduct().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getProductResp)
		s.Assert().Equal(createProductResp.Msg.GetProduct().GetId(), getProductResp.Msg.GetProduct().GetId())
		s.Assert().Equal(createProductResp.Msg.GetProduct().GetPrices(), getProductResp.Msg.GetProduct().GetPrices())
		s.Assert().Equal(createProductResp.Msg.GetProduct().GetFeatures(), getProductResp.Msg.GetProduct().GetFeatures())
		s.Assert().Len(getProductResp.Msg.GetProduct().GetFeatures(), 1)
		s.Assert().Equal(int64(2), getProductResp.Msg.GetProduct().GetBehaviorConfig().GetMinQuantity())
	})
	s.Run("2. Update a product successfully", func() {
		createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "test-product-2",
				Title:       "Test Product-2",
				Description: "Test Product-2",
				PlanId:      "",
				Prices: []*frontierv1beta1.Price{
					{
						Currency: "usd",
						Amount:   100,
						Interval: "month",
					},
				},
				Features: []*frontierv1beta1.Feature{
					{
						Name: "test-feature",
					},
				},
				BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
					CreditAmount: 400,
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createProductResp)
		s.Assert().NotNil(createProductResp.Msg.GetProduct().GetPrices())

		// add additional feature and remove existing feature
		updateProductResp, err := s.testBench.Client.UpdateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.UpdateProductRequest{
			Id: createProductResp.Msg.GetProduct().GetId(),
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "test-product-2",
				Title:       "Test Product-2",
				Description: "Test Product-2",
				PlanId:      "",
				Features: []*frontierv1beta1.Feature{
					{
						Name: "test-feature-2",
					},
				},
				BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
					CreditAmount: 400,
					MaxQuantity:  20,
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(updateProductResp)
		s.Assert().Equal(updateProductResp.Msg.GetProduct().GetId(), createProductResp.Msg.GetProduct().GetId())
		s.Assert().Equal(updateProductResp.Msg.GetProduct().GetPrices(), createProductResp.Msg.GetProduct().GetPrices())
		s.Assert().Equal(1, len(updateProductResp.Msg.GetProduct().GetFeatures()))
		s.Assert().Equal("test-feature-2", updateProductResp.Msg.GetProduct().GetFeatures()[0].GetName())
		s.Assert().Equal(int64(400), updateProductResp.Msg.GetProduct().GetBehaviorConfig().GetCreditAmount())
		s.Assert().Equal(int64(20), updateProductResp.Msg.GetProduct().GetBehaviorConfig().GetMaxQuantity())
	})
	s.Run("create a feature in existing product successfully", func() {
		createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "test-product-3",
				Title:       "Test Product-3",
				Description: "Test Product-3",
				PlanId:      "",
				Prices: []*frontierv1beta1.Price{
					{
						Currency: "usd",
						Amount:   100,
						Interval: "month",
					},
				},
				Features: []*frontierv1beta1.Feature{
					{
						Name: "test-feature",
					},
				},
				BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
					CreditAmount: 400,
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createProductResp)
		s.Assert().NotNil(createProductResp.Msg.GetProduct().GetPrices())

		// add additional feature
		createFeatureResp, err := s.testBench.Client.CreateFeature(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateFeatureRequest{
			Body: &frontierv1beta1.FeatureRequestBody{
				Name:       "test-feature-3",
				Title:      "Test Feature-3",
				ProductIds: []string{createProductResp.Msg.GetProduct().GetId()},
				Metadata: Must(structpb.NewStruct(map[string]interface{}{
					"key": "value",
				})),
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createFeatureResp)
		s.Assert().Equal("test-feature-3", createFeatureResp.Msg.GetFeature().GetName())
		s.Assert().Equal("Test Feature-3", createFeatureResp.Msg.GetFeature().GetTitle())
		s.Assert().Equal(1, len(createFeatureResp.Msg.GetFeature().GetProductIds()))
	})
}

func (s *BillingRegressionTestSuite) TestCheckoutAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	// create dummy org
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-checkout-1",
		},
	}))
	s.Assert().NoError(err)
	s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

	// create dummy billing customer
	createBillingResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Body: &frontierv1beta1.BillingAccountRequestBody{
			Email:    "test@frontier-example.com",
			Currency: "usd",
			Phone:    "1234567890",
			Name:     "Test Customer",
			Address: &frontierv1beta1.BillingAccount_Address{
				Line1: "123 Main St",
				City:  "San Francisco",
				State: "CA",
			},
		},
	}))
	s.Assert().NoError(err)

	s.Run("1. checkout the credit product to buy some credits", func() {
		createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "store-credits",
				Title:       "Store Credits",
				Description: "Store Credits",
				PlanId:      "",
				Prices: []*frontierv1beta1.Price{
					{
						Currency: "usd",
						Amount:   100,
						Interval: "month",
					},
				},
				BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
					CreditAmount: 400,
				},
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createProductResp)

		checkoutResp, err := s.testBench.Client.CreateCheckout(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
			OrgId:      createOrgResp.Msg.GetOrganization().GetId(),
			BillingId:  createBillingResp.Msg.GetBillingAccount().GetId(),
			SuccessUrl: "https://example.com/success?checkout_id={{.CheckoutID}}",
			CancelUrl:  "https://example.com/cancel",
			ProductBody: &frontierv1beta1.CheckoutProductBody{
				Product: createProductResp.Msg.GetProduct().GetId(),
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(checkoutResp)
		s.Assert().NotEmpty(checkoutResp.Msg.GetCheckoutSession().GetCheckoutUrl())
		s.Assert().Equal("https://example.com/success?checkout_id="+checkoutResp.Msg.GetCheckoutSession().GetId(), checkoutResp.Msg.GetCheckoutSession().GetSuccessUrl())
		s.Assert().Equal("https://example.com/cancel", checkoutResp.Msg.GetCheckoutSession().GetCancelUrl())

		listCheckout, err := s.testBench.Client.ListCheckouts(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListCheckoutsRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(listCheckout)
		// we can't really pay the checkout session in test so automatic credit update won't happen
	})
	s.Run("2. checkout the subscription for a plan", func() {
		checkoutResp, err := s.testBench.Client.CreateCheckout(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
			OrgId:      createOrgResp.Msg.GetOrganization().GetId(),
			BillingId:  createBillingResp.Msg.GetBillingAccount().GetId(),
			SuccessUrl: "https://example.com/success",
			CancelUrl:  "https://example.com/cancel",
			SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
				Plan: "enterprise_yearly",
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(checkoutResp)
		s.Assert().NotEmpty(checkoutResp.Msg.GetCheckoutSession().GetCheckoutUrl())

		listCheckout, err := s.testBench.Client.ListCheckouts(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListCheckoutsRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(listCheckout)
		// we can't really pay the checkout session in test so automatic credit update won't happen
	})
	s.Run("3. delegate checkout the credits product", func() {
		createProduct, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "store-credits-checkout-1",
				Behavior:    "credits",
				Title:       "Store Credits",
				Description: "Store Credits",
				BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
					CreditAmount: 400,
				},
			},
		}))
		s.Assert().NoError(err)

		delegateCheckoutResp, err := s.testBench.AdminClient.DelegatedCheckout(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.DelegatedCheckoutRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			ProductBody: &frontierv1beta1.CheckoutProductBody{
				Product:  createProduct.Msg.GetProduct().GetId(),
				Quantity: 2,
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(delegateCheckoutResp)
		s.Assert().NotEmpty(delegateCheckoutResp.Msg.GetProduct())

		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(int64(800), getBalanceResp.Msg.GetBalance().GetAmount())
	})
}

func (s *BillingRegressionTestSuite) TestUsageAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	// create dummy org
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-usage-1",
		},
	}))
	s.Assert().NoError(err)

	creteProjectResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Name:  "project-usage-1",
			Title: "Project Usage 1",
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		},
	}))
	s.Assert().NoError(err)
	s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

	// create dummy billing customer
	createBillingResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Body: &frontierv1beta1.BillingAccountRequestBody{
			Email:    "test@frontier-example.com",
			Currency: "usd",
			Phone:    "1234567890",
			Name:     "Test Customer",
			Address: &frontierv1beta1.BillingAccount_Address{
				Line1: "123 Main St",
				City:  "San Francisco",
				State: "CA",
			},
		},
	}))
	s.Assert().NoError(err)

	// create a product with credit behavior
	createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProductRequest{
		Body: &frontierv1beta1.ProductRequestBody{
			Name:        "store-credits-usage",
			Title:       "Store Credits",
			Description: "Store Credits",
			PlanId:      "",
			Prices: []*frontierv1beta1.Price{
				{
					Currency: "usd",
					Amount:   100,
					Interval: "month",
				},
			},
			BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
				CreditAmount: 400,
			},
		},
	}))
	s.Assert().NoError(err)
	testUserID := uuid.New().String()

	s.Run("1. report usage to an account having no credits", func() {
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          uuid.New().String(),
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().Error(err)
		s.Assert().ErrorContains(err, "insufficient credits")

		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          uuid.New().String(),
					Source:      "billing.test",
					Description: "billing test",
					Amount:      -20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().Error(err)
	})
	s.Run("2. report usage to an account having some credits", func() {
		_, err = s.testBench.AdminClient.DelegatedCheckout(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.DelegatedCheckoutRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			ProductBody: &frontierv1beta1.CheckoutProductBody{
				Product: createProductResp.Msg.GetProduct().GetId(),
			},
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          uuid.New().String(),
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance-20, getBalanceResp.Msg.GetBalance().GetAmount())
	})
	s.Run("3. revert partial reported usage to an account", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    10,
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance-10, getBalanceResp.Msg.GetBalance().GetAmount())
	})
	s.Run("4. revert full reported usage to an account", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    20,
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance, getBalanceResp.Msg.GetBalance().GetAmount())
	})
	s.Run("5. revert more than full reported usage to an account should fail", func() {
		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    30,
		}))
		s.Assert().ErrorContains(err, usage.ErrRevertAmountExceeds.Error())
	})
	s.Run("6. revert reported usage multiple times should pass", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    5,
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    8,
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance-7, getBalanceResp.Msg.GetBalance().GetAmount())
	})
	s.Run("7. revert reported usage multiple times more than original amount should fail", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    15,
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    8,
		}))
		s.Assert().Error(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance-5, getBalanceResp.Msg.GetBalance().GetAmount())
	})
	s.Run("8. reverting a revert usage should fail", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      20,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   usageID,
			Amount:    10,
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance-10, getBalanceResp.Msg.GetBalance().GetAmount())

		listTransactions, err := s.testBench.Client.ListBillingTransactions(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		lastRevertID := listTransactions.Msg.GetTransactions()[0].GetId()

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			UsageId:   lastRevertID,
			Amount:    10,
		}))
		s.Assert().ErrorContains(err, usage.ErrExistingRevertedUsage.Error())
	})
	s.Run("9. revert full reported usage to an account using project id", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		usageID := uuid.New().String()
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			ProjectId: creteProjectResp.Msg.GetProject().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      5,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			ProjectId: creteProjectResp.Msg.GetProject().GetId(),
			UsageId:   usageID,
			Amount:    5,
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(beforeBalance, getBalanceResp.Msg.GetBalance().GetAmount())
	})
	s.Run("10. allow customer overdraft if set", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		// set limit to -20
		_, err = s.testBench.AdminClient.UpdateBillingAccountLimits(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.UpdateBillingAccountLimitsRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			Id:        createBillingResp.Msg.GetBillingAccount().GetId(),
			CreditMin: -20,
		}))
		s.Assert().NoError(err)

		usageID := uuid.New().String()
		// go overdraft
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			ProjectId: creteProjectResp.Msg.GetProject().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          usageID,
					Source:      "billing.test",
					Description: "billing test",
					Amount:      beforeBalance + 10,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		// check balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(int64(-10), getBalanceResp.Msg.GetBalance().GetAmount())

		// can't go over overdraft
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			ProjectId: creteProjectResp.Msg.GetProject().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          uuid.NewString(),
					Source:      "billing.test",
					Description: "billing test",
					Amount:      50,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().ErrorContains(err, credit.ErrInsufficientCredits.Error())

		// revert usage
		_, err = s.testBench.AdminClient.RevertBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.RevertBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			ProjectId: creteProjectResp.Msg.GetProject().GetId(),
			UsageId:   usageID,
			Amount:    beforeBalance + 10,
		}))
		s.Assert().NoError(err)

		// reset limit
		_, err = s.testBench.AdminClient.UpdateBillingAccountLimits(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.UpdateBillingAccountLimitsRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			Id:        createBillingResp.Msg.GetBillingAccount().GetId(),
			CreditMin: 0,
		}))
		s.Assert().NoError(err)
	})
	s.Run("11. check for concurrent transactions", func() {
		// check initial balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()

		// Create multiple concurrent usage requests
		numRequests := 20
		errChan := make(chan error, numRequests)
		for i := 0; i < numRequests; i++ {
			go func() {
				_, err := s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
					OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
					BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
					Usages: []*frontierv1beta1.Usage{
						{
							Id:     uuid.New().String(),
							Source: "billing.test",
							Amount: 2,
							UserId: testUserID,
						},
					},
				}))
				errChan <- err
			}()
		}

		// Wait for all requests to complete
		var successCount int
		for i := 0; i < numRequests; i++ {
			err := <-errChan
			if err == nil {
				successCount++
			} else {
				s.Assert().ErrorContains(err, credit.ErrInsufficientCredits.Error())
			}
		}

		// Verify final balance
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)

		// Verify the balance was deducted exactly by successful transactions amount
		expectedBalance := beforeBalance - int64(successCount*2)
		s.Assert().Equal(expectedBalance, getBalanceResp.Msg.GetBalance().GetAmount())
	})
}

func (s *BillingRegressionTestSuite) TestInvoiceAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	// create dummy org
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-invoice-1",
		},
	}))
	s.Assert().NoError(err)

	creteProjectResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Name:  "project-invoice-1",
			Title: "Project Usage 1",
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		},
	}))
	s.Assert().NoError(err)
	s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

	// create dummy billing customer
	createBillingResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Body: &frontierv1beta1.BillingAccountRequestBody{
			Email:    "test@frontier-example.com",
			Currency: "usd",
			Phone:    "1234567890",
			Name:     "Test Customer",
			Address: &frontierv1beta1.BillingAccount_Address{
				Line1: "123 Main St",
				City:  "San Francisco",
				State: "CA",
			},
		},
	}))
	s.Assert().NoError(err)

	// set limit for overdraft
	_, err = s.testBench.AdminClient.UpdateBillingAccountDetails(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.UpdateBillingAccountDetailsRequest{
		OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
		Id:        createBillingResp.Msg.GetBillingAccount().GetId(),
		CreditMin: -500,
		DueInDays: 0,
	}))
	s.Assert().NoError(err)

	testUserID := uuid.New().String()
	s.Run("1. generate invoice for overdraft credits on demand", func() {
		// check balance
		getBalanceResp, err := s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		beforeBalance := getBalanceResp.Msg.GetBalance().GetAmount()
		s.Assert().Equal(int64(0), beforeBalance)

		// go overdraft
		_, err = s.testBench.Client.CreateBillingUsage(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			ProjectId: creteProjectResp.Msg.GetProject().GetId(),
			Usages: []*frontierv1beta1.Usage{
				{
					Id:          uuid.New().String(),
					Source:      "billing.test",
					Description: "billing test",
					Amount:      30,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
				{
					Id:          uuid.New().String(),
					Source:      "billing.test",
					Description: "billing test",
					Amount:      50,
					UserId:      testUserID,
					Metadata: Must(structpb.NewStruct(map[string]interface{}{
						"key": "value",
					})),
				},
			},
		}))
		s.Assert().NoError(err)

		// check balance after overdraft
		getBalanceResp, err = s.testBench.Client.GetBillingBalance(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GetBillingBalanceRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
			Id:    createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(int64(-80), getBalanceResp.Msg.GetBalance().GetAmount())

		// trigger invoice generation
		_, err = s.testBench.AdminClient.GenerateInvoices(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.GenerateInvoicesRequest{}))
		s.Assert().NoError(err)

		// check created invoices for the customer
		listInvoicesResp, err := s.testBench.Client.ListInvoices(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().Len(listInvoicesResp.Msg.GetInvoices(), 1)
		inv := listInvoicesResp.Msg.GetInvoices()[0]
		s.Assert().Equal(createBillingResp.Msg.GetBillingAccount().GetId(), inv.GetCustomerId())
		s.Assert().Equal("usd", inv.GetCurrency())
		// can't assert amount as it's calculated based on usage and plan
		// can't test re-triggering as stripe mock doesn't return current line items
	})
}

func (s *BillingRegressionTestSuite) TestCheckFeatureEntitlementAPI() {
	ctxOrgAdminAuth := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	// create dummy org
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-entitlement-1",
		},
	}))
	s.Assert().NoError(err)

	// create dummy project
	_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Name:  "project-entitlement-1",
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		},
	}))
	s.Assert().NoError(err)
	s.disableExistingBillingAccounts(ctxOrgAdminAuth, createOrgResp.Msg.GetOrganization().GetId())

	createPlanResp, err := s.testBench.Client.CreatePlan(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
		Body: &frontierv1beta1.PlanRequestBody{
			Name:        "test-plan-entitlement-1",
			Title:       "Test Plan 1",
			Description: "Test Plan 1",
			Interval:    "month",
			Products: []*frontierv1beta1.Product{
				{
					Name:        "test-plan-product-2",
					Title:       "Test Product 2",
					Description: "Test Product 2",
					Prices: []*frontierv1beta1.Price{
						{
							Currency: "usd",
							Amount:   100,
							Interval: "month",
						},
					},
					Features: []*frontierv1beta1.Feature{
						{
							Name: "test-feature-entitlement-1",
						},
					},
				},
			},
		},
	}))
	s.Assert().NoError(err)

	// create dummy billing customer
	createBillingResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CreateBillingAccountRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Body: &frontierv1beta1.BillingAccountRequestBody{
			Email:    "test@frontier-example.com",
			Currency: "usd",
			Phone:    "1234567890",
			Name:     "Test Customer",
			Address: &frontierv1beta1.BillingAccount_Address{
				Line1: "123 Main St",
				City:  "San Francisco",
				State: "CA",
			},
		},
	}))
	s.Assert().NoError(err)

	s.Run("1. should return a org is not entitled to feature if not subscribed", func() {
		status, err := s.testBench.Client.CheckFeatureEntitlement(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			Feature:   "test-feature-entitlement-1",
		}))
		s.Assert().NoError(err)
		s.Assert().False(status.Msg.GetStatus())
	})
	s.Run("2. should return the org is entitled to feature correctly", func() {
		_, err = s.testBench.AdminClient.DelegatedCheckout(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.DelegatedCheckoutRequest{
			OrgId:     createOrgResp.Msg.GetOrganization().GetId(),
			BillingId: createBillingResp.Msg.GetBillingAccount().GetId(),
			SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
				Plan: createPlanResp.Msg.GetPlan().GetId(),
			},
		}))
		s.Assert().NoError(err)

		status, err := s.testBench.Client.CheckFeatureEntitlement(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
			OrgId:   createOrgResp.Msg.GetOrganization().GetId(),
			Feature: "test-feature-entitlement-1",
		}))
		s.Assert().NoError(err)
		s.Assert().True(status.Msg.GetStatus())

		// should also work with org_id directly
		status, err = s.testBench.Client.CheckFeatureEntitlement(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
			OrgId:   createOrgResp.Msg.GetOrganization().GetId(),
			Feature: "test-feature-entitlement-1",
		}))
		s.Assert().NoError(err)
		s.Assert().True(status.Msg.GetStatus())
	})
}

func (s *BillingRegressionTestSuite) TestBillingWebhookCallbackAPI() {
	ctxStripeHeader := testbench.ContextWithHeaders(context.Background(), map[string]string{
		"Stripe-Signature": "invalid-signature",
	})
	s.Run("1. shouldn fail to accept a webhook with invalid signatures", func() {
		stripeEvent := stripe.Event{}
		eventBytes, err := json.Marshal(stripeEvent)
		s.Assert().NoError(err)
		_, err = s.testBench.Client.BillingWebhookCallback(ctxStripeHeader, connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
			Provider: "stripe",
			Body:     eventBytes,
		}))
		s.Assert().Error(err)
		s.Assert().Equal(connect.CodeInternal, connect.CodeOf(err))
	})
}

func TestEndToEndBillingRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(BillingRegressionTestSuite))
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func (s *BillingRegressionTestSuite) disableExistingBillingAccounts(ctxOrgAdminAuth context.Context, orgID string) {
	s.T().Helper()

	var billingAccounts []*frontierv1beta1.BillingAccount
	// wait for billing account to be created
	s.Assert().Eventually(func() bool {
		listCustomersResp, err := s.testBench.Client.ListBillingAccounts(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.ListBillingAccountsRequest{
			OrgId: orgID,
		}))
		s.Assert().NoError(err)
		billingAccounts = listCustomersResp.Msg.GetBillingAccounts()
		return len(billingAccounts) > 0
	}, 2*time.Second, time.Millisecond*20)

	// disable existing billing account
	for _, billingAccount := range billingAccounts {
		_, err := s.testBench.Client.DisableBillingAccount(ctxOrgAdminAuth, connect.NewRequest(&frontierv1beta1.DisableBillingAccountRequest{
			OrgId: orgID,
			Id:    billingAccount.GetId(),
		}))
		s.Assert().NoError(err)
	}
}
