package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the billing invoice reads against a real postgres to check that a
// soft-deleted invoice stays out of every read, and that the admin search also
// drops invoices whose billing customer or organization is soft-deleted.
type BillingInvoiceRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.BillingInvoiceRepository
}

func (s *BillingInvoiceRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewBillingInvoiceRepository(s.client)
}

func (s *BillingInvoiceRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingInvoiceRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES ('bi-live', 'Live Org'), ('bi-gone', 'Gone Org')`)

	s.customer("bi-cust-live", "bi-live")
	s.customer("bi-cust-gone", "bi-live")
	s.customer("bi-cust-other", "bi-gone")

	s.invoice("bi-cust-live", "bi-inv-live", "paid")
	s.invoice("bi-cust-live", "bi-inv-deleted", "paid")
	s.invoice("bi-cust-gone", "bi-inv-orphan", "open")
	s.invoice("bi-cust-other", "bi-inv-otherorg", "paid")

	s.exec(`UPDATE billing_invoices SET deleted_at = now() WHERE hosted_url = 'bi-inv-deleted'`)
	s.exec(`UPDATE billing_customers SET deleted_at = now() WHERE name = 'bi-cust-gone'`)
	s.exec(`UPDATE organizations SET deleted_at = now() WHERE name = 'bi-gone'`)
}

func (s *BillingInvoiceRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_INVOICES, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingInvoiceRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *BillingInvoiceRepositoryPGTestSuite) orgID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *BillingInvoiceRepositoryPGTestSuite) customerID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_customers WHERE name = $1`, name)
}

func (s *BillingInvoiceRepositoryPGTestSuite) invoiceID(hostedURL string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_invoices WHERE hosted_url = $1`, hostedURL)
}

func (s *BillingInvoiceRepositoryPGTestSuite) customer(name, orgName string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_customers (org_id, provider_id, name, email)
		VALUES ($1, $2, $2, $2)`, s.orgID(orgName), name)
}

func (s *BillingInvoiceRepositoryPGTestSuite) invoice(customer, hostedURL, state string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_invoices (customer_id, provider_id, amount, currency, hosted_url, state)
		VALUES ((SELECT id FROM billing_customers WHERE name = $1), $2, 100, 'usd', $2, $3)`,
		customer, hostedURL, state)
}

func (s *BillingInvoiceRepositoryPGTestSuite) links(invoices []invoice.Invoice) []string {
	out := make([]string, 0, len(invoices))
	for _, i := range invoices {
		out = append(out, i.HostedURL)
	}
	return out
}

func (s *BillingInvoiceRepositoryPGTestSuite) TestGetByIDSkipsDeleted() {
	got, err := s.repository.GetByID(s.ctx, s.invoiceID("bi-inv-live"))
	s.Require().NoError(err)
	s.Equal("bi-inv-live", got.HostedURL)

	_, err = s.repository.GetByID(s.ctx, s.invoiceID("bi-inv-deleted"))
	s.ErrorIs(err, invoice.ErrNotFound)
}

func (s *BillingInvoiceRepositoryPGTestSuite) TestListSkipsDeletedInvoicesAndCountsAgree() {
	page := pagination.NewPagination(1, 10)
	got, err := s.repository.List(s.ctx, invoice.Filter{
		CustomerID: s.customerID("bi-cust-live"),
		Pagination: page,
	})
	s.Require().NoError(err)
	s.Equal([]string{"bi-inv-live"}, s.links(got))
	s.Equal(int32(1), page.Count, "the count statement carries the same filter as the rows")
}

func (s *BillingInvoiceRepositoryPGTestSuite) TestListReadsOnlyTheInvoiceTable() {
	// List has no join, so an invoice of a soft-deleted customer still lists.
	// That matches every other single-table List in the store. The admin
	// search below is where the customer and organization are checked.
	got, err := s.repository.List(s.ctx, invoice.Filter{CustomerID: s.customerID("bi-cust-gone")})
	s.Require().NoError(err)
	s.Equal([]string{"bi-inv-orphan"}, s.links(got))
}

func (s *BillingInvoiceRepositoryPGTestSuite) TestSearchSkipsDeletedInvoicesCustomersAndOrgs() {
	res, err := s.repository.Search(s.ctx, &rql.Query{Limit: 50})
	s.Require().NoError(err)
	out := make([]string, 0, len(res))
	for _, i := range res {
		out = append(out, i.InvoiceLink)
	}
	s.Equal([]string{"bi-inv-live"}, out)
}

func TestBillingInvoiceRepositoryPG(t *testing.T) {
	suite.Run(t, new(BillingInvoiceRepositoryPGTestSuite))
}
