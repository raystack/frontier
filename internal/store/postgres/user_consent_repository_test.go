package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"

	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
)

type UserConsentRepositoryTestSuite struct {
	suite.Suite
	ctx            context.Context
	client         *db.Client
	pool           *dockertest.Pool
	resource       *dockertest.Resource
	repository     *postgres.UserConsentRepository
	userRepository *postgres.UserRepository
}

func (s *UserConsentRepositoryTestSuite) SetupSuite() {
	var err error

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewUserConsentRepository(s.client)
	s.userRepository = postgres.NewUserRepository(s.client)
}

func (s *UserConsentRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserConsentRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserConsentRepositoryTestSuite) cleanup() error {
	// TRUNCATE is not a DELETE, so the row level BEFORE DELETE trigger that
	// makes the table immutable does not fire on it.
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_USER_CONSENTS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_USERS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func testDocuments() []consent.Document {
	return []consent.Document{
		{
			ID:      "privacy_policy",
			Title:   "Privacy Policy",
			Version: "2026-04-01",
			URL:     "https://example.org/legal/privacy/2026-04-01",
		},
		{
			ID:      "terms_of_service",
			Title:   "Terms & Conditions",
			Version: "2026-04-01",
			URL:     "https://example.org/legal/terms/2026-04-01",
		},
	}
}

func newUser(email string) user.User {
	return user.User{
		Name:  email,
		Email: email,
		Title: "Consenting User",
		State: user.Enabled,
	}
}

func (s *UserConsentRepositoryTestSuite) countConsents(userID string) int {
	var count int
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE user_id = $1", postgres.TABLE_USER_CONSENTS)
	s.Require().NoError(s.client.DB.QueryRowxContext(s.ctx, query, userID).Scan(&count))
	return count
}

// TestCreate covers the write itself: what the record keeps, and the two rules
// the table enforces on it.
func (s *UserConsentRepositoryTestSuite) TestCreate() {
	consentedAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)

	s.Run("should write a consent record inside the transaction it is given", func() {
		defer func() { s.Require().NoError(s.cleanup()) }()

		createdUser, err := s.userRepository.Create(s.ctx, newUser("complete@example.com"))
		s.Require().NoError(err)

		var granted consent.Consent
		err = s.client.WithTxn(s.ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
			var txErr error
			granted, txErr = s.repository.Create(s.ctx, tx, consent.Consent{
				UserID:       createdUser.ID,
				UserEmail:    createdUser.Email,
				Documents:    testDocuments(),
				Source:       consent.SourceSignup,
				AuthStrategy: "mailotp",
				IPAddress:    "203.0.113.9",
				ConsentedAt:  consentedAt,
			})
			return txErr
		})
		s.Require().NoError(err)

		s.Assert().NotEmpty(granted.ID)
		s.Assert().Equal(createdUser.ID, granted.UserID)
		s.Assert().Equal(createdUser.Email, granted.UserEmail)
		s.Assert().Equal(consent.SourceSignup, granted.Source)
		s.Assert().Equal("mailotp", granted.AuthStrategy)
		s.Assert().Equal("203.0.113.9", granted.IPAddress)
		s.Assert().True(consentedAt.Equal(granted.ConsentedAt))
		s.Assert().False(granted.CreatedAt.IsZero())
		// every document keeps all four fields, copied from config at write
		// time, so the record stays readable after the document leaves config
		s.Assert().Equal(testDocuments(), granted.Documents)
	})

	s.Run("should store a missing strategy and ip as null rather than failing", func() {
		defer func() { s.Require().NoError(s.cleanup()) }()

		createdUser, err := s.userRepository.Create(s.ctx, newUser("noip@example.com"))
		s.Require().NoError(err)

		var granted consent.Consent
		err = s.client.WithTxn(s.ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
			var txErr error
			granted, txErr = s.repository.Create(s.ctx, tx, consent.Consent{
				UserID:      createdUser.ID,
				UserEmail:   createdUser.Email,
				Documents:   testDocuments(),
				Source:      consent.SourceSignup,
				ConsentedAt: consentedAt,
			})
			return txErr
		})
		s.Require().NoError(err)
		s.Assert().Empty(granted.AuthStrategy)
		s.Assert().Empty(granted.IPAddress)
	})

	s.Run("should reject a second signup consent for the same user", func() {
		defer func() { s.Require().NoError(s.cleanup()) }()

		createdUser, err := s.userRepository.Create(s.ctx, newUser("twice@example.com"))
		s.Require().NoError(err)

		write := func() error {
			return s.client.WithTxn(s.ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
				_, txErr := s.repository.Create(s.ctx, tx, consent.Consent{
					UserID:      createdUser.ID,
					UserEmail:   createdUser.Email,
					Documents:   testDocuments(),
					Source:      consent.SourceSignup,
					ConsentedAt: consentedAt,
				})
				return txErr
			})
		}
		s.Require().NoError(write())

		// the partial unique index: nothing repairs a record, so a second
		// signup write is a bug and has to fail rather than leave two rows
		s.Assert().ErrorIs(write(), consent.ErrConsentExists)
		s.Assert().Equal(1, s.countConsents(createdUser.ID))
	})

	s.Run("should reject a record that names no document", func() {
		defer func() { s.Require().NoError(s.cleanup()) }()

		createdUser, err := s.userRepository.Create(s.ctx, newUser("empty@example.com"))
		s.Require().NoError(err)

		err = s.client.WithTxn(s.ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
			_, txErr := s.repository.Create(s.ctx, tx, consent.Consent{
				UserID:      createdUser.ID,
				UserEmail:   createdUser.Email,
				Source:      consent.SourceSignup,
				ConsentedAt: consentedAt,
			})
			return txErr
		})
		s.Assert().Error(err)
		s.Assert().Equal(0, s.countConsents(createdUser.ID))
	})

	s.Run("should refuse to write without a transaction", func() {
		_, err := s.repository.Create(s.ctx, nil, consent.Consent{})
		s.Assert().ErrorIs(err, consent.ErrInvalidGrant)
	})
}

