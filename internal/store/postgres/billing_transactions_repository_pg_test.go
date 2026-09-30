package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/suite"
)

type BillingTransactionRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.BillingTransactionRepository
}

func (s *BillingTransactionRepositoryTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewBillingTransactionRepository(s.client)
}

func (s *BillingTransactionRepositoryTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingTransactionRepositoryTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_TRANSACTIONS, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingTransactionRepositoryTestSuite) TestCreateEntry() {
	debitEntry := credit.Transaction{
		ID:          uuid.New().String(),
		CustomerID:  schema.PlatformOrgID.String(),
		Type:        credit.DebitType,
		Amount:      10,
		Source:      credit.SourceSystemAwardedEvent,
		Description: "awarded credits",
	}
	creditEntry := credit.Transaction{
		ID:          uuid.New().String(),
		CustomerID:  uuid.New().String(),
		Type:        credit.CreditType,
		Amount:      10,
		Source:      credit.SourceSystemAwardedEvent,
		Description: "awarded credits",
	}

	s.Run("creates debit and credit entries", func() {
		entries, err := s.repository.CreateEntry(s.ctx, debitEntry, creditEntry)
		s.Assert().NoError(err)
		s.Assert().Len(entries, 2)
	})

	s.Run("replaying the same transaction returns ErrAlreadyApplied", func() {
		_, err := s.repository.CreateEntry(s.ctx, debitEntry, creditEntry)
		s.Assert().ErrorIs(err, credit.ErrAlreadyApplied)
	})
}

func (s *BillingTransactionRepositoryTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

// txn writes one ledger row straight into the table and returns its id.
func (s *BillingTransactionRepositoryTestSuite) txn(account, txType, source string, amount int64) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client,
		`INSERT INTO billing_transactions (account_id, type, source, amount, description)
		VALUES ($1, $2, $3, $4, $2) RETURNING id`, account, txType, source, amount)
}

func (s *BillingTransactionRepositoryTestSuite) softDelete(id string) {
	s.T().Helper()
	s.exec(`UPDATE billing_transactions SET deleted_at = now() WHERE id = $1`, id)
}

// customer seeds an organization and a live billing customer and returns the
// customer id, so CreateEntry can run its spending gate against a real account.
func (s *BillingTransactionRepositoryTestSuite) customer(name string, creditMin int64) string {
	s.T().Helper()
	s.exec(`INSERT INTO organizations (name, title) VALUES ($1, $1)`, name)
	return scalarSQL(s.T(), s.ctx, s.client,
		`INSERT INTO billing_customers (org_id, provider_id, name, email, credit_min)
		VALUES ((SELECT id FROM organizations WHERE name = $1), $1, $1, $1, $2) RETURNING id`, name, creditMin)
}

func (s *BillingTransactionRepositoryTestSuite) spend(account string, amount int64) error {
	s.T().Helper()
	_, err := s.repository.CreateEntry(s.ctx,
		credit.Transaction{ID: uuid.New().String(), CustomerID: account, Type: credit.DebitType, Amount: amount, Source: "usage"},
		credit.Transaction{CustomerID: schema.PlatformOrgID.String(), Type: credit.CreditType, Amount: amount, Source: "usage"})
	return err
}

func (s *BillingTransactionRepositoryTestSuite) TestBalanceSkipsDeletedTransactions() {
	account := uuid.New().String()
	s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 100)
	s.txn(account, string(credit.DebitType), "usage", 30)
	s.txn(account, string(credit.CreditType), credit.SourceSystemOverdraftEvent, 25)
	s.softDelete(s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 50))
	s.softDelete(s.txn(account, string(credit.DebitType), "usage", 20))
	s.softDelete(s.txn(account, string(credit.CreditType), credit.SourceSystemOverdraftEvent, 10))

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)

	balance, err := s.repository.GetBalance(s.ctx, account)
	s.Require().NoError(err)
	s.Equal(int64(95), balance, "100 + 25 - 30, deleted rows ignored")

	balance, err = s.repository.GetBalanceForRange(s.ctx, account, start, end)
	s.Require().NoError(err)
	s.Equal(int64(95), balance)

	balance, err = s.repository.GetBalanceForRangeWithoutOverdraft(s.ctx, account, start, end)
	s.Require().NoError(err)
	s.Equal(int64(70), balance, "100 - 30, overdraft credits and deleted rows ignored")

	debited, err := s.repository.GetTotalDebitedAmount(s.ctx, account)
	s.Require().NoError(err)
	s.Equal(int64(30), debited)
}

func (s *BillingTransactionRepositoryTestSuite) TestListAndGetByIDSkipDeleted() {
	account := uuid.New().String()
	live := s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 100)
	gone := s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 50)
	s.exec(`UPDATE billing_transactions SET metadata = '{"tag": "keep"}'::jsonb WHERE id IN ($1, $2)`, live, gone)
	s.softDelete(gone)

	got, err := s.repository.List(s.ctx, credit.Filter{CustomerID: account})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(live, got[0].ID)

	got, err = s.repository.List(s.ctx, credit.Filter{CustomerID: account, Metadata: metadata.Metadata{"tag": "keep"}})
	s.Require().NoError(err, "the metadata filter still works next to the live filter")
	s.Require().Len(got, 1)
	s.Equal(live, got[0].ID)

	_, err = s.repository.GetByID(s.ctx, live)
	s.NoError(err)
	_, err = s.repository.GetByID(s.ctx, gone)
	s.ErrorIs(err, credit.ErrNotFound)
}

func (s *BillingTransactionRepositoryTestSuite) TestSpendingGateAgreesWithBalance() {
	account := s.customer("bt-gate", 0)
	s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 100)
	s.softDelete(s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 50))

	s.ErrorIs(s.spend(account, 120), credit.ErrInsufficientCredits, "the deleted 50 must not be spendable")
	s.NoError(s.spend(account, 80))

	balance, err := s.repository.GetBalance(s.ctx, account)
	s.Require().NoError(err)
	s.Equal(int64(20), balance)
}

func (s *BillingTransactionRepositoryTestSuite) TestCreateEntryRefusesDeletedCustomer() {
	account := s.customer("bt-gone", 0)
	s.txn(account, string(credit.CreditType), credit.SourceSystemBuyEvent, 100)
	s.exec(`UPDATE billing_customers SET deleted_at = now() WHERE id = $1`, account)

	s.ErrorIs(s.spend(account, 10), customer.ErrNotFound)
}

func TestBillingTransactionRepository(t *testing.T) {
	suite.Run(t, new(BillingTransactionRepositoryTestSuite))
}
