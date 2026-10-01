package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

// Runs the billing customer reads against a real postgres to check that a
// soft-deleted customer stays out of every read.
type BillingCustomerRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.BillingCustomerRepository
}

func (s *BillingCustomerRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewBillingCustomerRepository(s.client)
}

func (s *BillingCustomerRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingCustomerRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES ('bc-live', 'Live Org')`)

	s.customer("bc-cust-live", "bc-live", -100)
	s.customer("bc-cust-gone", "bc-live", -100)

	s.exec(`UPDATE billing_customers SET deleted_at = now() WHERE name = 'bc-cust-gone'`)
}

func (s *BillingCustomerRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_CUSTOMERS, postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingCustomerRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *BillingCustomerRepositoryPGTestSuite) orgID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *BillingCustomerRepositoryPGTestSuite) customerID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_customers WHERE name = $1`, name)
}

func (s *BillingCustomerRepositoryPGTestSuite) customer(name, orgName string, creditMin int64) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_customers (org_id, provider_id, name, email, credit_min)
		VALUES ($1, $2, $2, $2, $3)`, s.orgID(orgName), name, creditMin)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestGetByIDSkipsDeleted() {
	got, err := s.repository.GetByID(s.ctx, s.customerID("bc-cust-live"))
	s.Require().NoError(err)
	s.Equal("bc-cust-live", got.Name)

	_, err = s.repository.GetByID(s.ctx, s.customerID("bc-cust-gone"))
	s.ErrorIs(err, customer.ErrNotFound)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestListSkipsDeleted() {
	got, err := s.repository.List(s.ctx, customer.Filter{OrgID: s.orgID("bc-live")})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("bc-cust-live", got[0].Name)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestGetDetailsByIDSkipsDeleted() {
	got, err := s.repository.GetDetailsByID(s.ctx, s.customerID("bc-cust-live"))
	s.Require().NoError(err)
	s.Equal(int64(-100), got.CreditMin)

	_, err = s.repository.GetDetailsByID(s.ctx, s.customerID("bc-cust-gone"))
	s.ErrorIs(err, customer.ErrNotFound)
}

func TestBillingCustomerRepositoryPG(t *testing.T) {
	suite.Run(t, new(BillingCustomerRepositoryPGTestSuite))
}
