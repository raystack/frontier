package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

type BillingTransactionRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.BillingTransactionRepository
}

func (s *BillingTransactionRepositoryTestSuite) SetupSuite() {
	var err error

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewBillingTransactionRepository(s.client)
}

func (s *BillingTransactionRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *BillingTransactionRepositoryTestSuite) TearDownTest() {
	queries := []string{
		"TRUNCATE TABLE " + postgres.TABLE_BILLING_TRANSACTIONS,
	}
	if err := execQueries(context.TODO(), s.client, queries); err != nil {
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

func TestBillingTransactionRepository(t *testing.T) {
	suite.Run(t, new(BillingTransactionRepositoryTestSuite))
}
