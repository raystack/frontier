package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

// Runs the billing checkout reads against a real postgres to check that a
// soft-deleted checkout stays out of every read.
type BillingCheckoutRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.BillingCheckoutRepository
}

func (s *BillingCheckoutRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewBillingCheckoutRepository(s.client)
}

func (s *BillingCheckoutRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingCheckoutRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES ('bch-live', 'Live Org')`)
	s.exec(`INSERT INTO billing_customers (org_id, provider_id, name, email)
		VALUES ((SELECT id FROM organizations WHERE name = 'bch-live'), 'bch-cust', 'bch-cust', 'bch-cust')`)

	s.checkout("bch-live-session")
	s.checkout("bch-gone-session")

	s.exec(`UPDATE billing_checkouts SET deleted_at = now() WHERE provider_id = 'bch-gone-session'`)
}

func (s *BillingCheckoutRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_CHECKOUTS, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingCheckoutRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *BillingCheckoutRepositoryPGTestSuite) customerID() string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_customers WHERE name = 'bch-cust'`)
}

func (s *BillingCheckoutRepositoryPGTestSuite) checkoutID(providerID string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		`SELECT id FROM billing_checkouts WHERE provider_id = $1`, providerID)
}

func (s *BillingCheckoutRepositoryPGTestSuite) checkout(providerID string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_checkouts (customer_id, provider_id, checkout_url, state)
		VALUES ((SELECT id FROM billing_customers WHERE name = 'bch-cust'), $1, $1, 'pending')`, providerID)
}

func (s *BillingCheckoutRepositoryPGTestSuite) TestGetByIDSkipsDeleted() {
	got, err := s.repository.GetByID(s.ctx, s.checkoutID("bch-live-session"))
	s.Require().NoError(err)
	s.Equal("bch-live-session", got.ProviderID)

	_, err = s.repository.GetByID(s.ctx, s.checkoutID("bch-gone-session"))
	s.ErrorIs(err, checkout.ErrNotFound)
}

func (s *BillingCheckoutRepositoryPGTestSuite) TestListSkipsDeleted() {
	got, err := s.repository.List(s.ctx, checkout.Filter{CustomerID: s.customerID()})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("bch-live-session", got[0].ProviderID)
}

func TestBillingCheckoutRepositoryPG(t *testing.T) {
	suite.Run(t, new(BillingCheckoutRepositoryPGTestSuite))
}
