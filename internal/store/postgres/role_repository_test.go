package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

type RoleRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.RoleRepository
}

func (s *RoleRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewRoleRepository(s.client)

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *RoleRepositoryTestSuite) SetupTest() {
	var err error
	_, err = bootstrapRole(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *RoleRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RoleRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RoleRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ROLES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *RoleRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description  string
		SelectedID   string
		ExpectedRole role.Role
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should get a role",
			SelectedID:  "role1",
			ExpectedRole: role.Role{
				ID:   "role1",
				Name: "role member",
				Types: []string{
					"member",
					"user",
				},
				NamespaceID: "ns1",
				Namespace: namespace.Namespace{
					ID:   "ns1",
					Name: "ns1",
				},
			},
		},
		{
			Description: "should return error if id is empty",
			ErrString:   role.ErrInvalidID.Error(),
		},
		{
			Description: "should return error no exist if can't found role",
			SelectedID:  "10000",
			ErrString:   role.ErrNotExist.Error(),
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
			if !cmp.Equal(got, tc.ExpectedRole, cmpopts.IgnoreFields(role.Role{}, "Metadata", "Namespace.CreatedAt", "Namespace.UpdatedAt", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRole)
			}
		})
	}
}

func (s *RoleRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description  string
		RoleToCreate role.Role
		ExpectedID   string
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should create a role",
			RoleToCreate: role.Role{
				ID:   "role3",
				Name: "role other",
				Types: []string{
					"some-type1",
					"some-type2",
				},
				NamespaceID: "ns1",
			},
			ExpectedID: "role3",
		},
		{
			Description: "should return error if role name conflicted",
			RoleToCreate: role.Role{
				ID:   "role-conflict",
				Name: "role other",
				Types: []string{
					"some-type1",
					"some-type2",
				},
				NamespaceID: "ns1",
			},
			ErrString: role.ErrConflict.Error(),
		},
		{
			Description: "should return error if namespace id does not exist",
			RoleToCreate: role.Role{
				ID:   "role-new",
				Name: "role other new",
				Types: []string{
					"some-type1",
					"some-type2",
				},
				NamespaceID: "random-ns",
			},
			ErrString: role.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if role id is empty",
			ErrString:   role.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.RoleToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ExpectedID != "" && (got != tc.ExpectedID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedID)
			}
		})
	}
}

func (s *RoleRepositoryTestSuite) TestList() {
	type testCase struct {
		Description   string
		ExpectedRoles []role.Role
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should get all roles",
			ExpectedRoles: []role.Role{
				{
					ID:   "role1",
					Name: "role member",
					Types: []string{
						"member",
						"user",
					},
					NamespaceID: "ns1",
					Namespace: namespace.Namespace{
						ID:   "ns1",
						Name: "ns1",
					},
				},
				{
					ID:   "role2",
					Name: "role admin",
					Types: []string{
						"admin",
						"user",
					},
					NamespaceID: "ns2",
					Namespace: namespace.Namespace{
						ID:   "ns2",
						Name: "ns2",
					},
					Metadata: map[string]any{
						"key-string":  "value-string",
						"key-integer": 123,
						"key-json": map[string]any{
							"k1": "v1",
							"k2": "v2",
						},
					},
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
			if !cmp.Equal(got, tc.ExpectedRoles, cmpopts.IgnoreFields(role.Role{}, "Metadata", "Namespace.CreatedAt", "Namespace.UpdatedAt", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRoles)
			}
		})
	}
}

func (s *RoleRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description  string
		RoleToUpdate role.Role
		ExpectedID   string
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should update a role",
			RoleToUpdate: role.Role{
				ID:   "role1",
				Name: "role member new updated",
				Types: []string{
					"member",
					"user",
					"role-member",
				},
				NamespaceID: "ns1",
			},
			ExpectedID: "role1",
		},
		{
			Description: "should return error if role name conflicted",
			RoleToUpdate: role.Role{
				ID:   "role2",
				Name: "role member new updated",
				Types: []string{
					"member",
					"user",
					"role-member",
				},
				NamespaceID: "ns1",
			},
			ErrString: role.ErrConflict.Error(),
		},
		{
			Description: "should return error if namespace id not exist",
			RoleToUpdate: role.Role{
				ID:   "role1",
				Name: "role member new updated",
				Types: []string{
					"member",
					"user",
					"role-member",
				},
				NamespaceID: "random-ns",
			},
			ErrString: role.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if role not found",
			RoleToUpdate: role.Role{
				ID:   "123131",
				Name: "not-exist",
			},
			ErrString: role.ErrNotExist.Error(),
		},
		{
			Description: "should return error if role id is empty",
			ErrString:   role.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.RoleToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ExpectedID != "" && (got != tc.ExpectedID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedID)
			}
		})
	}
}

func TestRoleRepository(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}
