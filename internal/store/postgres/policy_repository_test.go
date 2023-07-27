package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
)

type PolicyRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.PolicyRepository
	policyIDs  []string
	userID     string
	orgID      string
	roleID     string
}

func (s *PolicyRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewPolicyRepository(s.client)

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapPermissions(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	orgs, err := bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.orgID = orgs[0].ID

	roles, err := bootstrapRole(s.client, orgs[0].ID)
	if err != nil {
		s.T().Fatal(err)
	}
	s.roleID = roles[0]

	users, err := bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.userID = users[0].ID
}

func (s *PolicyRepositoryTestSuite) SetupTest() {
	var err error
	s.policyIDs, err = bootstrapPolicy(s.client, s.orgID, s.roleID, s.userID)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *PolicyRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *PolicyRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *PolicyRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_POLICIES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *PolicyRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description    string
		SelectedID     string
		ExpectedPolicy policy.Policy
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should get a policy",
			SelectedID:  s.policyIDs[0],
			ExpectedPolicy: policy.Policy{
				RoleID:        s.roleID,
				ResourceType:  "ns1",
				PrincipalID:   s.userID,
				PrincipalType: schema.UserPrincipal,
			},
		},
		{
			Description: "should return error invalid id if empty",
			ErrString:   policy.ErrInvalidID.Error(),
		},
		{
			Description: "should return error no exist if can't found policy",
			SelectedID:  uuid.NewString(),
			ErrString:   policy.ErrNotExist.Error(),
		},
		{
			Description: "should return error no exist if id is not uuid",
			SelectedID:  "some-id",
			ErrString:   policy.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Get(s.ctx, tc.SelectedID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedPolicy, cmpopts.IgnoreFields(policy.Policy{},
				"ID", "ResourceID")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPolicy)
			}
		})
	}
}

func (s *PolicyRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description    string
		PolicyToCreate policy.Policy
		Err            error
	}

	var testCases = []testCase{
		{
			Description: "should create a policy",
			PolicyToCreate: policy.Policy{
				RoleID:        s.roleID,
				ResourceID:    uuid.NewString(),
				ResourceType:  "ns1",
				PrincipalID:   s.userID,
				PrincipalType: schema.UserPrincipal,
			},
		},
		{
			Description: "should return error if role id does not exist",
			PolicyToCreate: policy.Policy{
				RoleID:       "role2-random",
				ResourceType: "ns1",
			},
			Err: policy.ErrInvalidDetail,
		},
		{
			Description: "should return error if namespace id does not exist",
			PolicyToCreate: policy.Policy{
				RoleID:       s.roleID,
				ResourceType: "ns1-random",
			},
			Err: policy.ErrInvalidDetail,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Upsert(s.ctx, tc.PolicyToCreate)
			if tc.Err != nil {
				if errors.Is(tc.Err, err) {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err.Error())
				}
			} else {
				s.Assert().NoError(err)
				if len(got) != len(uuid.NewString()) {
					s.T().Fatalf("got result %s, expected was a uuid", got)
				}
			}
		})
	}
}

func (s *PolicyRepositoryTestSuite) TestList() {
	type testCase struct {
		Description     string
		ExpectedPolicys []policy.Policy
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description: "should get all policies",
			ExpectedPolicys: []policy.Policy{
				{
					RoleID:       s.roleID,
					PrincipalID:  s.userID,
					ResourceID:   s.orgID,
					ResourceType: "ns1",
				},
				{
					RoleID:       s.roleID,
					PrincipalID:  s.userID,
					ResourceID:   s.orgID,
					ResourceType: "ns2",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, policy.Filter{})
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if len(got) != len(tc.ExpectedPolicys) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPolicys)
			}
			if !cmp.Equal(got, tc.ExpectedPolicys, cmpopts.IgnoreFields(policy.Policy{},
				"ID", "CreatedAt", "UpdatedAt")) {
			}
		})
	}
}

func (s *PolicyRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description      string
		PolicyToUpdate   policy.Policy
		ExpectedPolicyID string
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should update an policy",
			PolicyToUpdate: policy.Policy{
				ID:           s.policyIDs[0],
				RoleID:       s.roleID,
				ResourceType: "ns1",
			},
			ExpectedPolicyID: s.policyIDs[0],
		},
		{
			Description:      "should return error if policy id is empty",
			ErrString:        policy.ErrInvalidID.Error(),
			ExpectedPolicyID: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.PolicyToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedPolicyID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPolicyID)
			}
		})
	}
}

func TestPolicyRepository(t *testing.T) {
	suite.Run(t, new(PolicyRepositoryTestSuite))
}
