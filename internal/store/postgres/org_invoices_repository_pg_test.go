package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the search against a real postgres to check that soft-deleted invoices,
// billing customers, and organizations stay out of both the list and the counts.
type OrgInvoicesRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.OrgInvoicesRepository
}

func (s *OrgInvoicesRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgInvoicesRepository(s.client)
}

func (s *OrgInvoicesRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgInvoicesRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES ('oi-live', 'Live Org'), ('oi-gone', 'Gone Org')`)

	s.customer("oi-cust-live", "oi-live")
	s.customer("oi-cust-gone", "oi-live")
	s.customer("oi-cust-other", "oi-gone")

	s.invoice("oi-cust-live", "oi-inv-live", "paid")
	s.invoice("oi-cust-live", "oi-inv-deleted", "paid")
	s.invoice("oi-cust-gone", "oi-inv-orphan", "open")
	s.invoice("oi-cust-other", "oi-inv-otherorg", "paid")

	s.exec(`UPDATE billing_invoices SET deleted_at = now() WHERE hosted_url = 'oi-inv-deleted'`)
	s.exec(`UPDATE billing_customers SET deleted_at = now() WHERE name = 'oi-cust-gone'`)
	s.exec(`UPDATE organizations SET deleted_at = now() WHERE name = 'oi-gone'`)
}

func (s *OrgInvoicesRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_INVOICES, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgInvoicesRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *OrgInvoicesRepositoryPGTestSuite) orgID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *OrgInvoicesRepositoryPGTestSuite) customer(name, orgName string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_customers (org_id, provider_id, name, email)
		VALUES ($1, $2, $2, $2)`, s.orgID(orgName), name)
}

func (s *OrgInvoicesRepositoryPGTestSuite) invoice(customer, hostedURL, state string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_invoices (customer_id, provider_id, amount, currency, hosted_url, state)
		VALUES ((SELECT id FROM billing_customers WHERE name = $1), $2, 100, 'usd', $2, $3)`,
		customer, hostedURL, state)
}

func (s *OrgInvoicesRepositoryPGTestSuite) links(orgName string) []string {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.orgID(orgName), &rql.Query{Limit: 50})
	s.Require().NoError(err)
	out := make([]string, 0, len(res.Invoices))
	for _, i := range res.Invoices {
		out = append(out, i.InvoiceLink)
	}
	return out
}

func (s *OrgInvoicesRepositoryPGTestSuite) TestSkipsDeletedInvoicesCustomersAndOrgs() {
	s.Equal([]string{"oi-inv-live"}, s.links("oi-live"))
}

func (s *OrgInvoicesRepositoryPGTestSuite) TestSoftDeletedOrgHasNoInvoices() {
	s.Empty(s.links("oi-gone"))
}

func (s *OrgInvoicesRepositoryPGTestSuite) TestGroupCountsSkipDeletedRows() {
	res, err := s.repository.Search(s.ctx, s.orgID("oi-live"),
		&rql.Query{Limit: 50, GroupBy: []string{"state"}})
	s.Require().NoError(err)
	s.Require().Len(res.Group.Data, 1)
	s.Equal("paid", res.Group.Data[0].Name)
	s.Equal(1, res.Group.Data[0].Count)
}

func TestOrgInvoicesRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgInvoicesRepositoryPGTestSuite))
}
