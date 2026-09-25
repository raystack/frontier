package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/aggregates/userprojects"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the search against a real postgres to check that soft-deleted projects,
// policies, and users stay out of the listing and out of the member arrays.
type UserProjectsRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.UserProjectsRepository
	orgID      string
}

func (s *UserProjectsRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewUserProjectsRepository(s.client)
}

func (s *UserProjectsRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserProjectsRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/project'), ('app/user') ON CONFLICT (name) DO NOTHING`)

	s.exec(`INSERT INTO organizations (name, title) VALUES ('up-org', 'User Projects Org')`)
	s.orgID = s.scalar(`SELECT id FROM organizations WHERE name = 'up-org'`)

	s.exec(`INSERT INTO projects (name, title, org_id) VALUES
		('up-atlas',  'Atlas',  $1),
		('up-beacon', 'Beacon', $1),
		('up-cargo',  'Cargo',  $1)`, s.orgID)
	s.exec(`UPDATE projects SET deleted_at = now() WHERE name = 'up-beacon'`)

	s.exec(`INSERT INTO users (name, email, title, avatar) VALUES
		('up-dana', 'up-dana@example.com', 'Dana', 'dana.png'),
		('up-evan', 'up-evan@example.com', 'Evan', 'evan.png'),
		('up-finn', 'up-finn@example.com', 'Finn', 'finn.png'),
		('up-gale', 'up-gale@example.com', 'Gale', 'gale.png')`)
	s.exec(`UPDATE users SET deleted_at = now() WHERE name = 'up-finn'`)

	s.exec(`INSERT INTO roles (name, org_id) VALUES ('up-role', $1)`, s.orgID)

	for _, project := range []string{"up-atlas", "up-beacon", "up-cargo"} {
		s.policy(project, "up-dana")
	}
	s.policy("up-atlas", "up-evan")
	s.policy("up-atlas", "up-finn")
	s.policy("up-atlas", "up-gale")

	s.softDeletePolicy("up-cargo", "up-dana")
	s.softDeletePolicy("up-atlas", "up-gale")
}

func (s *UserProjectsRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_POLICIES, postgres.TABLE_ROLES, postgres.TABLE_USERS,
		postgres.TABLE_PROJECTS, postgres.TABLE_ORGANIZATIONS, postgres.TABLE_NAMESPACES} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserProjectsRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *UserProjectsRepositoryPGTestSuite) scalar(query string, args ...any) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *UserProjectsRepositoryPGTestSuite) projectID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM projects WHERE name = $1`, name)
}

func (s *UserProjectsRepositoryPGTestSuite) userID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM users WHERE name = $1`, name)
}

func (s *UserProjectsRepositoryPGTestSuite) policy(projectName, userName string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = 'up-role'), $1, 'app/project', $2, 'app/user')`,
		s.projectID(projectName), s.userID(userName))
}

func (s *UserProjectsRepositoryPGTestSuite) softDeletePolicy(projectName, userName string) {
	s.T().Helper()
	res := execSQL(s.T(), s.ctx, s.client, `UPDATE policies SET deleted_at = now()
		WHERE resource_id = $1 AND principal_id = $2`, s.projectID(projectName), s.userID(userName))
	n, err := res.RowsAffected()
	s.Require().NoError(err)
	s.Require().EqualValues(1, n, "soft delete of %s for %s", projectName, userName)
}

func (s *UserProjectsRepositoryPGTestSuite) search(text string) []userprojects.AggregatedProject {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.userID("up-dana"), s.orgID, &rql.Query{Limit: 50, Search: text})
	s.Require().NoError(err)
	return res.Projects
}

func (s *UserProjectsRepositoryPGTestSuite) names(projects []userprojects.AggregatedProject) []string {
	s.T().Helper()
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, p.ProjectName)
	}
	return out
}

func (s *UserProjectsRepositoryPGTestSuite) TestSkipsSoftDeletedProjectsAndPolicies() {
	s.Equal([]string{"up-atlas"}, s.names(s.search("")))
}

func (s *UserProjectsRepositoryPGTestSuite) TestMembersComeOnlyFromLiveRows() {
	projects := s.search("")
	s.Require().Len(projects, 1)
	s.Equal(sortedIDs([]string{s.userID("up-dana"), s.userID("up-evan")}), projects[0].UserIDs)
	s.Equal([]string{"up-dana", "up-evan"}, projects[0].UserNames)
	s.Equal([]string{"Dana", "Evan"}, projects[0].UserTitles)
	s.Equal([]string{"dana.png", "evan.png"}, projects[0].UserAvatars)
}

func (s *UserProjectsRepositoryPGTestSuite) TestSearchMatchesLiveProjectsOnly() {
	s.Equal([]string{"up-atlas"}, s.names(s.search("atlas")))
	s.Empty(s.names(s.search("beacon")), "the project is soft-deleted")
	s.Empty(s.names(s.search("cargo")), "the grant on it is soft-deleted")
}

func sortedIDs(ids []string) []string {
	out := append([]string(nil), ids...)
	slices.Sort(out)
	return out
}

func TestUserProjectsRepositoryPG(t *testing.T) {
	suite.Run(t, new(UserProjectsRepositoryPGTestSuite))
}
