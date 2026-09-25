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

// Every aggregate search binds a caller supplied id to a uuid column. A value
// the database cannot read as a uuid has to come back as bad input, so the
// handlers answer with invalid argument instead of internal.
type SearchInvalidUUIDTestSuite struct {
	suite.Suite
	ctx      context.Context
	client   *db.Client
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (s *SearchInvalidUUIDTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
}

func (s *SearchInvalidUUIDTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *SearchInvalidUUIDTestSuite) TestEverySearchReportsBadInput() {
	const bad = "not-a-uuid"
	query := func() *rql.Query { return &rql.Query{Limit: 10} }

	searches := map[string]func() error{
		"org users": func() error {
			_, err := postgres.NewOrgUsersRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"org projects": func() error {
			_, err := postgres.NewOrgProjectsRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"org service users": func() error {
			_, err := postgres.NewOrgServiceUserRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"org personal access tokens": func() error {
			_, err := postgres.NewOrgPATsRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"org invoices": func() error {
			_, err := postgres.NewOrgInvoicesRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"org tokens": func() error {
			_, err := postgres.NewOrgTokensRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"org service user credentials": func() error {
			_, err := postgres.NewOrgServiceUserCredentialsRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"project users": func() error {
			_, err := postgres.NewProjectUsersRepository(s.client).Search(s.ctx, bad, query())
			return err
		},
		"user organizations": func() error {
			_, err := postgres.NewUserOrgsRepository(s.client).Search(s.ctx, bad, &rql.Query{})
			return err
		},
		"user projects": func() error {
			_, err := postgres.NewUserProjectsRepository(s.client).Search(s.ctx, bad, bad, query())
			return err
		},
		// this one takes no id, so the bad value arrives through its id filter
		"org billing": func() error {
			_, err := postgres.NewOrgBillingRepository(s.client).Search(s.ctx, &rql.Query{
				Limit:   10,
				Filters: []rql.Filter{{Name: "id", Operator: "eq", Value: bad}},
			})
			return err
		},
	}

	for name, search := range searches {
		s.Run(name, func() {
			s.ErrorIs(search(), postgres.ErrBadInput)
		})
	}
}

func TestSearchInvalidUUID(t *testing.T) {
	suite.Run(t, new(SearchInvalidUUIDTestSuite))
}
