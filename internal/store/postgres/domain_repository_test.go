package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

type DomainRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.DomainRepository
	orgID      string
}

func (s *DomainRepositoryTestSuite) SetupSuite() {
	var err error

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewDomainRepository(logger, s.client)

	orgs, err := bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.orgID = orgs[0].ID
}

func (s *DomainRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *DomainRepositoryTestSuite) TestSkipsSoftDeletedDomains() {
	live, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "live.example.com", Token: "t1"})
	s.Require().NoError(err)
	deleted, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "gone.example.com", Token: "t2"})
	s.Require().NoError(err)

	_, err = s.client.ExecContext(s.ctx, "UPDATE domains SET deleted_at = now() WHERE id = $1", deleted.ID)
	s.Require().NoError(err)

	_, err = s.repository.Get(s.ctx, deleted.ID)
	s.Assert().ErrorIs(err, domain.ErrNotExist)

	_, err = s.repository.Get(s.ctx, uuid.NewString())
	s.Assert().ErrorIs(err, domain.ErrNotExist)

	got, err := s.repository.List(s.ctx, domain.Filter{OrgID: s.orgID})
	s.Assert().NoError(err)
	s.Require().Len(got, 1)
	s.Assert().Equal(live.ID, got[0].ID)

	_, err = s.repository.Update(s.ctx, domain.Domain{ID: deleted.ID, Token: "t3", State: domain.Pending})
	s.Assert().ErrorIs(err, domain.ErrNotExist)
}

func TestDomainRepository(t *testing.T) {
	suite.Run(t, new(DomainRepositoryTestSuite))
}
