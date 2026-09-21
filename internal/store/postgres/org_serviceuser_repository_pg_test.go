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

/*
The SQL-string tests in org_serviceuser_repository_test.go pin the generated query,
but they would happily pass with the inner joins that caused the original bug. These
run the query against a real postgres with the shapes that actually broke:

	a-with-project      one policy on a live project
	b-org-policy-only   a policy, but on the org rather than a project
	c-no-policy         no policy at all
	d-deleted-project   a policy on a project that has been soft deleted
	e-two-roles         two policies, two roles, one and the same project
*/
type OrgServiceUserRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.OrgServiceUserRepository
	orgID      string
}

func (s *OrgServiceUserRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgServiceUserRepository(s.client)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgServiceUserRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO namespaces (name) VALUES ('app/project'), ('app/organization'), ('app/serviceuser')
		ON CONFLICT (name) DO NOTHING`)

	s.exec(`INSERT INTO organizations (name, title) VALUES ('su-search-org', 'SU Search Org')`)
	s.orgID = s.scalar(`SELECT id FROM organizations WHERE name = 'su-search-org'`)

	s.exec(`INSERT INTO projects (name, title, org_id) VALUES
		('su-proj-live', 'Live Project', $1), ('su-proj-gone', 'Gone Project', $1)`, s.orgID)
	s.exec(`UPDATE projects SET deleted_at = now() WHERE name = 'su-proj-gone'`)

	s.exec(`INSERT INTO serviceusers (org_id, title, created_at) VALUES
		($1, 'a-with-project',    '2026-01-01'),
		($1, 'b-org-policy-only', '2026-06-01'),
		($1, 'c-no-policy',       '2026-03-01'),
		($1, 'd-deleted-project', '2026-09-01'),
		($1, 'e-two-roles',       '2026-04-01')`, s.orgID)

	s.exec(`INSERT INTO roles (name, org_id) VALUES ('su-role-one', $1), ('su-role-two', $1)`, s.orgID)

	s.policy("su-role-one", "su-proj-live", "app/project", "a-with-project")
	s.orgPolicy("su-role-one", "b-org-policy-only")
	s.policy("su-role-one", "su-proj-gone", "app/project", "d-deleted-project")
	s.policy("su-role-one", "su-proj-live", "app/project", "e-two-roles")
	s.policy("su-role-two", "su-proj-live", "app/project", "e-two-roles")
}

func (s *OrgServiceUserRepositoryPGTestSuite) TearDownTest() {
	for _, table := range []string{"policies", "roles", "serviceusers", "projects", "organizations", "namespaces"} {
		s.exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
}

func (s *OrgServiceUserRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	if _, err := s.client.DB.ExecContext(s.ctx, query, args...); err != nil {
		s.T().Fatalf("%s: %v", query, err)
	}
}

func (s *OrgServiceUserRepositoryPGTestSuite) scalar(query string, args ...any) string {
	s.T().Helper()
	var out string
	if err := s.client.DB.QueryRowxContext(s.ctx, query, args...).Scan(&out); err != nil {
		s.T().Fatalf("%s: %v", query, err)
	}
	return out
}

func (s *OrgServiceUserRepositoryPGTestSuite) policy(role, project, resourceType, serviceUser string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = $1), (SELECT id FROM projects WHERE name = $2), $3,
		        (SELECT id FROM serviceusers WHERE title = $4), 'app/serviceuser')`,
		role, project, resourceType, serviceUser)
}

func (s *OrgServiceUserRepositoryPGTestSuite) orgPolicy(role, serviceUser string) {
	s.T().Helper()
	s.exec(`INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type)
		VALUES ((SELECT id FROM roles WHERE name = $1), $2, 'app/organization',
		        (SELECT id FROM serviceusers WHERE title = $3), 'app/serviceuser')`,
		role, s.orgID, serviceUser)
}

