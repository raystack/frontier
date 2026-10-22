package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/auditrecord"
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

	subscriptions *postgres.BillingSubscriptionRepository
	invoices      *postgres.BillingInvoiceRepository
	checkouts     *postgres.BillingCheckoutRepository
	transactions  *postgres.BillingTransactionRepository
}

func (s *BillingCustomerRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewBillingCustomerRepository(s.client)
	s.subscriptions = postgres.NewBillingSubscriptionRepository(s.client)
	s.invoices = postgres.NewBillingInvoiceRepository(s.client)
	s.checkouts = postgres.NewBillingCheckoutRepository(s.client)
	s.transactions = postgres.NewBillingTransactionRepository(s.client)
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
	for _, table := range []string{postgres.TABLE_BILLING_TRANSACTIONS, postgres.TABLE_BILLING_CHECKOUTS, postgres.TABLE_BILLING_INVOICES, postgres.TABLE_BILLING_SUBSCRIPTIONS, postgres.TABLE_BILLING_PLANS, postgres.TABLE_BILLING_CUSTOMERS, postgres.TABLE_ORGANIZATIONS, postgres.TABLE_AUDITRECORDS} {
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

func (s *BillingCustomerRepositoryPGTestSuite) TestUpdateCreditMinByIDSkipsDeleted() {
	got, err := s.repository.UpdateCreditMinByID(s.ctx, s.customerID("bc-cust-live"), -500)
	s.Require().NoError(err)
	s.Equal(int64(-500), got.CreditMin)

	goneID := s.customerID("bc-cust-gone")
	updatedAtBefore := scalarSQL(s.T(), s.ctx, s.client, `SELECT updated_at::text FROM billing_customers WHERE id = $1`, goneID)

	_, err = s.repository.UpdateCreditMinByID(s.ctx, goneID, -500)
	s.ErrorIs(err, customer.ErrNotFound)

	s.Equal("-100", scalarSQL(s.T(), s.ctx, s.client, `SELECT credit_min::text FROM billing_customers WHERE id = $1`, goneID))
	s.Equal(updatedAtBefore, scalarSQL(s.T(), s.ctx, s.client, `SELECT updated_at::text FROM billing_customers WHERE id = $1`, goneID))
}

func TestBillingCustomerRepositoryPG(t *testing.T) {
	suite.Run(t, new(BillingCustomerRepositoryPGTestSuite))
}

func (s *BillingCustomerRepositoryPGTestSuite) TestDeleteKeepsTheRowAndHidesIt() {
	id := s.customerID("bc-cust-live")

	s.deleteAccount(id)

	deletedAt := scalarSQL(s.T(), s.ctx, s.client,
		`SELECT coalesce(deleted_at::text, '') FROM billing_customers WHERE id = $1`, id)
	s.NotEmpty(deletedAt, "the row stays with its delete time")
	occurredAt := scalarSQL(s.T(), s.ctx, s.client,
		`SELECT occurred_at::text FROM audit_records WHERE event = $1 AND resource_id = $2`,
		string(auditrecord.BillingCustomerDeletedEvent), id)
	s.Equal(deletedAt, occurredAt, "the audit record carries the delete time")
	s.Equal("0", s.recordedMetadata(id, "forfeited_tokens"), "an account without transactions forfeits zero")
	s.Equal("0", s.recordedMetadata(id, "purchased_tokens"))
	_, err := s.repository.GetByID(s.ctx, id)
	s.ErrorIs(err, customer.ErrNotFound)
	s.ErrorIs(s.deleteError(id), customer.ErrNotFound, "a second delete answers not found")
}

func (s *BillingCustomerRepositoryPGTestSuite) TestDeleteRefusesWhileLiveRecordsPointAtTheAccount() {
	cases := []struct {
		table string
		seed  string
	}{
		{
			table: "billing_subscriptions",
			seed: `INSERT INTO billing_subscriptions (customer_id, provider_id, plan_id, state)
				VALUES ($1, 'bc-sub', (SELECT id FROM billing_plans WHERE name = 'bc-plan'), 'active')`,
		},
		{
			table: "billing_invoices",
			seed: `INSERT INTO billing_invoices (customer_id, provider_id, amount, currency, hosted_url, state)
				VALUES ($1, 'bc-inv', 100, 'usd', 'bc-inv', 'open')`,
		},
		{
			table: "billing_checkouts",
			seed: `INSERT INTO billing_checkouts (customer_id, provider_id, checkout_url, state)
				VALUES ($1, 'bc-session', 'bc-session', 'pending')`,
		},
	}
	for _, tc := range cases {
		s.Run(tc.table, func() {
			s.exec(`INSERT INTO billing_plans (name, description) VALUES ('bc-plan', 'plan under test')`)
			id := s.customerID("bc-cust-live")
			s.exec(tc.seed, id)

			s.ErrorIs(s.deleteError(id), customer.ErrHasLiveRecords)
			s.Equal("", scalarSQL(s.T(), s.ctx, s.client,
				`SELECT coalesce(deleted_at::text, '') FROM billing_customers WHERE id = $1`, id),
				"the refused delete leaves the row live")

			s.exec(fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE customer_id = $1`, tc.table), id)
			s.NoError(s.deleteError(id), "a deleted child no longer blocks")

			s.exec(`TRUNCATE TABLE billing_plans RESTART IDENTITY CASCADE`)
			s.exec(`UPDATE billing_customers SET deleted_at = NULL WHERE id = $1`, id)
			s.exec(fmt.Sprintf(`DELETE FROM %s`, tc.table))
		})
	}
}

func (s *BillingCustomerRepositoryPGTestSuite) seedChildRows(cust string) (subID, invID string) {
	s.T().Helper()
	s.exec(`INSERT INTO billing_plans (name, description) VALUES ('bc-plan', 'plan under test')`)
	s.exec(`INSERT INTO billing_subscriptions (customer_id, provider_id, plan_id, state)
		VALUES ($1, 'bc-sub', (SELECT id FROM billing_plans WHERE name = 'bc-plan'), 'canceled')`, cust)
	s.exec(`INSERT INTO billing_invoices (customer_id, provider_id, amount, currency, hosted_url, state)
		VALUES ($1, 'bc-inv', 100, 'usd', 'bc-inv', 'paid')`, cust)
	s.exec(`INSERT INTO billing_checkouts (customer_id, provider_id, checkout_url, state)
		VALUES ($1, 'bc-session', 'bc-session', 'expired')`, cust)
	s.exec(`INSERT INTO billing_transactions (account_id, type, source, amount, description)
		VALUES ($1, 'credit', 'system.buy', 100, 'bought'), ($1, 'debit', 'usage', 30, 'spent')`, cust)
	subID = scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_subscriptions WHERE provider_id = 'bc-sub'`)
	invID = scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM billing_invoices WHERE provider_id = 'bc-inv'`)
	return subID, invID
}

func (s *BillingCustomerRepositoryPGTestSuite) deletedRows(table, column, id string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		fmt.Sprintf(`SELECT count(*)::text FROM %s WHERE %s = $1 AND deleted_at IS NOT NULL`, table, column), id)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestTeardownInTheDeleterOrderForfeitsTheTokens() {
	cust := s.customerID("bc-cust-live")
	subID, invID := s.seedChildRows(cust)

	s.Require().NoError(s.subscriptions.Delete(s.ctx, subID))
	s.Require().NoError(s.invoices.Delete(s.ctx, invID))
	s.Require().NoError(s.checkouts.DeleteByCustomerID(s.ctx, cust))
	s.Require().NoError(s.repository.Delete(s.ctx, cust, customer.DeleteOptions{OrgTeardown: true}))
	s.Equal("true", s.recordedMetadata(cust, "org_teardown"))
	s.Equal("70", s.recordedMetadata(cust, "forfeited_tokens"), "the customer delete hides the transactions and records what they held")
	s.Equal("70", s.recordedMetadata(cust, "purchased_tokens"))

	s.Equal("1", s.deletedRows(postgres.TABLE_BILLING_SUBSCRIPTIONS, "customer_id", cust))
	s.Equal("1", s.deletedRows(postgres.TABLE_BILLING_INVOICES, "customer_id", cust))
	s.Equal("1", s.deletedRows(postgres.TABLE_BILLING_CHECKOUTS, "customer_id", cust))
	s.Equal("2", s.deletedRows(postgres.TABLE_BILLING_TRANSACTIONS, "account_id", cust))
	s.Equal("1", s.deletedRows(postgres.TABLE_BILLING_CUSTOMERS, "id", cust))

	subs, err := s.subscriptions.List(s.ctx, subscription.Filter{CustomerID: cust})
	s.Require().NoError(err)
	s.Empty(subs)
	invs, err := s.invoices.List(s.ctx, invoice.Filter{CustomerID: cust})
	s.Require().NoError(err)
	s.Empty(invs)
	chks, err := s.checkouts.List(s.ctx, checkout.Filter{CustomerID: cust})
	s.Require().NoError(err)
	s.Empty(chks)
	balance, err := s.transactions.GetBalance(s.ctx, cust)
	s.Require().NoError(err)
	s.Equal(int64(0), balance)
	custs, err := s.repository.List(s.ctx, customer.Filter{OrgID: s.orgID("bc-live")})
	s.Require().NoError(err)
	s.Empty(custs)
	txns, err := s.transactions.List(s.ctx, credit.Filter{CustomerID: cust})
	s.Require().NoError(err)
	s.Empty(txns)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestTeardownCanBeRunAgain() {
	cust := s.customerID("bc-cust-live")
	subID, invID := s.seedChildRows(cust)
	s.Require().NoError(s.subscriptions.Delete(s.ctx, subID))
	s.Require().NoError(s.invoices.Delete(s.ctx, invID))
	s.Require().NoError(s.checkouts.DeleteByCustomerID(s.ctx, cust))

	s.Require().NoError(s.subscriptions.Delete(s.ctx, subID))
	s.Require().NoError(s.invoices.Delete(s.ctx, invID))
	s.Require().NoError(s.checkouts.DeleteByCustomerID(s.ctx, cust))
	s.Require().NoError(s.deleteError(cust))
	s.ErrorIs(s.deleteError(cust), customer.ErrNotFound)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestDeleteWaitsForAnInFlightChildInsertAndRefuses() {
	id := s.customerID("bc-cust-live")
	tx, err := s.client.DB.BeginTxx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck
	var holderPID int
	s.Require().NoError(tx.QueryRowContext(s.ctx, "SELECT pg_backend_pid()").Scan(&holderPID))
	_, err = tx.ExecContext(s.ctx, `INSERT INTO billing_checkouts (customer_id, provider_id, checkout_url, state)
		VALUES ($1, 'bc-session', 'bc-session', 'pending')`, id)
	s.Require().NoError(err)

	result := make(chan error, 1)
	go func() { result <- s.deleteError(id) }()

	s.Require().Eventually(func() bool {
		waiting, err := blockedBy(s.ctx, s.client, holderPID)
		return err == nil && waiting == 1
	}, 5*time.Second, 20*time.Millisecond, "the delete waits on the row lock")

	s.Require().NoError(tx.Commit())
	s.ErrorIs(<-result, customer.ErrHasLiveRecords)
	s.Equal("", scalarSQL(s.T(), s.ctx, s.client,
		`SELECT coalesce(deleted_at::text, '') FROM billing_customers WHERE id = $1`, id))
}

func (s *BillingCustomerRepositoryPGTestSuite) seedTransaction(account, txType, source string, amount int64) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		`INSERT INTO billing_transactions (account_id, type, source, amount, description)
		VALUES ($1, $2, $3, $4, $2) RETURNING id`, account, txType, source, amount)
}

func (s *BillingCustomerRepositoryPGTestSuite) recordedMetadata(accountID, key string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		`SELECT coalesce(metadata->>$3, '') FROM audit_records WHERE event = $1 AND resource_id = $2`,
		string(auditrecord.BillingCustomerDeletedEvent), accountID, key)
}

func (s *BillingCustomerRepositoryPGTestSuite) transactionDeletedAt(id string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		`SELECT coalesce(deleted_at::text, '') FROM billing_transactions WHERE id = $1`, id)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestDeleteHidesTheAccountsTransactions() {
	id := s.customerID("bc-cust-live")
	bought := s.seedTransaction(id, "credit", "system.buy", 200)
	refunded := s.seedTransaction(id, "debit", "system.buy", 20)
	awarded := s.seedTransaction(id, "credit", "system.awarded", 100)
	used := s.seedTransaction(id, "debit", "system", 30)
	alreadyHidden := s.seedTransaction(id, "credit", "system.buy", 50)
	s.exec(`UPDATE billing_transactions SET deleted_at = now() - interval '1 day' WHERE id = $1`, alreadyHidden)
	hiddenAt := s.transactionDeletedAt(alreadyHidden)

	s.Require().NoError(s.repository.Delete(s.ctx, id, customer.DeleteOptions{}))

	for _, txn := range []string{bought, refunded, awarded, used} {
		s.NotEmpty(s.transactionDeletedAt(txn))
	}
	s.Equal(hiddenAt, s.transactionDeletedAt(alreadyHidden), "an older delete time is kept")
	s.Equal("false", s.recordedMetadata(id, "org_teardown"), "a delete outside the org teardown is not marked")
	s.Equal("250", s.recordedMetadata(id, "forfeited_tokens"), "the live credits minus the live debits")
	s.Equal("180", s.recordedMetadata(id, "purchased_tokens"), "the share of it bought, net of refunds")
	balance, err := s.transactions.GetBalance(s.ctx, id)
	s.Require().NoError(err)
	s.Equal(int64(0), balance)
}

func (s *BillingCustomerRepositoryPGTestSuite) TestDeleteRefusedByLiveRecordsLeavesTheTransactionsLive() {
	id := s.customerID("bc-cust-live")
	s.exec(`INSERT INTO billing_checkouts (customer_id, provider_id, checkout_url, state)
		VALUES ($1, 'bc-session', 'bc-session', 'pending')`, id)
	txn := s.seedTransaction(id, "credit", "system.buy", 200)

	s.ErrorIs(s.deleteError(id), customer.ErrHasLiveRecords)

	s.Empty(s.transactionDeletedAt(txn))
}

func (s *BillingCustomerRepositoryPGTestSuite) TestDeleteRollsBackWhenTheTransactionsNetToDebt() {
	id := s.customerID("bc-cust-live")
	credited := s.seedTransaction(id, "credit", "system.buy", 20)
	used := s.seedTransaction(id, "debit", "system", 70)

	s.ErrorIs(s.deleteError(id), customer.ErrHasTokenDebt)

	s.Empty(s.transactionDeletedAt(credited))
	s.Empty(s.transactionDeletedAt(used))
	s.Equal("", scalarSQL(s.T(), s.ctx, s.client,
		`SELECT coalesce(deleted_at::text, '') FROM billing_customers WHERE id = $1`, id))
}

func (s *BillingCustomerRepositoryPGTestSuite) TestHasLiveRecordsSeesOnlyLiveChildren() {
	id := s.customerID("bc-cust-live")

	has, err := s.repository.HasLiveRecords(s.ctx, id)
	s.Require().NoError(err)
	s.False(has)

	s.exec(`INSERT INTO billing_checkouts (customer_id, provider_id, checkout_url, state)
		VALUES ($1, 'bc-session', 'bc-session', 'pending')`, id)
	has, err = s.repository.HasLiveRecords(s.ctx, id)
	s.Require().NoError(err)
	s.True(has)

	s.exec(`UPDATE billing_checkouts SET deleted_at = now() WHERE customer_id = $1`, id)
	has, err = s.repository.HasLiveRecords(s.ctx, id)
	s.Require().NoError(err)
	s.False(has, "a deleted child does not count")
}

func (s *BillingCustomerRepositoryPGTestSuite) deleteError(id string) error {
	return s.repository.Delete(s.ctx, id, customer.DeleteOptions{})
}

func (s *BillingCustomerRepositoryPGTestSuite) deleteAccount(id string) {
	s.T().Helper()
	s.Require().NoError(s.deleteError(id))
}
