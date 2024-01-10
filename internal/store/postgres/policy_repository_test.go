package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/raystack/frontier/core/role"

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
	policies   []policy.Policy
	userID     string
	orgID      string
	roles      []role.Role
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
	s.roles = roles

	users, err := bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.userID = users[0].ID
}

func (s *PolicyRepositoryTestSuite) SetupTest() {
	var err error
	s.policies, err = bootstrapPolicy(s.client, s.orgID, s.roles[0], s.userID)
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
			SelectedID:  s.policies[0].ID,
			ExpectedPolicy: policy.Policy{
				RoleID:        s.roles[0].ID,
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
				RoleID:        s.roles[0].ID,
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
				RoleID:       s.roles[0].ID,
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
				if got.ID == "" {
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
		flt             policy.Filter
	}

	var testCases = []testCase{
		{
			Description: "should get all policies",
			ExpectedPolicys: []policy.Policy{
				{
					RoleID:       s.roles[0].ID,
					PrincipalID:  s.userID,
					ResourceID:   s.orgID,
					ResourceType: "ns1",
				},
				{
					RoleID:       s.roles[0].ID,
					PrincipalID:  s.userID,
					ResourceID:   s.orgID,
					ResourceType: "ns2",
				},
			},
		},
		{
			Description:     "should get all policies with filters for policy",
			ExpectedPolicys: nil,
			flt: policy.Filter{
				RoleID:        s.roles[0].ID,
				OrgID:         s.orgID,
				PrincipalID:   s.userID,
				PrincipalType: schema.UserPrincipal,
				ProjectID:     uuid.NewString(),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, tc.flt)
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
				ID:           s.policies[0].ID,
				RoleID:       s.roles[0].ID,
				ResourceType: "ns1",
			},
			ExpectedPolicyID: s.policies[0].ID,
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

func (s *PolicyRepositoryTestSuite) TestDelete() {
	type testCase struct {
		Description string
		PolicyID    string
		ErrString   string
	}

	var testCases = []testCase{
		{
			Description: "should delete a policy",
			PolicyID:    s.policies[0].ID,
			ErrString:   "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			err := s.repository.Delete(s.ctx, tc.PolicyID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
		})
	}
}

func TestPolicyRepository(t *testing.T) {
	suite.Run(t, new(PolicyRepositoryTestSuite))
}

func (s *PolicyRepositoryTestSuite) TestGroupMemberCount() {
	type testCase struct {
		Description    string
		PolicyToCreate []policy.Policy
		GroupIDs       []string
		Want           []policy.MemberCount
		Err            error
	}
	g1 := uuid.NewString()
	var testCases = []testCase{
		{
			Description: "count group users of different roles as same",
			PolicyToCreate: []policy.Policy{
				{
					RoleID:        s.roles[0].ID,
					ResourceID:    g1,
					ResourceType:  schema.GroupNamespace,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
				},
				{
					RoleID:        s.roles[1].ID,
					ResourceID:    g1,
					ResourceType:  schema.GroupNamespace,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
				},
			},
			GroupIDs: []string{
				g1,
			},
			Want: []policy.MemberCount{
				{
					ID:    g1,
					Count: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			for _, p := range tc.PolicyToCreate {
				_, err := s.repository.Upsert(s.ctx, p)
				s.Assert().NoError(err)
			}

			got, err := s.repository.GroupMemberCount(s.ctx, tc.GroupIDs)
			if tc.Err != nil {
				s.Assert().ErrorAs(tc.Err, err, "got error %s, expected was %s", err.Error(), tc.Err.Error())
			} else {
				s.Assert().EqualValues(tc.Want, got, "got result %v, expected was %v", got, tc.Want)
			}
		})
	}
}

func (s *PolicyRepositoryTestSuite) TestProjectMemberCount() {
	type testCase struct {
		Description    string
		PolicyToCreate []policy.Policy
		ProjectIDs     []string
		Want           []policy.MemberCount
		Err            error
	}
	p1 := uuid.NewString()
	var testCases = []testCase{
		{
			Description: "count project users of different roles as same",
			PolicyToCreate: []policy.Policy{
				{
					RoleID:        s.roles[0].ID,
					ResourceID:    p1,
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
				},
				{
					RoleID:        s.roles[1].ID,
					ResourceID:    p1,
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
				},
				{
					RoleID:        s.roles[1].ID,
					ResourceID:    p1,
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   uuid.NewString(),
					PrincipalType: schema.GroupNamespace,
				},
			},
			ProjectIDs: []string{
				p1,
			},
			Want: []policy.MemberCount{
				{
					ID:    p1,
					Count: 2,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			for _, p := range tc.PolicyToCreate {
				_, err := s.repository.Upsert(s.ctx, p)
				s.Assert().NoError(err)
			}

			got, err := s.repository.ProjectMemberCount(s.ctx, tc.ProjectIDs)
			if tc.Err != nil {
				s.Assert().ErrorAs(tc.Err, err, "got error %s, expected was %s", err.Error(), tc.Err.Error())
			} else {
				s.Assert().EqualValues(tc.Want, got, "got result %v, expected was %v", got, tc.Want)
			}
		})
	}
}

func (s *PolicyRepositoryTestSuite) TestOrgMemberCount() {
	type testCase struct {
		Description    string
		PolicyToCreate []policy.Policy
		OrgID          string
		Want           policy.MemberCount
		Err            error
	}
	o1 := uuid.NewString()
	var testCases = []testCase{
		{
			Description: "count org users of different roles as same",
			PolicyToCreate: []policy.Policy{
				{
					RoleID:        s.roles[0].ID,
					ResourceID:    o1,
					ResourceType:  schema.OrganizationNamespace,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
				},
				{
					RoleID:        s.roles[1].ID,
					ResourceID:    o1,
					ResourceType:  schema.OrganizationNamespace,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
				},
				{
					RoleID:        s.roles[1].ID,
					ResourceID:    o1,
					ResourceType:  schema.OrganizationNamespace,
					PrincipalID:   uuid.NewString(),
					PrincipalType: schema.GroupNamespace,
				},
			},
			OrgID: o1,
			Want: policy.MemberCount{
				ID:    o1,
				Count: 1,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			for _, p := range tc.PolicyToCreate {
				_, err := s.repository.Upsert(s.ctx, p)
				s.Assert().NoError(err)
			}

			got, err := s.repository.OrgMemberCount(s.ctx, tc.OrgID)
			if tc.Err != nil {
				s.Assert().ErrorAs(tc.Err, err, "got error %s, expected was %s", err.Error(), tc.Err.Error())
			} else {
				s.Assert().EqualValues(tc.Want, got, "got result %v, expected was %v", got, tc.Want)
			}
		})
	}
}
