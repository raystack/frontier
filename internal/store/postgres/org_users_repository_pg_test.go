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

// Runs the search against a real postgres to check that soft-deleted users,
// policies, and roles stay out of the result and out of the role filters.
type OrgUsersRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.OrgUsersRepository
	orgID      string
}

func (s *OrgUsersRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgUsersRepository(s.client)
}

func (s *OrgUsersRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgUsersRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/organization'), ('app/user')
		ON CONFLICT (name) DO NOTHING`)

	s.exec(`INSERT INTO organizations (name, title) VALUES ('ou-org', 'Org Users Org')`)
	s.orgID = s.scalar(`SELECT id FROM organizations WHERE name = 'ou-org'`)

	s.exec(`INSERT INTO users (name, email, title) VALUES
		('ou-a-live',           'ou-a@example.com', 'A Live'),
		('ou-b-deleted-user',   'ou-b@example.com', 'B Deleted User'),
		('ou-c-deleted-policy', 'ou-c@example.com', 'C Deleted Policy'),
		('ou-d-mixed',          'ou-d@example.com', 'D Mixed')`)
	s.exec(`UPDATE users SET deleted_at = now() WHERE name = 'ou-b-deleted-user'`)

	s.exec(`INSERT INTO roles (name, org_id) VALUES
		('ou-role-live', $1), ('ou-role-two', $1), ('ou-role-gone', $1)`, s.orgID)
	s.exec(`UPDATE roles SET deleted_at = now() WHERE name = 'ou-role-gone'`)

	s.policy("ou-role-live", "ou-a-live")
	s.policy("ou-role-live", "ou-b-deleted-user")
	s.policy("ou-role-live", "ou-c-deleted-policy")
	s.policy("ou-role-live", "ou-d-mixed")
	s.policy("ou-role-gone", "ou-d-mixed")
	s.policy("ou-role-two", "ou-d-mixed")
	s.softDeletePolicy("ou-role-live", "ou-c-deleted-policy")
	s.softDeletePolicy("ou-role-two", "ou-d-mixed")
}

func (s *OrgUsersRepositoryPGTestSuite) TearDownTest() {
	for _, table := range []string{"policies", "roles", "users", "organizations", "namespaces"} {
		s.exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
}

func (s *OrgUsersRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	if _, err := s.client.DB.ExecContext(s.ctx, query, args...); err != nil {
		s.T().Fatalf("%s: %v", query, err)
	}
}

func (s *OrgUsersRepositoryPGTestSuite) scalar(query string, args ...any) string {
	s.T().Helper()
	var out string
	if err := s.client.DB.QueryRowxContext(s.ctx, query, args...).Scan(&out); err != nil {
		s.T().Fatalf("%s: %v", query, err)
	}
	return out
}

func (s *OrgUsersRepositoryPGTestSuite) policy(role, userName string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = $1), $2, 'app/organization',
		        (SELECT id FROM users WHERE name = $3), 'app/user')`,
		role, s.orgID, userName)
}

func (s *OrgUsersRepositoryPGTestSuite) softDeletePolicy(role, userName string) {
	s.T().Helper()
	s.exec(`UPDATE policies SET deleted_at = now()
		WHERE role_id = (SELECT id FROM roles WHERE name = $1)
		  AND principal_id = (SELECT id FROM users WHERE name = $2)`,
		role, userName)
}

func (s *OrgUsersRepositoryPGTestSuite) names(filters ...rql.Filter) []string {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{
		Limit:   50,
		Filters: filters,
		Sort:    []rql.Sort{{Name: "name", Order: "asc"}},
	})
	s.Require().NoError(err)
	out := make([]string, 0, len(res.Users))
	for _, u := range res.Users {
		out = append(out, u.Name)
	}
	return out
}

func roleFilter(operator, roleName string) rql.Filter {
	return rql.Filter{Name: "role_names", Operator: operator, Value: roleName}
}

func (s *OrgUsersRepositoryPGTestSuite) TestSkipsSoftDeletedUsersAndPolicies() {
	s.Equal([]string{"ou-a-live", "ou-d-mixed"}, s.names())
}

func (s *OrgUsersRepositoryPGTestSuite) TestRolesComeOnlyFromLiveRows() {
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50, Filters: []rql.Filter{
		{Name: "name", Operator: "eq", Value: "ou-d-mixed"},
	}})
	s.Require().NoError(err)
	s.Require().Len(res.Users, 1)
	s.Equal([]string{"ou-role-live"}, res.Users[0].RoleNames)
}

func (s *OrgUsersRepositoryPGTestSuite) TestRoleFiltersIgnoreSoftDeletedRows() {
	s.Equal([]string{"ou-a-live", "ou-d-mixed"}, s.names(roleFilter("eq", "ou-role-live")))
	s.Empty(s.names(roleFilter("eq", "ou-role-two")), "granted only by a soft-deleted policy")
	s.Empty(s.names(roleFilter("eq", "ou-role-gone")), "the role is soft-deleted")
	s.Equal([]string{"ou-a-live", "ou-d-mixed"}, s.names(roleFilter("neq", "ou-role-two")))
	s.Equal([]string{"ou-a-live", "ou-d-mixed"}, s.names(roleFilter("neq", "ou-role-gone")))
}

func TestOrgUsersRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgUsersRepositoryPGTestSuite))
}
