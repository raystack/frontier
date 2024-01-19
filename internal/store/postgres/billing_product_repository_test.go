package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/product"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type BillingProductRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.BillingProductRepository
}

func (s *BillingProductRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewBillingProductRepository(s.client)
}

func (s *BillingProductRepositoryTestSuite) SetupTest() {
}

func (s *BillingProductRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingProductRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingProductRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_BILLING_PRODUCTS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *BillingProductRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description string
		Product     product.Product
		Expected    product.Product
		ErrString   string
	}

	sampleID1 := uuid.New().String()
	sampleID2 := uuid.New().String()
	var testCases = []testCase{
		{
			Description: "should create a basic product successfully",
			Product: product.Product{
				ID:          sampleID1,
				ProviderID:  sampleID1,
				PlanIDs:     []string{sampleID1},
				Name:        "new_product1",
				Title:       "p1",
				Description: "d1",
				Behavior:    product.BasicBehavior,
				Config:      product.BehaviorConfig{},
				Prices:      nil,
				Features:    nil,
				State:       "",
				Metadata:    metadata.Metadata{},
				CreatedAt:   time.Time{},
				UpdatedAt:   time.Time{},
				DeletedAt:   nil,
			},
			Expected: product.Product{
				Name:        "new_product1",
				PlanIDs:     []string{sampleID1},
				ProviderID:  sampleID1,
				Title:       "p1",
				Description: "d1",
				Behavior:    product.BasicBehavior,
				Config:      product.BehaviorConfig{},
				State:       "",
				Metadata:    metadata.Metadata{},
			},
		},
		{
			Description: "should create a credit product successfully",
			Product: product.Product{
				ID:          sampleID2,
				ProviderID:  sampleID2,
				PlanIDs:     []string{sampleID2},
				Name:        "new_product2",
				Title:       "p2",
				Description: "d2",
				Behavior:    product.CreditBehavior,
				Config: product.BehaviorConfig{
					CreditAmount: 20,
				},
				Prices:    nil,
				Features:  nil,
				State:     "",
				Metadata:  metadata.Metadata{},
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: nil,
			},
			Expected: product.Product{
				Name:        "new_product2",
				PlanIDs:     []string{sampleID2},
				ProviderID:  sampleID2,
				Title:       "p2",
				Description: "d2",
				Behavior:    product.CreditBehavior,
				Config: product.BehaviorConfig{
					CreditAmount: 20,
				},
				State:    "",
				Metadata: metadata.Metadata{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.Product)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.Expected, cmpopts.IgnoreFields(product.Product{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.Expected)
			}
		})
	}
}

func TestBillingProductRepository(t *testing.T) {
	suite.Run(t, new(BillingProductRepositoryTestSuite))
}
