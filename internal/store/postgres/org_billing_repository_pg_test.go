package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

type OrgBillingRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.OrgBillingRepository
}

func (s *OrgBillingRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgBillingRepository(s.client)
}

func (s *OrgBillingRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgBillingRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_SUBSCRIPTIONS, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_BILLING_PLANS, postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgBillingRepositoryPGTestSuite) TestSearchSkipsDeletedOrgs() {
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO organizations (name, title, avatar) VALUES ('ob-live', 'Live Org', ''), ('ob-gone', 'Gone Org', '')`)
	execSQL(s.T(), s.ctx, s.client, `UPDATE organizations SET deleted_at = now() WHERE name = 'ob-gone'`)

	got, err := s.repository.Search(s.ctx, &rql.Query{Limit: 10})
	s.Require().NoError(err)

	s.Require().Len(got.Organizations, 1)
	s.Assert().Equal("ob-live", got.Organizations[0].Name)
}

func (s *OrgBillingRepositoryPGTestSuite) seedBilling(org string, creditMin int) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO organizations (name, title, avatar) VALUES ($1, $1, '')`, org)
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO billing_customers (org_id, provider_id, name, email, credit_min)
		VALUES ((SELECT id FROM organizations WHERE name = $1), $1, $1, $1, $2)`, org, creditMin)
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO billing_plans (name, description) VALUES ('plan-'||$1::text, 'plan')`, org)
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO billing_subscriptions (customer_id, provider_id, plan_id, state)
		VALUES ((SELECT id FROM billing_customers WHERE name = $1), $1,
			(SELECT id FROM billing_plans WHERE name = 'plan-'||$1::text), 'active')`, org)
}

func (s *OrgBillingRepositoryPGTestSuite) searchOne(org string) orgbilling.AggregatedOrganization {
	s.T().Helper()
	got, err := s.repository.Search(s.ctx, &rql.Query{Limit: 10})
	s.Require().NoError(err)
	for _, o := range got.Organizations {
		if o.Name == org {
			return o
		}
	}
	s.T().Fatalf("org %s not listed", org)
	return orgbilling.AggregatedOrganization{}
}

func (s *OrgBillingRepositoryPGTestSuite) TestSearchKeepsLiveBillingRows() {
	s.seedBilling("ob-live-bill", -10)

	got := s.searchOne("ob-live-bill")

	s.Assert().Equal("plan-ob-live-bill", got.PlanName)
	s.Assert().Equal("active", got.SubscriptionState)
	s.Assert().Equal("postpaid", got.PaymentMode)
}

func (s *OrgBillingRepositoryPGTestSuite) TestSearchIgnoresDeletedBillingCustomer() {
	s.seedBilling("ob-no-cust", -10)
	execSQL(s.T(), s.ctx, s.client, `UPDATE billing_customers SET deleted_at = now() WHERE name = 'ob-no-cust'`)

	got := s.searchOne("ob-no-cust")

	s.Assert().Empty(got.PlanName)
	s.Assert().Empty(got.SubscriptionState)
	s.Assert().Empty(got.PaymentMode)
}

func (s *OrgBillingRepositoryPGTestSuite) TestSearchIgnoresDeletedSubscription() {
	s.seedBilling("ob-no-sub", -10)
	execSQL(s.T(), s.ctx, s.client, `UPDATE billing_subscriptions SET deleted_at = now() WHERE provider_id = 'ob-no-sub'`)

	got := s.searchOne("ob-no-sub")

	s.Assert().Empty(got.PlanName)
	s.Assert().Empty(got.SubscriptionState)
	s.Assert().Equal("postpaid", got.PaymentMode)
}

func TestOrgBillingRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgBillingRepositoryPGTestSuite))
}