// TestCreateIsAtomicWithTheUserRow is the invariant this whole feature rests
// on: a user row without a consent record is impossible. It runs against a real
// database on purpose — a mocked transaction can only pretend to roll back.
func (s *UserConsentRepositoryTestSuite) TestCreateIsAtomicWithTheUserRow() {
	consentedAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)

	s.Run("should keep both rows when both inserts succeed", func() {
		defer func() { s.Require().NoError(s.cleanup()) }()

		var createdUser user.User
		err := s.client.WithTxn(s.ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
			var txErr error
			createdUser, txErr = s.userRepository.CreateWithTx(s.ctx, tx, newUser("both@example.com"))
			if txErr != nil {
				return txErr
			}
			_, txErr = s.repository.Create(s.ctx, tx, consent.Consent{
				UserID:       createdUser.ID,
				UserEmail:    createdUser.Email,
				Documents:    testDocuments(),
				Source:       consent.SourceSignup,
				AuthStrategy: "mailotp",
				ConsentedAt:  consentedAt,
			})
			return txErr
		})
		s.Require().NoError(err)

		fetched, err := s.userRepository.GetByID(s.ctx, createdUser.ID)
		s.Require().NoError(err)
		s.Assert().Equal("both@example.com", fetched.Email)
		s.Assert().Equal(1, s.countConsents(createdUser.ID))
	})

	s.Run("should roll the user row back when the consent insert fails", func() {
		defer func() { s.Require().NoError(s.cleanup()) }()

		var createdUser user.User
		err := s.client.WithTxn(s.ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
			var txErr error
			createdUser, txErr = s.userRepository.CreateWithTx(s.ctx, tx, newUser("rollback@example.com"))
			if txErr != nil {
				return txErr
			}
			// an empty document list violates documents_not_empty, so this is
			// a real failure from Postgres inside a real transaction
			_, txErr = s.repository.Create(s.ctx, tx, consent.Consent{
				UserID:      createdUser.ID,
				UserEmail:   createdUser.Email,
				Source:      consent.SourceSignup,
				ConsentedAt: consentedAt,
			})
			return txErr
		})
		s.Require().Error(err)
		s.Require().NotEmpty(createdUser.ID)

		_, err = s.userRepository.GetByID(s.ctx, createdUser.ID)
		s.Assert().ErrorIs(err, user.ErrNotExist)
		s.Assert().Equal(0, s.countConsents(createdUser.ID))

		var count int
		query := fmt.Sprintf("SELECT count(*) FROM %s WHERE email = $1", postgres.TABLE_USERS)
		s.Require().NoError(s.client.DB.QueryRowxContext(s.ctx, query, "rollback@example.com").Scan(&count))
		s.Assert().Zero(count)
	})
}

func TestUserConsentRepository(t *testing.T) {
	suite.Run(t, new(UserConsentRepositoryTestSuite))
}
