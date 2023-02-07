package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
)

type PolicyRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.PolicyRepository
	policyIDs  []string
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

	_, err = bootstrapAction(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapRole(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *PolicyRepositoryTestSuite) SetupTest() {
	var err error
	s.policyIDs, err = bootstrapPolicy(s.client)
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
				RoleID:      "ns1:role1",
				NamespaceID: "ns1",
				ActionID:    "policy1",
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
			if !cmp.Equal(got, tc.ExpectedPolicy, cmpopts.IgnoreFields(policy.Policy{},
				"ID", "CreatedAt", "UpdatedAt")) {
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
				RoleID:      "ns1:role2",
				ActionID:    "policy4",
				NamespaceID: "ns1",
			},
		},
		{
			Description: "should return error if role id does not exist",
			PolicyToCreate: policy.Policy{
				RoleID:      "role2-random",
				ActionID:    "policy4",
				NamespaceID: "ns1",
			},
			Err: policy.ErrInvalidDetail,
		},
		{
			Description: "should return error if policy id does not exist",
			PolicyToCreate: policy.Policy{
				RoleID:      "role2",
				ActionID:    "policy4-random",
				NamespaceID: "ns1",
			},
			Err: policy.ErrInvalidDetail,
		},
		{
			Description: "should return error if namespace id does not exist",
			PolicyToCreate: policy.Policy{
				RoleID:      "role2",
				ActionID:    "policy4",
				NamespaceID: "ns1-random",
			},
			Err: policy.ErrInvalidDetail,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.PolicyToCreate)
			if tc.Err != nil {
				if errors.Is(tc.Err, err) {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err.Error())
				}
			} else {
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
			Description: "should get all policys",
			ExpectedPolicys: []policy.Policy{
				{
					RoleID:      "ns1:role1",
					NamespaceID: "ns1",
					ActionID:    "action1",
				},
				{
					RoleID:      "ns2:role2",
					NamespaceID: "ns2",
					ActionID:    "action2",
				},
				{
					RoleID:      "ns1:role2",
					NamespaceID: "ns1",
					ActionID:    "action3",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			//TODO figure out how to compare metadata map[string]any
			if !cmp.Equal(got, tc.ExpectedPolicys, cmpopts.IgnoreFields(policy.Policy{},
				"ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPolicys)
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
				ID:          s.policyIDs[0],
				RoleID:      "ns1:role1",
				ActionID:    "action4",
				NamespaceID: "ns1",
			},
			ExpectedPolicyID: s.policyIDs[0],
		},
		{
			Description: "should return error if namespace id does not exist",
			PolicyToUpdate: policy.Policy{
				ID:          s.policyIDs[1],
				RoleID:      "ns2:role2",
				ActionID:    "action2",
				NamespaceID: "random-ns2",
			},
			ErrString:        namespace.ErrNotExist.Error(),
			ExpectedPolicyID: "",
		},
		{
			Description: "should return error if policy not found",
			PolicyToUpdate: policy.Policy{
				ID:          uuid.NewString(),
				RoleID:      "ns1:role2",
				NamespaceID: "ns1",
				ActionID:    "action3",
			},
			ErrString:        policy.ErrNotExist.Error(),
			ExpectedPolicyID: "",
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
			if !cmp.Equal(got, tc.ExpectedPolicyID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPolicyID)
			}
		})
	}
}

func TestPolicyRepository(t *testing.T) {
	suite.Run(t, new(PolicyRepositoryTestSuite))
}