func (s *OrgServiceUserRepositoryPGTestSuite) titles(q *rql.Query) []string {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, s.orgID, q)
	s.Require().NoError(err)
	out := make([]string, 0, len(res.ServiceUsers))
	for _, su := range res.ServiceUsers {
		out = append(out, su.Title)
	}
	return out
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestReturnsEveryServiceUserInTheOrg() {
	s.Equal(
		[]string{"a-with-project", "b-org-policy-only", "c-no-policy", "d-deleted-project", "e-two-roles"},
		s.titles(&rql.Query{Limit: 50}),
	)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestProjectsPerServiceUser() {
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50})
	s.Require().NoError(err)

	got := map[string][]string{}
	for _, su := range res.ServiceUsers {
		names := []string{}
		for _, p := range su.Projects {
			names = append(names, p.Name)
		}
		got[su.Title] = names
	}

	s.Equal([]string{"su-proj-live"}, got["a-with-project"])
	s.Empty(got["b-org-policy-only"])
	s.Empty(got["c-no-policy"])
	s.Empty(got["d-deleted-project"])
	s.Equal([]string{"su-proj-live"}, got["e-two-roles"])
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestSortIsTheOneThatWasAskedFor() {
	s.Equal(
		[]string{"d-deleted-project", "b-org-policy-only", "e-two-roles", "c-no-policy", "a-with-project"},
		s.titles(&rql.Query{Limit: 50, Sort: []rql.Sort{{Name: "created_at", Order: "desc"}}}),
		"created_at desc should be newest first, not alphabetical",
	)
	s.Equal(
		[]string{"e-two-roles", "d-deleted-project", "c-no-policy", "b-org-policy-only", "a-with-project"},
		s.titles(&rql.Query{Limit: 50, Sort: []rql.Sort{{Name: "title", Order: "desc"}}}),
		"title desc should be descending",
	)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestFilterOnCreatedAt() {
	s.Equal(
		[]string{"b-org-policy-only", "d-deleted-project"},
		s.titles(&rql.Query{Limit: 50, Filters: []rql.Filter{
			{Name: "created_at", Operator: "gt", Value: "2026-05-01T00:00:00Z"},
		}}),
	)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestFilterOnProjects() {
	s.Equal(
		[]string{"a-with-project", "e-two-roles"},
		s.titles(&rql.Query{Limit: 50, Filters: []rql.Filter{
			{Name: "projects", Operator: "ilike", Value: "%Live%"},
		}}),
	)
	s.Empty(
		s.titles(&rql.Query{Limit: 50, Filters: []rql.Filter{
			{Name: "projects", Operator: "ilike", Value: "%nothing-matches%"},
		}}),
		"a filter that matches nothing must return nothing, not everything",
	)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestSearchCoversTitleAndProjects() {
	s.Equal([]string{"b-org-policy-only", "c-no-policy"}, s.titles(&rql.Query{Limit: 50, Search: "policy"}))
	s.Equal([]string{"a-with-project", "e-two-roles"}, s.titles(&rql.Query{Limit: 50, Search: "Live"}))
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestPagesDoNotOverlap() {
	sort := []rql.Sort{{Name: "title", Order: "asc"}}
	first := s.titles(&rql.Query{Limit: 2, Offset: 0, Sort: sort})
	second := s.titles(&rql.Query{Limit: 2, Offset: 2, Sort: sort})
	third := s.titles(&rql.Query{Limit: 2, Offset: 4, Sort: sort})

	s.Equal([]string{"a-with-project", "b-org-policy-only"}, first)
	s.Equal([]string{"c-no-policy", "d-deleted-project"}, second)
	s.Equal([]string{"e-two-roles"}, third)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestRejectsUnsupportedInput() {
	_, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50, GroupBy: []string{"title"}})
	s.ErrorIs(err, postgres.ErrBadInput)

	_, err = s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50, Sort: []rql.Sort{{Name: "state", Order: "asc"}}})
	s.ErrorIs(err, postgres.ErrBadInput)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestPaginationMetadata() {
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{})
	s.Require().NoError(err)
	s.Equal(50, res.Pagination.Limit, "a request with no limit reports the default that was applied")
	s.Equal(0, res.Pagination.Offset)
	s.Len(res.ServiceUsers, 5)

	res, err = s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 2, Offset: 2})
	s.Require().NoError(err)
	s.Equal(2, res.Pagination.Limit)
	s.Equal(2, res.Pagination.Offset)
	s.Len(res.ServiceUsers, 2)

	res, err = s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 2, Offset: 4})
	s.Require().NoError(err)
	s.Equal(2, res.Pagination.Limit)
	s.Equal(4, res.Pagination.Offset)
	s.Len(res.ServiceUsers, 1)
}

func (s *OrgServiceUserRepositoryPGTestSuite) TestSoftDeletedRowsAreSkipped() {
	s.Equal(
		[]string{"a-with-project", "b-org-policy-only", "c-no-policy", "d-deleted-project", "e-two-roles"},
		s.titles(&rql.Query{Limit: 50}),
	)

	s.exec(`UPDATE serviceusers SET deleted_at = now() WHERE title = 'c-no-policy'`)
	s.Equal(
		[]string{"a-with-project", "b-org-policy-only", "d-deleted-project", "e-two-roles"},
		s.titles(&rql.Query{Limit: 50}),
	)

	s.exec(`UPDATE policies SET deleted_at = now()
	        WHERE principal_id = (SELECT id FROM serviceusers WHERE title = 'a-with-project')`)
	res, err := s.repository.Search(s.ctx, s.orgID, &rql.Query{Limit: 50})
	s.Require().NoError(err)
	for _, su := range res.ServiceUsers {
		if su.Title == "a-with-project" {
			s.Empty(su.Projects, "a project granted by a soft-deleted policy must not be listed")
		}
	}
	s.Contains(s.titles(&rql.Query{Limit: 50}), "a-with-project")
}

func TestOrgServiceUserRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgServiceUserRepositoryPGTestSuite))
}
