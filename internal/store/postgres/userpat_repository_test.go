package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type UserPATRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.UserPATRepository
	users      []user.User
	orgs       []organization.Organization
}

func (s *UserPATRepositoryTestSuite) SetupSuite() {
	var err error
	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewUserPATRepository(s.client)
}

func (s *UserPATRepositoryTestSuite) SetupTest() {
	var err error
	s.orgs, err = bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.users, err = bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserPATRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserPATRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserPATRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_USER_TOKENS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_USERS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ORGANIZATIONS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *UserPATRepositoryTestSuite) TestCreate() {
	s.Run("should create a token and return it with generated ID", func() {
		pat := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "test-token",
			SecretHash: "hash123",
			ExpiresAt:  time.Now().Add(24 * time.Hour).Truncate(time.Microsecond),
		}

		created, err := s.repository.Create(s.ctx, pat)
		s.Require().NoError(err)
		s.NotEmpty(created.ID)
		s.Equal(pat.UserID, created.UserID)
		s.Equal(pat.OrgID, created.OrgID)
		s.Equal(pat.Title, created.Title)
		s.Equal(pat.SecretHash, created.SecretHash)
		s.WithinDuration(pat.ExpiresAt, created.ExpiresAt, time.Microsecond)
		s.False(created.CreatedAt.IsZero())
		s.False(created.UpdatedAt.IsZero())
	})

	s.Run("should use provided ID if set", func() {
		customID := uuid.New().String()
		pat := userpat.PersonalAccessToken{
			ID:         customID,
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "token-with-id",
			SecretHash: "hash456",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		created, err := s.repository.Create(s.ctx, pat)
		s.Require().NoError(err)
		s.Equal(customID, created.ID)
	})

	s.Run("should store and return metadata", func() {
		pat := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "token-with-meta",
			SecretHash: "hash789",
			Metadata:   map[string]any{"env": "staging", "purpose": "ci"},
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		created, err := s.repository.Create(s.ctx, pat)
		s.Require().NoError(err)
		s.Equal("staging", created.Metadata["env"])
		s.Equal("ci", created.Metadata["purpose"])
	})

	s.Run("should return ErrConflict for duplicate title per user per org", func() {
		pat := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "duplicate-title",
			SecretHash: "hashA",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		_, err := s.repository.Create(s.ctx, pat)
		s.Require().NoError(err)

		pat.ID = ""
		pat.SecretHash = "hashB"
		_, err = s.repository.Create(s.ctx, pat)
		s.ErrorIs(err, userpat.ErrConflict)
	})

	s.Run("should return ErrConflict for duplicate secret hash", func() {
		pat1 := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "token-unique-hash-1",
			SecretHash: "same-hash",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		_, err := s.repository.Create(s.ctx, pat1)
		s.Require().NoError(err)

		pat2 := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "token-unique-hash-2",
			SecretHash: "same-hash",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		_, err = s.repository.Create(s.ctx, pat2)
		s.ErrorIs(err, userpat.ErrConflict)
	})

	s.Run("should allow same title for different users in same org", func() {
		pat1 := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "shared-title",
			SecretHash: "hashUser1",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		_, err := s.repository.Create(s.ctx, pat1)
		s.Require().NoError(err)

		pat2 := userpat.PersonalAccessToken{
			UserID:     s.users[1].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "shared-title",
			SecretHash: "hashUser2",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		_, err = s.repository.Create(s.ctx, pat2)
		s.NoError(err)
	})

	s.Run("should allow same title for same user in different orgs", func() {
		pat1 := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      "cross-org-title",
			SecretHash: "hashOrg1",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		_, err := s.repository.Create(s.ctx, pat1)
		s.Require().NoError(err)

		pat2 := userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[1].ID,
			Title:      "cross-org-title",
			SecretHash: "hashOrg2",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		_, err = s.repository.Create(s.ctx, pat2)
		s.NoError(err)
	})
}

func (s *UserPATRepositoryTestSuite) truncateTokens() {
	err := execQueries(context.TODO(), s.client, []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_USER_TOKENS),
	})
	s.Require().NoError(err)
}

func (s *UserPATRepositoryTestSuite) TestCountActive_Empty() {
	count, err := s.repository.CountActive(s.ctx, s.users[0].ID, s.orgs[0].ID)
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

func (s *UserPATRepositoryTestSuite) TestCountActive_ExcludesExpired() {
	s.truncateTokens()

	// create an active token
	_, err := s.repository.Create(s.ctx, userpat.PersonalAccessToken{
		UserID:     s.users[0].ID,
		OrgID:      s.orgs[0].ID,
		Title:      "active-token",
		SecretHash: "hashActive",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	})
	s.Require().NoError(err)

	// create an expired token
	_, err = s.repository.Create(s.ctx, userpat.PersonalAccessToken{
		UserID:     s.users[0].ID,
		OrgID:      s.orgs[0].ID,
		Title:      "expired-token",
		SecretHash: "hashExpired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	})
	s.Require().NoError(err)

	count, err := s.repository.CountActive(s.ctx, s.users[0].ID, s.orgs[0].ID)
	s.Require().NoError(err)
	s.Equal(int64(1), count)
}

func (s *UserPATRepositoryTestSuite) TestCountActive_FiltersByUserAndOrg() {
	s.truncateTokens()

	// token for user[0] in org[0]
	_, err := s.repository.Create(s.ctx, userpat.PersonalAccessToken{
		UserID:     s.users[0].ID,
		OrgID:      s.orgs[0].ID,
		Title:      "user0-org0",
		SecretHash: "hashU0O0",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	})
	s.Require().NoError(err)

	// token for user[1] in org[0]
	_, err = s.repository.Create(s.ctx, userpat.PersonalAccessToken{
		UserID:     s.users[1].ID,
		OrgID:      s.orgs[0].ID,
		Title:      "user1-org0",
		SecretHash: "hashU1O0",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	})
	s.Require().NoError(err)

	// token for user[0] in org[1]
	_, err = s.repository.Create(s.ctx, userpat.PersonalAccessToken{
		UserID:     s.users[0].ID,
		OrgID:      s.orgs[1].ID,
		Title:      "user0-org1",
		SecretHash: "hashU0O1",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	})
	s.Require().NoError(err)

	count, err := s.repository.CountActive(s.ctx, s.users[0].ID, s.orgs[0].ID)
	s.Require().NoError(err)
	s.Equal(int64(1), count)
}

func (s *UserPATRepositoryTestSuite) TestCountActive_MultipleTokens() {
	s.truncateTokens()

	for i := 0; i < 3; i++ {
		_, err := s.repository.Create(s.ctx, userpat.PersonalAccessToken{
			UserID:     s.users[0].ID,
			OrgID:      s.orgs[0].ID,
			Title:      fmt.Sprintf("multi-token-%d", i),
			SecretHash: fmt.Sprintf("hashMulti%d", i),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		})
		s.Require().NoError(err)
	}

	count, err := s.repository.CountActive(s.ctx, s.users[0].ID, s.orgs[0].ID)
	s.Require().NoError(err)
	s.Equal(int64(3), count)
}

func TestUserPATRepository(t *testing.T) {
	suite.Run(t, new(UserPATRepositoryTestSuite))
}
