package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the search against a real postgres to check that revoked credentials and
// soft-deleted service users and organizations stay out of the answer.
type OrgServiceUserCredentialsRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.OrgServiceUserCredentialsRepository
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgServiceUserCredentialsRepository(s.client)
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES
		('osc-live', 'Live Org'), ('osc-gone', 'Gone Org')`)

	s.exec(`INSERT INTO serviceusers (org_id, title) VALUES
		($1, 'osc-su-live'), ($1, 'osc-su-gone'), ($2, 'osc-su-other')`,
		s.orgID("osc-live"), s.orgID("osc-gone"))

	s.credential("osc-su-live", "osc-cred-live")
	s.credential("osc-su-live", "osc-cred-revoked")
	s.credential("osc-su-gone", "osc-cred-orphan")
	s.credential("osc-su-other", "osc-cred-otherorg")

	s.exec(`UPDATE serviceuser_credentials SET deleted_at = now() WHERE title = 'osc-cred-revoked'`)
	s.exec(`UPDATE serviceusers SET deleted_at = now() WHERE title = 'osc-su-gone'`)
	s.exec(`UPDATE organizations SET deleted_at = now() WHERE name = 'osc-gone'`)
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_SERVICEUSERCREDENTIALS, postgres.TABLE_SERVICEUSER,
		postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) orgID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) credential(serviceUser, title string) {
	s.T().Helper()
	s.exec(`INSERT INTO serviceuser_credentials (serviceuser_id, title)
		VALUES ((SELECT id FROM serviceusers WHERE title = $1), $2)`, serviceUser, title)
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) titles(orgName, text string) []string {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.orgID(orgName), &rql.Query{Limit: 50, Search: text})
	s.Require().NoError(err)
	out := make([]string, 0, len(res.Credentials))
	for _, c := range res.Credentials {
		out = append(out, c.Title)
	}
	return out
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) TestSkipsRevokedCredentialsAndDeletedServiceUsers() {
	s.Equal([]string{"osc-cred-live"}, s.titles("osc-live", ""))
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) TestSoftDeletedOrgHasNoCredentials() {
	s.Empty(s.titles("osc-gone", ""))
}

func (s *OrgServiceUserCredentialsRepositoryPGTestSuite) TestSearchMatchesLiveRowsOnly() {
	s.Equal([]string{"osc-cred-live"}, s.titles("osc-live", "cred-live"))
	s.Empty(s.titles("osc-live", "revoked"), "the credential is soft-deleted")
	s.Empty(s.titles("osc-live", "orphan"), "the service user is soft-deleted")
}

func TestOrgServiceUserCredentialsRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgServiceUserCredentialsRepositoryPGTestSuite))
}
