package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type PreferenceRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.PreferenceRepository
}

func (s *PreferenceRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewPreferenceRepository(s.client)

	// Bootstrap namespaces for foreign key constraints
	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	// Add app/platform namespace for global scope preferences
	_, err = s.client.DB.ExecContext(s.ctx, "INSERT INTO namespaces (name) VALUES ('app/platform') ON CONFLICT DO NOTHING")
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *PreferenceRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *PreferenceRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *PreferenceRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_PREFERENCES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *PreferenceRepositoryTestSuite) TestSet() {
	type testCase struct {
		Description        string
		PreferenceToCreate preference.Preference
		ExpectedPreference preference.Preference
		ErrString          string
	}

	var testCases = []testCase{
		{
			Description: "should create a preference with global scope when scope not provided",
			PreferenceToCreate: preference.Preference{
				Name:         "test_pref",
				Value:        "test_value",
				ResourceType: schema.UserPrincipal,
				ResourceID:   "user-123",
			},
			ExpectedPreference: preference.Preference{
				Name:         "test_pref",
				Value:        "test_value",
				ResourceType: schema.UserPrincipal,
				ResourceID:   "user-123",
				ScopeType:    "", // transformed back to empty from global
				ScopeID:      "", // transformed back to empty from global
			},
		},
		{
			Description: "should create a preference with specific scope",
			PreferenceToCreate: preference.Preference{
				Name:         "scoped_pref",
				Value:        "scoped_value",
				ResourceType: schema.UserPrincipal,
				ResourceID:   "user-456",
				ScopeType:    schema.OrganizationNamespace,
				ScopeID:      "org-123",
			},
			ExpectedPreference: preference.Preference{
				Name:         "scoped_pref",
				Value:        "scoped_value",
				ResourceType: schema.UserPrincipal,
				ResourceID:   "user-456",
				ScopeType:    schema.OrganizationNamespace,
				ScopeID:      "org-123",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Set(s.ctx, tc.PreferenceToCreate)
			if tc.ErrString != "" {
				if err == nil || err.Error() != tc.ErrString {
					s.T().Fatalf("got error %v, expected was %s", err, tc.ErrString)
				}
				return
			}
			s.Assert().NoError(err)
			if !cmp.Equal(got, tc.ExpectedPreference, cmpopts.IgnoreFields(preference.Preference{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPreference)
			}
		})
	}
}

func (s *PreferenceRepositoryTestSuite) TestListWithScopeFilter() {
	// Setup: Create preferences with different scopes
	userID := "user-test-123"

	// Global scoped preference (no scope specified)
	globalPref := preference.Preference{
		Name:         "global_pref",
		Value:        "global_value",
		ResourceType: schema.UserPrincipal,
		ResourceID:   userID,
	}
	_, err := s.repository.Set(s.ctx, globalPref)
	s.Require().NoError(err)

	// Org1 scoped preference
	org1Pref := preference.Preference{
		Name:         "org_pref",
		Value:        "org1_value",
		ResourceType: schema.UserPrincipal,
		ResourceID:   userID,
		ScopeType:    schema.OrganizationNamespace,
		ScopeID:      "org-1",
	}
	_, err = s.repository.Set(s.ctx, org1Pref)
	s.Require().NoError(err)

	// Org2 scoped preference
	org2Pref := preference.Preference{
		Name:         "org_pref",
		Value:        "org2_value",
		ResourceType: schema.UserPrincipal,
		ResourceID:   userID,
		ScopeType:    schema.OrganizationNamespace,
		ScopeID:      "org-2",
	}
	_, err = s.repository.Set(s.ctx, org2Pref)
	s.Require().NoError(err)

	type testCase struct {
		Description     string
		Filter          preference.Filter
		ExpectedCount   int
		ExpectedValue   string
		ExpectedName    string
	}

	var testCases = []testCase{
		{
			Description: "should return only global scoped preferences when no scope filter provided",
			Filter: preference.Filter{
				UserID: userID,
				// No ScopeType/ScopeID = defaults to global
			},
			ExpectedCount: 1,
			ExpectedValue: "global_value",
			ExpectedName:  "global_pref",
		},
		{
			Description: "should return org1 scoped preferences when filtered by org1",
			Filter: preference.Filter{
				UserID:    userID,
				ScopeType: schema.OrganizationNamespace,
				ScopeID:   "org-1",
			},
			ExpectedCount: 1,
			ExpectedValue: "org1_value",
			ExpectedName:  "org_pref",
		},
		{
			Description: "should return org2 scoped preferences when filtered by org2",
			Filter: preference.Filter{
				UserID:    userID,
				ScopeType: schema.OrganizationNamespace,
				ScopeID:   "org-2",
			},
			ExpectedCount: 1,
			ExpectedValue: "org2_value",
			ExpectedName:  "org_pref",
		},
		{
			Description: "should return empty when filtered by non-existent scope",
			Filter: preference.Filter{
				UserID:    userID,
				ScopeType: schema.OrganizationNamespace,
				ScopeID:   "org-non-existent",
			},
			ExpectedCount: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, tc.Filter)
			s.Assert().NoError(err)
			s.Assert().Len(got, tc.ExpectedCount)
			if tc.ExpectedCount > 0 {
				s.Assert().Equal(tc.ExpectedName, got[0].Name)
				s.Assert().Equal(tc.ExpectedValue, got[0].Value)
			}
		})
	}
}

func TestPreferenceRepository(t *testing.T) {
	suite.Run(t, new(PreferenceRepositoryTestSuite))
}
