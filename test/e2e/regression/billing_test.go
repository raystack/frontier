package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/pkg/server"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/metadata"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type BillingRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
}

func (s *BillingRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	appConfig := &config.Frontier{
		Log: logger.Config{
			Level:       "error",
			AuditEvents: "db",
		},
		App: server.Config{
			Host: "localhost",
			Port: apiPort,
			GRPC: server.GRPCConfig{
				Port:           grpcPort,
				MaxRecvMsgSize: 2 << 10,
				MaxSendMsgSize: 2 << 10,
			},
			IdentityProxyHeader: testbench.IdentityHeader,
			ResourcesConfigPath: path.Join(testDataPath, "resource"),
		},
		Billing: billing.Config{
			StripeKey: "sk_test_XXX",
			PlansPath: path.Join(testDataPath, "plans"),
		},
	}

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	ctx := context.Background()

	s.Require().NoError(testbench.BootstrapUsers(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapOrganizations(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.testBench.Client, testbench.OrgAdminEmail))
}

func (s *BillingRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *BillingRegressionTestSuite) TestBillingCustomerAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. create and fetch billing customers successfully", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-billing-customer-1",
		})
		s.Assert().NoError(err)

		createCustomerResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, &frontierv1beta1.CreateBillingAccountRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
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
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createCustomerResp)

		getCustomerResp, err := s.testBench.Client.GetBillingAccount(ctxOrgAdminAuth, &frontierv1beta1.GetBillingAccountRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
			Id:    createCustomerResp.GetBillingAccount().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCustomerResp)
		s.Assert().Equal(createCustomerResp.GetBillingAccount().GetId(), getCustomerResp.GetBillingAccount().GetId())
		s.Assert().Equal(createCustomerResp.GetBillingAccount().GetEmail(), getCustomerResp.GetBillingAccount().GetEmail())
	})
}

func (s *BillingRegressionTestSuite) TestPlansAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. fetch existing plans successfully", func() {
		listPlansResp, err := s.testBench.Client.ListPlans(ctxOrgAdminAuth, &frontierv1beta1.ListPlansRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listPlansResp)
		s.Assert().NotEmpty(listPlansResp.GetPlans())
	})
	s.Run("2. create a plan successfully", func() {
		createPlanResp, err := s.testBench.Client.CreatePlan(ctxOrgAdminAuth, &frontierv1beta1.CreatePlanRequest{
			Body: &frontierv1beta1.PlanRequestBody{
				Name:        "test-plan-2",
				Title:       "Test Plan 2",
				Description: "Test Plan 2",
				Interval:    "month",
				Products: []*frontierv1beta1.Product{
					{
						Name:        "test-product-2",
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
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createPlanResp)
		s.Assert().NotNil(createPlanResp.GetPlan().GetProducts())

		getPlanResp, err := s.testBench.Client.GetPlan(ctxOrgAdminAuth, &frontierv1beta1.GetPlanRequest{
			Id: createPlanResp.GetPlan().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getPlanResp)
		s.Assert().Equal(createPlanResp.GetPlan().GetId(), getPlanResp.GetPlan().GetId())
		s.Assert().Equal(createPlanResp.GetPlan().GetProducts(), getPlanResp.GetPlan().GetProducts())
	})
}

func (s *BillingRegressionTestSuite) TestProductsAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. create a credit buying product successfully", func() {
		createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, &frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "test-product",
				Title:       "Test Product",
				Description: "Test Product",
				PlanId:      "",
				Prices: []*frontierv1beta1.Price{
					{
						Currency: "usd",
						Amount:   100,
					},
				},
				Features: []*frontierv1beta1.Feature{
					{
						Name: "test-feature",
					},
				},
				CreditAmount: 400,
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createProductResp)
		s.Assert().NotNil(createProductResp.GetProduct().GetPrices())

		getProductResp, err := s.testBench.Client.GetProduct(ctxOrgAdminAuth, &frontierv1beta1.GetProductRequest{
			Id: createProductResp.GetProduct().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getProductResp)
		s.Assert().Equal(createProductResp.GetProduct().GetId(), getProductResp.GetProduct().GetId())
		s.Assert().Equal(createProductResp.GetProduct().GetPrices(), getProductResp.GetProduct().GetPrices())
		s.Assert().Equal(createProductResp.GetProduct().GetFeatures(), getProductResp.GetProduct().GetFeatures())
		s.Assert().Len(getProductResp.GetProduct().GetFeatures(), 1)
	})
}

func (s *BillingRegressionTestSuite) TestCheckoutAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// create dummy org
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-checkout-1",
		},
	})
	s.Assert().NoError(err)

	// create dummy billing customer
	createBillingResp, err := s.testBench.Client.CreateBillingAccount(ctxOrgAdminAuth, &frontierv1beta1.CreateBillingAccountRequest{
		OrgId: createOrgResp.GetOrganization().GetId(),
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
	})
	s.Assert().NoError(err)

	s.Run("1. checkout the credit product to buy some credits", func() {
		createProductResp, err := s.testBench.Client.CreateProduct(ctxOrgAdminAuth, &frontierv1beta1.CreateProductRequest{
			Body: &frontierv1beta1.ProductRequestBody{
				Name:        "store-credits",
				Title:       "Store Credits",
				Description: "Store Credits",
				PlanId:      "",
				Prices: []*frontierv1beta1.Price{
					{
						Currency: "usd",
						Amount:   100,
					},
				},
				CreditAmount: 400,
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createProductResp)

		checkoutResp, err := s.testBench.Client.CreateCheckout(ctxOrgAdminAuth, &frontierv1beta1.CreateCheckoutRequest{
			OrgId:      createOrgResp.GetOrganization().GetId(),
			BillingId:  createBillingResp.GetBillingAccount().GetId(),
			SuccessUrl: "https://example.com/success",
			CancelUrl:  "https://example.com/cancel",
			ProductBody: &frontierv1beta1.CheckoutProductBody{
				Product: createProductResp.GetProduct().GetId(),
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkoutResp)
		s.Assert().NotEmpty(checkoutResp.GetCheckoutSession().GetCheckoutUrl())

		listCheckout, err := s.testBench.Client.ListCheckouts(ctxOrgAdminAuth, &frontierv1beta1.ListCheckoutsRequest{
			OrgId:     createOrgResp.GetOrganization().GetId(),
			BillingId: createBillingResp.GetBillingAccount().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(listCheckout)
		// we can't really pay the checkout session in test so automatic credit update won't happen
	})
	s.Run("2. checkout the subscription for a plan", func() {
		checkoutResp, err := s.testBench.Client.CreateCheckout(ctxOrgAdminAuth, &frontierv1beta1.CreateCheckoutRequest{
			OrgId:      createOrgResp.GetOrganization().GetId(),
			BillingId:  createBillingResp.GetBillingAccount().GetId(),
			SuccessUrl: "https://example.com/success",
			CancelUrl:  "https://example.com/cancel",
			SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
				Plan: "enterprise_yearly",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkoutResp)
		s.Assert().NotEmpty(checkoutResp.GetCheckoutSession().GetCheckoutUrl())

		listCheckout, err := s.testBench.Client.ListCheckouts(ctxOrgAdminAuth, &frontierv1beta1.ListCheckoutsRequest{
			OrgId:     createOrgResp.GetOrganization().GetId(),
			BillingId: createBillingResp.GetBillingAccount().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(listCheckout)
		// we can't really pay the checkout session in test so automatic credit update won't happen
	})
}

func TestEndToEndBillingRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(BillingRegressionTestSuite))
}
