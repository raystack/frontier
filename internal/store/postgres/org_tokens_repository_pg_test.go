package postgres_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/aggregates/orgtokens"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

type OrgTokensRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.OrgTokensRepository
	orgID      string
}

func (s *OrgTokensRepositoryPGTestSuite) SetupSuite() {
	var err error
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgTokensRepository(s.client)
}

func (s *OrgTokensRepositoryPGTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgTokensRepositoryPGTestSuite) SetupTest() {
	s.exec(`INSERT INTO organizations (name, title) VALUES
		('ot-org',      'Org Tokens Org'),
		('ot-org-gone', 'Gone Org')`)
	s.orgID = s.orgIDOf("ot-org")

	s.exec(`INSERT INTO billing_customers (org_id, provider_id, name, email) VALUES
		($1, 'ot-cust-live', 'Live Account', 'live@example.com'),
		($1, 'ot-cust-gone', 'Gone Account', 'gone@example.com'),
		($2, 'ot-cust-of-gone-org', 'Gone Org Account', 'goneorg@example.com')`,
		s.orgID, s.orgIDOf("ot-org-gone"))

	s.exec(`INSERT INTO users (name, email, title) VALUES
		('ot-uma',  'ot-uma@example.com',  'Uma'),
		('ot-gary', 'ot-gary@example.com', 'Gary')`)

	s.exec(`INSERT INTO billing_transactions (account_id, type, source, amount, description, user_id) VALUES
		($1, 'credit', 'system', 100, 'Live top up',         $2),
		($1, 'credit', 'system', 50,  'Top up by gone user', $3),
		($1, 'credit', 'system', 40,  'No user top up',      NULL),
		($1, 'credit', 'system', 30,  'Empty user top up',   ''),
		($1, 'credit', 'system', 25,  'Deleted top up',      $2),
		($4, 'credit', 'system', 10,  'Gone account top up', $2),
		($5, 'credit', 'system', 5,   'Gone org top up',     $2)`,
		s.customerID("ot-cust-live"), s.userID("ot-uma"), s.userID("ot-gary"),
		s.customerID("ot-cust-gone"), s.customerID("ot-cust-of-gone-org"))

	s.exec(`UPDATE billing_transactions SET deleted_at = now() WHERE description = 'Deleted top up'`)
	s.exec(`UPDATE billing_customers SET deleted_at = now() WHERE provider_id = 'ot-cust-gone'`)
	s.exec(`UPDATE organizations SET deleted_at = now() WHERE name = 'ot-org-gone'`)
	s.exec(`UPDATE users SET deleted_at = now() WHERE name = 'ot-gary'`)
}

func (s *OrgTokensRepositoryPGTestSuite) TearDownTest() {
	queries := []string{}
	for _, table := range []string{postgres.TABLE_BILLING_TRANSACTIONS, postgres.TABLE_BILLING_CUSTOMERS,
		postgres.TABLE_USERS, postgres.TABLE_ORGANIZATIONS} {
		queries = append(queries, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgTokensRepositoryPGTestSuite) exec(query string, args ...any) {
	s.T().Helper()
	execSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *OrgTokensRepositoryPGTestSuite) scalar(query string, args ...any) string {
	s.T().Helper()
	return scalarSQL(s.T(), s.ctx, s.client, query, args...)
}

func (s *OrgTokensRepositoryPGTestSuite) orgIDOf(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM organizations WHERE name = $1`, name)
}

func (s *OrgTokensRepositoryPGTestSuite) customerID(providerID string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM billing_customers WHERE provider_id = $1`, providerID)
}

func (s *OrgTokensRepositoryPGTestSuite) userID(name string) string {
	s.T().Helper()
	return s.scalar(`SELECT id FROM users WHERE name = $1`, name)
}

func (s *OrgTokensRepositoryPGTestSuite) search(text string) []orgtokens.AggregatedToken {
	s.T().Helper()
	return s.searchOrg(s.orgID, text)
}

func (s *OrgTokensRepositoryPGTestSuite) searchOrg(orgID, text string) []orgtokens.AggregatedToken {
	s.T().Helper()
	res, err := s.repository.Search(s.ctx, orgID, &rql.Query{
		Limit:  50,
		Search: text,
		Sort:   []rql.Sort{{Name: "description", Order: "asc"}},
	})
	s.Require().NoError(err)
	return res.Tokens
}

func (s *OrgTokensRepositoryPGTestSuite) tokenWithDescription(description string) orgtokens.AggregatedToken {
	s.T().Helper()
	for _, token := range s.search("") {
		if token.Description == description {
			return token
		}
	}
	s.T().Fatalf("no token with description %q", description)
	return orgtokens.AggregatedToken{}
}

func (s *OrgTokensRepositoryPGTestSuite) descriptions(tokens []orgtokens.AggregatedToken) []string {
	s.T().Helper()
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, t.Description)
	}
	return out
}

func (s *OrgTokensRepositoryPGTestSuite) TestSkipsSoftDeletedTokensAndAccounts() {
	s.Equal([]string{"Empty user top up", "Live top up", "No user top up", "Top up by gone user"},
		s.descriptions(s.search("")))
}

func (s *OrgTokensRepositoryPGTestSuite) TestSoftDeletedOrgHasNoTokens() {
	s.Empty(s.descriptions(s.searchOrg(s.orgIDOf("ot-org-gone"), "")))
}

func (s *OrgTokensRepositoryPGTestSuite) TestTokensWithoutAUserAreStillListed() {
	s.Empty(s.tokenWithDescription("No user top up").UserID)
	s.Empty(s.tokenWithDescription("Empty user top up").UserID)
}

func (s *OrgTokensRepositoryPGTestSuite) TestSoftDeletedUserKeepsTheTokenAndHidesTheName() {
	live := s.tokenWithDescription("Live top up")
	s.Equal(s.userID("ot-uma"), live.UserID)
	s.Equal("Uma", live.UserTitle)

	gone := s.tokenWithDescription("Top up by gone user")
	s.Equal(s.userID("ot-gary"), gone.UserID)
	s.Empty(gone.UserTitle, "the user is soft-deleted")
}

func (s *OrgTokensRepositoryPGTestSuite) TestSearchMatchesLiveRowsOnly() {
	s.Equal([]string{"Live top up"}, s.descriptions(s.search("Live top")))
	s.Empty(s.descriptions(s.search("Deleted top up")), "the transaction is soft-deleted")
	s.Empty(s.descriptions(s.search("Gone account")), "the billing account is soft-deleted")
	s.Empty(s.descriptions(s.search("Gary")), "the user is soft-deleted")
	s.Empty(s.descriptions(s.search("Gone org")), "the organization is soft-deleted")
}

func TestOrgTokensRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgTokensRepositoryPGTestSuite))
}
