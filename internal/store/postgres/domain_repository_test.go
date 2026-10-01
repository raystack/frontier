package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

type DomainRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.DomainRepository
	orgID      string
}

func (s *DomainRepositoryTestSuite) SetupSuite() {
	var err error

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, err = newTestClient()
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
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *DomainRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *DomainRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_DOMAINS),
	}
	return execQueries(context.TODO(), s.client, queries)
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

func (s *DomainRepositoryTestSuite) TestDeleteKeepsTheRow() {
	dmn, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "acme.com", Token: "t1"})
	s.Require().NoError(err)

	s.Require().NoError(s.repository.Delete(s.ctx, dmn.ID))
	s.Assert().Equal("1", scalarSQL(s.T(), s.ctx, s.client,
		"SELECT count(*) FROM domains WHERE id = $1 AND deleted_at IS NOT NULL", dmn.ID))

	_, err = s.repository.Get(s.ctx, dmn.ID)
	s.Assert().ErrorIs(err, domain.ErrNotExist)

	s.Assert().ErrorIs(s.repository.Delete(s.ctx, dmn.ID), domain.ErrNotExist)
	s.Assert().ErrorIs(s.repository.Delete(s.ctx, uuid.NewString()), domain.ErrNotExist)

	again, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "acme.com", Token: "t2"})
	s.Require().NoError(err)
	s.Assert().NotEqual(dmn.ID, again.ID)
}

func (s *DomainRepositoryTestSuite) TestDeleteExpiredDomainRequestsSkipsSoftDeletedDomains() {
	stale, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "stale.example.com", Token: "t1"})
	s.Require().NoError(err)
	deleted, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "deleted.example.com", Token: "t2"})
	s.Require().NoError(err)
	fresh, err := s.repository.Create(s.ctx, domain.Domain{OrgID: s.orgID, Name: "fresh.example.com", Token: "t3"})
	s.Require().NoError(err)

	s.Require().NoError(s.repository.Delete(s.ctx, deleted.ID))
	execSQL(s.T(), s.ctx, s.client,
		"UPDATE domains SET created_at = now() - interval '8 days' WHERE id IN ($1, $2)", stale.ID, deleted.ID)

	s.Require().NoError(s.repository.DeleteExpiredDomainRequests(s.ctx))

	var remaining []string
	s.Require().NoError(s.client.SelectContext(s.ctx, &remaining, "SELECT id FROM domains"))
	s.Assert().ElementsMatch([]string{deleted.ID, fresh.ID}, remaining)
}

func TestDomainRepository(t *testing.T) {
	suite.Run(t, new(DomainRepositoryTestSuite))
}
