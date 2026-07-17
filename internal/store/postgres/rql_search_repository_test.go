package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// The rql search repositories must surface transaction failures instead of
// returning an empty result with a nil error. A canceled context makes the
// transaction begin fail before any query runs.
type RQLSearchRepositoryTestSuite struct {
	suite.Suite
	client   *db.Client
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (s *RQLSearchRepositoryTestSuite) SetupSuite() {
	var err error

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *RQLSearchRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func (s *RQLSearchRepositoryTestSuite) TestOrgBillingSearchReturnsTxnError() {
	_, err := postgres.NewOrgBillingRepository(s.client).Search(canceledContext(), &rql.Query{Limit: 10})
	s.Error(err)
}

func (s *RQLSearchRepositoryTestSuite) TestOrgProjectsSearchReturnsTxnError() {
	_, err := postgres.NewOrgProjectsRepository(s.client).Search(canceledContext(), "00000000-0000-0000-0000-000000000000", &rql.Query{Limit: 10})
	s.Error(err)
}

func (s *RQLSearchRepositoryTestSuite) TestOrgUsersSearchReturnsTxnError() {
	_, err := postgres.NewOrgUsersRepository(s.client).Search(canceledContext(), "00000000-0000-0000-0000-000000000000", &rql.Query{Limit: 10})
	s.Error(err)
}

func (s *RQLSearchRepositoryTestSuite) TestProjectUsersSearchReturnsTxnError() {
	_, err := postgres.NewProjectUsersRepository(s.client).Search(canceledContext(), "00000000-0000-0000-0000-000000000000", &rql.Query{Limit: 10})
	s.Error(err)
}

func TestRQLSearchRepository(t *testing.T) {
	suite.Run(t, new(RQLSearchRepositoryTestSuite))
}
