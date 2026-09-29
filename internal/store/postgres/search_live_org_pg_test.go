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

// A search scoped to an organization must report nothing once that organization
// is soft-deleted, even though the rows beneath it are untouched. Each search is
// seeded with the same shape under a live organization and a deleted one.
type SearchLiveOrgTestSuite struct {
	suite.Suite
	ctx    context.Context
	client *db.Client
}

func (s *SearchLiveOrgTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
}

func (s *SearchLiveOrgTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *SearchLiveOrgTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/organization'), ('app/project'), ('app/user')
		ON CONFLICT (name) DO NOTHING`)
	s.exec(`INSERT INTO organizations (name, title) VALUES ('slo-live', 'Live'), ('slo-gone', 'Gone')`)
	// the avatar is set because the user projects search cannot scan a null one
	s.exec(`INSERT INTO users (name, email, title, avatar) VALUES ('slo-user', 'slo-user@example.com', 'User', 'slo.png')`)

	for _, org := range []string{"slo-live", "slo-gone"} {
		s.exec(`INSERT INTO roles (name, title, org_id) VALUES ($1, 'Role', $2)`, "slo-role-"+org, s.orgID(org))
		s.exec(`INSERT INTO projects (name, title, org_id, state) VALUES ($1, 'Project', $2, 'enabled')`,
			"slo-proj-"+org, s.orgID(org))
		s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
			VALUES ((SELECT id FROM roles WHERE name = $1), $2, 'app/organization', $3, 'app/user')`,
			"slo-role-"+org, s.orgID(org), s.userID())
		s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
			VALUES ((SELECT id FROM roles WHERE name = $1), (SELECT id FROM projects WHERE name = $2),
			        'app/project', $3, 'app/user')`,
			"slo-role-"+org, "slo-proj-"+org, s.userID())
	}

	s.exec(`UPDATE organizations SET deleted_at = now() WHERE name = 'slo-gone'`)
}

func (s *SearchLiveOrgTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_POLICIES, postgres.TABLE_ROLES,
		postgres.TABLE_PROJECTS, postgres.TABLE_USERS, postgres.TABLE_ORGANIZATIONS, postgres.TABLE_NAMESPACES} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *SearchLiveOrgTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *SearchLiveOrgTestSuite) orgID(name string) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *SearchLiveOrgTestSuite) userID() string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, `SELECT id FROM users WHERE name = 'slo-user'`)
}

// counts returns how many rows each search reports for the given organization.
func (s *SearchLiveOrgTestSuite) counts(orgName string) map[string]int {
	s.T().Helper()
	orgID := s.orgID(orgName)
	query := func() *rql.Query { return &rql.Query{Limit: 50} }

	users, err := postgres.NewOrgUsersRepository(s.client).Search(s.ctx, orgID, query())
	s.Require().NoError(err)
	projects, err := postgres.NewOrgProjectsRepository(s.client).Search(s.ctx, orgID, query())
	s.Require().NoError(err)
	userProjects, err := postgres.NewUserProjectsRepository(s.client).Search(s.ctx, s.userID(), orgID, query())
	s.Require().NoError(err)

	return map[string]int{
		"org users":     len(users.Users),
		"org projects":  len(projects.Projects),
		"user projects": len(userProjects.Projects),
	}
}

func (s *SearchLiveOrgTestSuite) TestLiveOrgReportsItsRows() {
	for name, n := range s.counts("slo-live") {
		s.Positive(n, name)
	}
}

func (s *SearchLiveOrgTestSuite) TestSoftDeletedOrgReportsNothing() {
	for name, n := range s.counts("slo-gone") {
		s.Zero(n, name)
	}
}

func TestSearchLiveOrg(t *testing.T) {
	suite.Run(t, new(SearchLiveOrgTestSuite))
}
