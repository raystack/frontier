package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

// Runs the billing subscription reads against a real postgres to check that a
// soft-deleted subscription stays out of every read.
type BillingSubscriptionRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.BillingSubscriptionRepository
}

func (s *BillingSubscriptionRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewBillingSubscriptionRepository(s.client)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingSubscriptionRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES ('bs-live', 'Live Org')`)
	s.exec(`INSERT INTO billing_customers (org_id, provider_id, name, email)
		VALUES ((SELECT id FROM organizations WHERE name = 'bs-live'), 'bs-cust', 'bs-cust', 'bs-cust')`)
	s.exec(`INSERT INTO billing_plans (name, description) VALUES ('bs-plan', 'plan under test')`)

	s.subscription("bs-sub-live")
	s.subscription("bs-sub-gone")

	s.exec(`UPDATE billing_subscriptions SET deleted_at = now() WHERE provider_id = 'bs-sub-gone'`)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_SUBSCRIPTIONS, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_BILLING_PLANS, postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingSubscriptionRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) customerID() string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_customers WHERE name = 'bs-cust'`)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) subscriptionID(providerID string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		`SELECT id FROM billing_subscriptions WHERE provider_id = $1`, providerID)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) subscription(providerID string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_subscriptions (customer_id, provider_id, plan_id, state)
		VALUES ((SELECT id FROM billing_customers WHERE name = 'bs-cust'), $1,
			(SELECT id FROM billing_plans WHERE name = 'bs-plan'), 'active')`, providerID)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) TestGetByIDSkipsDeleted() {
	got, err := s.repository.GetByID(s.ctx, s.subscriptionID("bs-sub-live"))
	s.Require().NoError(err)
	s.Equal("bs-sub-live", got.ProviderID)

	_, err = s.repository.GetByID(s.ctx, s.subscriptionID("bs-sub-gone"))
	s.ErrorIs(err, subscription.ErrNotFound)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) TestGetByProviderIDSkipsDeleted() {
	got, err := s.repository.GetByProviderID(s.ctx, "bs-sub-live")
	s.Require().NoError(err)
	s.Equal("bs-sub-live", got.ProviderID)

	_, err = s.repository.GetByProviderID(s.ctx, "bs-sub-gone")
	s.ErrorIs(err, subscription.ErrNotFound)
}

func (s *BillingSubscriptionRepositoryPGTestSuite) TestListSkipsDeleted() {
	got, err := s.repository.List(s.ctx, subscription.Filter{CustomerID: s.customerID()})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("bs-sub-live", got[0].ProviderID)
}

func TestBillingSubscriptionRepositoryPG(t *testing.T) {
	suite.Run(t, new(BillingSubscriptionRepositoryPGTestSuite))
}
