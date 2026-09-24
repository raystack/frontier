package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/aggregates/projectusers"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the search against a real postgres to check that soft-deleted users,
// policies, roles, and projects stay out of the member list.
type ProjectUsersRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.ProjectUsersRepository
	orgID      string
}

func (s *ProjectUsersRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewProjectUsersRepository(s.client)
}

func (s *ProjectUsersRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProjectUsersRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/project'), ('app/user') ON CONFLICT (name) DO NOTHING`)

	s.exec(`INSERT INTO organizations (name, title) VALUES ('pu-org', 'Project Users Org')`)
	s.orgID = s.scalar(`SELECT id FROM organizations WHERE name = 'pu-org'`)

	s.exec(`INSERT INTO projects (name, title, org_id) VALUES
		('pu-live', 'Live Project', $1), ('pu-gone', 'Gone Project', $1)`, s.orgID)

	s.exec(`INSERT INTO users (name, email, title) VALUES
		('pu-alice', 'pu-alice@example.com', 'Alice'),
		('pu-bob',   'pu-bob@example.com',   'Bob'),
		('pu-cara',  'pu-cara@example.com',  'Cara'),
		('pu-dave',  'pu-dave@example.com',  'Dave')`)

	s.exec(`INSERT INTO roles (name, org_id) VALUES ('pu-role', $1), ('pu-role-gone', $1)`, s.orgID)

	s.policy("pu-live", "pu-alice", "pu-role")
	s.policy("pu-live", "pu-bob", "pu-role")
	s.policy("pu-live", "pu-cara", "pu-role")
	s.policy("pu-live", "pu-dave", "pu-role-gone")
	s.policy("pu-gone", "pu-alice", "pu-role")

	s.exec(`UPDATE projects SET deleted_at = now() WHERE name = 'pu-gone'`)
	s.exec(`UPDATE users SET deleted_at = now() WHERE name = 'pu-bob'`)
	s.exec(`UPDATE roles SET deleted_at = now() WHERE name = 'pu-role-gone'`)
	s.softDeletePolicy("pu-live", "pu-cara")
}

func (s *ProjectUsersRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_POLICIES, postgres.TABLE_ROLES, postgres.TABLE_USERS,
		postgres.TABLE_PROJECTS, postgres.TABLE_ORGANIZATIONS, postgres.TABLE_NAMESPACES} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProjectUsersRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *ProjectUsersRepositoryPGTestSuite) scalar(query string, args ...any) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *ProjectUsersRepositoryPGTestSuite) projectID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM projects WHERE name = $1`, name)
}

func (s *ProjectUsersRepositoryPGTestSuite) userID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM users WHERE name = $1`, name)
}

func (s *ProjectUsersRepositoryPGTestSuite) policy(projectName, userName, roleName string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = $1), $2, 'app/project', $3, 'app/user')`,
		roleName, s.projectID(projectName), s.userID(userName))
}

func (s *ProjectUsersRepositoryPGTestSuite) softDeletePolicy(projectName, userName string) {
	s.T().Helper()
	res := execSQL(s.T(), s.ctx, s.client, `UPDATE policies SET deleted_at = now()
		WHERE resource_id = $1 AND principal_id = $2`, s.projectID(projectName), s.userID(userName))
	n, err := res.RowsAffected()
	s.Require().NoError(err)
	s.Require().EqualValues(1, n, "soft delete of %s for %s", projectName, userName)
}

func (s *ProjectUsersRepositoryPGTestSuite) search(projectName, text string) []projectusers.AggregatedUser {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.projectID(projectName), &rql.Query{Limit: 50, Search: text})
	s.Require().NoError(err)
	return res.Users
}

func (s *ProjectUsersRepositoryPGTestSuite) names(users []projectusers.AggregatedUser) []string {
	s.T().Helper()
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.Name)
	}
	return out
}

func (s *ProjectUsersRepositoryPGTestSuite) TestSkipsSoftDeletedMembersGrantsAndRoles() {
	users := s.search("pu-live", "")
	s.Require().Len(users, 1)
	s.Equal("pu-alice", users[0].Name)
	s.Equal([]string{"pu-role"}, users[0].RoleNames)
}

func (s *ProjectUsersRepositoryPGTestSuite) TestSoftDeletedProjectHasNoMembers() {
	s.Empty(s.names(s.search("pu-gone", "")))
}

func (s *ProjectUsersRepositoryPGTestSuite) TestSearchMatchesLiveRowsOnly() {
	s.Equal([]string{"pu-alice"}, s.names(s.search("pu-live", "alice")))
	s.Empty(s.names(s.search("pu-live", "bob")), "the user is soft-deleted")
	s.Empty(s.names(s.search("pu-live", "cara")), "the grant is soft-deleted")
	s.Empty(s.names(s.search("pu-live", "dave")), "the role is soft-deleted")
}

func TestProjectUsersRepositoryPG(t *testing.T) {
	suite.Run(t, new(ProjectUsersRepositoryPGTestSuite))
}
