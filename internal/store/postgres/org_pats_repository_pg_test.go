package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

type OrgPATsRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.OrgPATsRepository
	orgID      string
}

func (s *OrgPATsRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgPATsRepository(s.client)
}

func (s *OrgPATsRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgPATsRepositoryPGTestSuite) SetupTest() {
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO organizations (name, title) VALUES ('op-org', 'Org PATs Org')`)
	s.orgID = scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = 'op-org'`)

	execSQL(s.T(), s.ctx, s.client, `INSERT INTO users (name, email, title) VALUES
		('op-uma',  'op-uma@example.com',  'Uma'),
		('op-gary', 'op-gary@example.com', 'Gary')`)
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO user_pats (user_id, org_id, title, secret_hash, expires_at) VALUES
		((SELECT id FROM users WHERE name = 'op-uma'),  $1, 'deploy-token', 'op-hash-uma',  now() + interval '1 day'),
		((SELECT id FROM users WHERE name = 'op-gary'), $1, 'backup-token', 'op-hash-gary', now() + interval '1 day')`,
		s.orgID)
	execSQL(s.T(), s.ctx, s.client, `UPDATE users SET deleted_at = now() WHERE name = 'op-gary'`)
}

func (s *OrgPATsRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_USER_PATS, postgres.TABLE_USERS, postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgPATsRepositoryPGTestSuite) search(text string) ([]string, int64) {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50, Search: text})
	s.Require().NoError(err)
	titles := make([]string, 0, len(res.PATs))
	for _, pat := range res.PATs {
		titles = append(titles, pat.Title)
	}
	return titles, res.Pagination.TotalCount
}

func (s *OrgPATsRepositoryPGTestSuite) TestSkipsPATsOfSoftDeletedUsers() {
	titles, total := s.search("")
	s.Equal([]string{"deploy-token"}, titles)
	s.Equal(int64(1), total)
}

func (s *OrgPATsRepositoryPGTestSuite) TestSearchMatchesLiveOwnersOnly() {
	titles, _ := s.search("Uma")
	s.Equal([]string{"deploy-token"}, titles)

	titles, total := s.search("Gary")
	s.Empty(titles, "the owner is soft-deleted")
	s.Zero(total)
}

func TestOrgPATsRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgPATsRepositoryPGTestSuite))
}
