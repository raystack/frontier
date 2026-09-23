package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

// Runs the search against a real postgres to check that soft-deleted projects,
// policies, and users stay out of the result and out of the member counts.
type OrgProjectsRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.OrgProjectsRepository
	orgID      string
}

func (s *OrgProjectsRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgProjectsRepository(s.client)
}

func (s *OrgProjectsRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgProjectsRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/project'), ('app/user') ON CONFLICT (name) DO NOTHING`)

	s.exec(`INSERT INTO organizations (name, title) VALUES ('op-org', 'Org Projects Org')`)
	s.orgID = s.scalar(`SELECT id FROM organizations WHERE name = 'op-org'`)

	s.exec(`INSERT INTO projects (name, title, org_id) VALUES
		('op-a-live',           'A Live',           $1),
		('op-b-deleted',        'B Deleted',        $1),
		('op-c-deleted-policy', 'C Deleted Policy', $1),
		('op-d-mixed',          'D Mixed',          $1)`, s.orgID)
	s.exec(`UPDATE projects SET deleted_at = now() WHERE name = 'op-b-deleted'`)

	s.exec(`INSERT INTO users (name, email) VALUES
		('op-u-live',    'op-u-live@example.com'),
		('op-u-deleted', 'op-u-deleted@example.com'),
		('op-u-gone',    'op-u-gone@example.com')`)
	s.exec(`UPDATE users SET deleted_at = now() WHERE name = 'op-u-deleted'`)

	s.exec(`INSERT INTO roles (name, org_id) VALUES ('op-role', $1)`, s.orgID)

	s.policy("op-a-live", "op-u-live")
	s.policy("op-b-deleted", "op-u-live")
	s.policy("op-c-deleted-policy", "op-u-live")
	s.policy("op-d-mixed", "op-u-live")
	s.policy("op-d-mixed", "op-u-deleted")
	s.policy("op-d-mixed", "op-u-gone")
	s.softDeletePolicy("op-c-deleted-policy", "op-u-live")
	s.softDeletePolicy("op-d-mixed", "op-u-gone")
}

func (s *OrgProjectsRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_POLICIES, postgres.TABLE_ROLES, postgres.TABLE_USERS,
		postgres.TABLE_PROJECTS, postgres.TABLE_ORGANIZATIONS, postgres.TABLE_NAMESPACES} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgProjectsRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *OrgProjectsRepositoryPGTestSuite) scalar(query string, args ...any) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *OrgProjectsRepositoryPGTestSuite) projectID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM projects WHERE name = $1`, name)
}

func (s *OrgProjectsRepositoryPGTestSuite) userID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM users WHERE name = $1`, name)
}

func (s *OrgProjectsRepositoryPGTestSuite) policy(projectName, userName string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = 'op-role'), $1, 'app/project', $2, 'app/user')`,
		s.projectID(projectName), s.userID(userName))
}

func (s *OrgProjectsRepositoryPGTestSuite) softDeletePolicy(projectName, userName string) {
	s.T().Helper()
	res := execSQL(s.T(), s.ctx, s.client, `UPDATE policies SET deleted_at = now()
		WHERE resource_id = $1 AND principal_id = $2`, s.projectID(projectName), s.userID(userName))
	n, err := res.RowsAffected()
	s.Require().NoError(err)
	s.Require().EqualValues(1, n, "soft delete of %s for %s", projectName, userName)
}

func (s *OrgProjectsRepositoryPGTestSuite) search() []orgprojects.AggregatedProject {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{
		Limit: 50,
		Sort:  []rql.Sort{{Name: "name", Order: "asc"}},
	})
	s.Require().NoError(err)
	return res.Projects
}

func (s *OrgProjectsRepositoryPGTestSuite) TestSkipsSoftDeletedProjectsAndPolicies() {
	names := []string{}
	for _, p := range s.search() {
		names = append(names, p.Name)
	}
	s.Equal([]string{"op-a-live", "op-d-mixed"}, names)
}

func (s *OrgProjectsRepositoryPGTestSuite) TestMembersComeOnlyFromLiveRows() {
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50, Filters: []rql.Filter{
		{Name: "name", Operator: "eq", Value: "op-d-mixed"},
	}})
	s.Require().NoError(err)
	s.Require().Len(res.Projects, 1)
	s.EqualValues(1, res.Projects[0].MemberCount)
	s.Equal([]string{s.userID("op-u-live")}, res.Projects[0].UserIDs)
}

func TestOrgProjectsRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgProjectsRepositoryPGTestSuite))
}
