package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stretchr/testify/assert"

	"github.com/raystack/frontier/billing/customer"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type BillingCustomerRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.BillingCustomerRepository
	orgIDs     []string
}

func (s *BillingCustomerRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewBillingCustomerRepository(s.client)

	orgs, err := bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.orgIDs = utils.Map(orgs, func(org organization.Organization) string {
		return org.ID
	})
}

func (s *BillingCustomerRepositoryTestSuite) SetupTest() {
}

func (s *BillingCustomerRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingCustomerRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingCustomerRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_BILLING_CUSTOMERS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *BillingCustomerRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description string
		Customer    customer.Customer
		Expected    customer.Customer
		ErrString   string
	}

	sampleID1 := uuid.New().String()
	sampleID2 := uuid.New().String()
	var testCases = []testCase{
		{
			Description: "should create a basic customer with provider successfully",
			Customer: customer.Customer{
				ID:         sampleID1,
				ProviderID: sampleID1,
				OrgID:      s.orgIDs[0],
				Name:       "customer 1",
				TaxData: []customer.Tax{
					{
						Type: "t1",
						ID:   "i1",
					},
				},
				Address: customer.Address{
					City: "city",
				},
				Email:     "email",
				State:     "",
				Metadata:  metadata.Metadata{},
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: nil,
			},
			Expected: customer.Customer{
				Name:       "customer 1",
				ProviderID: sampleID1,
				OrgID:      s.orgIDs[0],
				State:      customer.ActiveState,
				TaxData: []customer.Tax{
					{
						Type: "t1",
						ID:   "i1",
					},
				},
				Address: customer.Address{
					City: "city",
				},
				Email:    "email",
				Metadata: metadata.Metadata{},
			},
		},
		{
			Description: "should create a customer without provider successfully",
			Customer: customer.Customer{
				ID:         sampleID2,
				ProviderID: "",
				OrgID:      s.orgIDs[0],
				Name:       "new_product2",
				Currency:   "usd",
				State:      "",
				Metadata:   metadata.Metadata{},
				CreatedAt:  time.Time{},
				UpdatedAt:  time.Time{},
				DeletedAt:  nil,
			},
			Expected: customer.Customer{
				Name:       "new_product2",
				ProviderID: "",
				OrgID:      s.orgIDs[0],
				Currency:   "usd",
				State:      customer.ActiveState,
				Metadata:   metadata.Metadata{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.Customer)
			if err != nil {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(customer.Customer{}, "ID", "CreatedAt", "UpdatedAt")); diff != "" {
				s.T().Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func (s *BillingCustomerRepositoryTestSuite) TestList() {
	type testCase struct {
		Description string
		Expected    []customer.Customer
		ErrString   string
		filter      customer.Filter
	}

	sampleID1 := uuid.New().String()
	sampleID2 := uuid.New().String()
	sampleID3 := uuid.New().String()
	sampleID4 := uuid.New().String()
	customers := []struct {
		Customer customer.Customer
		Details  customer.Details
	}{
		{
			Customer: customer.Customer{
				ID:         sampleID1,
				ProviderID: sampleID1,
				OrgID:      s.orgIDs[0],
				Name:       "customer 1",
				TaxData: []customer.Tax{
					{
						Type: "t1",
						ID:   "i1",
					},
				},
				Address: customer.Address{
					City: "city",
				},
				Email:     "email",
				State:     "active",
				Metadata:  metadata.Metadata{},
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: nil,
			},
		},
		{
			Customer: customer.Customer{
				ID:         sampleID2,
				ProviderID: sampleID2,
				OrgID:      s.orgIDs[1],
				Name:       "customer 2",
				TaxData: []customer.Tax{
					{
						Type: "t1",
						ID:   "i1",
					},
				},
				Address: customer.Address{
					City: "city",
				},
				Email:     "email",
				State:     "",
				Metadata:  metadata.Metadata{},
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: nil,
			},
		},
		{
			Customer: customer.Customer{
				ID:        sampleID3,
				OrgID:     s.orgIDs[0],
				Name:      "customer 3",
				Email:     "email",
				State:     "active",
				Metadata:  metadata.Metadata{},
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: nil,
			},
		},
		{
			Customer: customer.Customer{
				ID:         sampleID4,
				ProviderID: sampleID4,
				OrgID:      s.orgIDs[0],
				Name:       "customer 4",
				Email:      "email",
				State:      "active",
				Metadata:   metadata.Metadata{},
				CreatedAt:  time.Time{},
				UpdatedAt:  time.Time{},
				DeletedAt:  nil,
			},
			Details: customer.Details{
				CreditMin: -200,
				DueInDays: 0,
			},
		},
	}

	var testCases = []testCase{
		{
			Description: "should create basic customer with provider successfully",
			Expected: []customer.Customer{
				customers[0].Customer,
				customers[3].Customer,
			},
			filter: customer.Filter{
				OrgID:  s.orgIDs[0],
				State:  customer.ActiveState,
				Online: utils.Bool(true),
			},
		},
		{
			Description: "should list customers with overdraft limits",
			Expected: []customer.Customer{
				customers[3].Customer,
			},
			filter: customer.Filter{
				OrgID:            s.orgIDs[0],
				State:            customer.ActiveState,
				Online:           utils.Bool(true),
				AllowedOverdraft: utils.Bool(true),
			},
		},
	}

	for _, c := range customers {
		newC, err := s.repository.Create(s.ctx, c.Customer)
		_, errDetails := s.repository.UpdateDetailsByID(s.ctx, newC.ID, c.Details)
		assert.NoError(s.T(), err)
		assert.NoError(s.T(), errDetails)
	}
	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, tc.filter)
			if err != nil {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			slices.SortFunc(got, func(i, j customer.Customer) int {
				return strings.Compare(i.Name, j.Name)
			})
			slices.SortFunc(tc.Expected, func(i, j customer.Customer) int {
				return strings.Compare(i.Name, j.Name)
			})
			if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(customer.Customer{}, "ID", "CreatedAt", "UpdatedAt")); diff != "" {
				s.T().Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBillingCustomerRepository(t *testing.T) {
	suite.Run(t, new(BillingCustomerRepositoryTestSuite))
}
