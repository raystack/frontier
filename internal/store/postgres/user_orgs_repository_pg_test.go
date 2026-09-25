package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the search against a real postgres to check that soft-deleted
// organizations, policies, roles, users, and projects stay out of the answer.
type UserOrgsRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.UserOrgsRepository
}

func (s *UserOrgsRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewUserOrgsRepository(s.client)
}

func (s *UserOrgsRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserOrgsRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/organization'), ('app/user') ON CONFLICT (name) DO NOTHING`)

	s.exec(`INSERT INTO organizations (name, title) VALUES
		('uo-live',    'Live Org'),
		('uo-gone',    'Gone Org'),
		('uo-revoked', 'Revoked Org')`)

	s.exec(`INSERT INTO users (name, email, title) VALUES
		('uo-dana', 'uo-dana@example.com', 'Dana'),
		('uo-evan', 'uo-evan@example.com', 'Evan')`)

	s.exec(`INSERT INTO projects (name, title, org_id, state) VALUES
		('uo-proj-one', 'One', $1, 'enabled'),
		('uo-proj-two', 'Two', $1, 'enabled'),
		('uo-proj-old', 'Old', $1, 'enabled')`, s.orgID("uo-live"))

	s.exec(`INSERT INTO roles (name, title, org_id) VALUES
		('uo-role',      'Live Role', $1),
		('uo-role-gone', 'Gone Role', $1)`, s.orgID("uo-live"))

	s.policy("uo-live", "uo-dana", "uo-role")
	s.policy("uo-live", "uo-dana", "uo-role-gone")
	s.policy("uo-gone", "uo-dana", "uo-role")
	s.policy("uo-revoked", "uo-dana", "uo-role")
	s.policy("uo-live", "uo-evan", "uo-role")

	s.exec(`UPDATE projects SET deleted_at = now() WHERE name = 'uo-proj-old'`)
	s.exec(`UPDATE organizations SET deleted_at = now() WHERE name = 'uo-gone'`)
	s.exec(`UPDATE roles SET deleted_at = now() WHERE name = 'uo-role-gone'`)
	s.exec(`UPDATE users SET deleted_at = now() WHERE name = 'uo-evan'`)
	s.softDeletePolicy("uo-revoked", "uo-dana")
}

func (s *UserOrgsRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_POLICIES, postgres.TABLE_ROLES, postgres.TABLE_USERS,
		postgres.TABLE_PROJECTS, postgres.TABLE_ORGANIZATIONS, postgres.TABLE_NAMESPACES} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserOrgsRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *UserOrgsRepositoryPGTestSuite) orgID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *UserOrgsRepositoryPGTestSuite) userID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM users WHERE name = $1`, name)
}

func (s *UserOrgsRepositoryPGTestSuite) policy(orgName, userName, roleName string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = $1), $2, 'app/organization', $3, 'app/user')`,
		roleName, s.orgID(orgName), s.userID(userName))
}

func (s *UserOrgsRepositoryPGTestSuite) softDeletePolicy(orgName, userName string) {
	s.T().Helper()
	res := execSQL(s.T(), s.ctx, s.client, `UPDATE policies SET deleted_at = now()
		WHERE resource_id = $1 AND principal_id = $2`, s.orgID(orgName), s.userID(userName))
	n, err := res.RowsAffected()
	s.Require().NoError(err)
	s.Require().EqualValues(1, n, "soft delete of %s for %s", orgName, userName)
}

func (s *UserOrgsRepositoryPGTestSuite) search(userName string) []userorgs.AggregatedUserOrganization {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.userID(userName), &rql.Query{})
	s.Require().NoError(err)
	return res.Organizations
}

func (s *UserOrgsRepositoryPGTestSuite) TestSkipsSoftDeletedOrgsAndGrants() {
	orgs := s.search("uo-dana")
	s.Require().Len(orgs, 1)
	s.Equal("uo-live", orgs[0].OrgName)
}

func (s *UserOrgsRepositoryPGTestSuite) TestRolesComeOnlyFromLiveRows() {
	orgs := s.search("uo-dana")
	s.Require().Len(orgs, 1)
	s.Equal([]string{"uo-role"}, orgs[0].RoleNames)
	s.Equal([]string{"Live Role"}, orgs[0].RoleTitles)
}

func (s *UserOrgsRepositoryPGTestSuite) TestProjectCountSkipsSoftDeletedProjects() {
	orgs := s.search("uo-dana")
	s.Require().Len(orgs, 1)
	s.EqualValues(2, orgs[0].ProjectCount)
}

func (s *UserOrgsRepositoryPGTestSuite) TestSoftDeletedUserHasNoOrgs() {
	s.Empty(s.search("uo-evan"))
}

func TestUserOrgsRepositoryPG(t *testing.T) {
	suite.Run(t, new(UserOrgsRepositoryPGTestSuite))
}
