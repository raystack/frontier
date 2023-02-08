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
	"github.com/odpf/shield/pkg/metadata"
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
	roleIDs    []string
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

	s.roleIDs, err = bootstrapRole(s.client)
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
			SelectedID:  "ns1:role1",
			ExpectedRole: role.Role{
				ID:   "ns1:role1",
				Name: "role member",
				Types: []string{
					"member",
					"user",
				},
				NamespaceID: "ns1",
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
			if !cmp.Equal(got, tc.ExpectedRole, cmpopts.IgnoreFields(role.Role{}, "Metadata", "CreatedAt", "UpdatedAt")) {
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
					ID:          "ns1:role1",
					NamespaceID: "ns1",
					Name:        "role member",
				},
				{
					ID:          "ns1:role2",
					NamespaceID: "ns1",
					Name:        "role member",
				},
				{
					ID:          "ns2:role2",
					NamespaceID: "ns2",
					Name:        "role admin",
				},
				{
					ID:          "ns1:ns2",
					NamespaceID: "ns1",
					Name:        "namespace",
				},
				{
					ID:          "shield/organization:owner",
					NamespaceID: "shield/organization",
					Name:        "owner",
				},
				{
					ID:          "shield/organization:editor",
					NamespaceID: "shield/organization",
					Name:        "editor",
				},
				{
					ID:          "shield/organization:viewer",
					NamespaceID: "shield/organization",
					Name:        "viewer",
				},
				{
					ID:          "shield/project:owner",
					NamespaceID: "shield/project",
					Name:        "owner",
				},
				{
					ID:          "shield/project:editor",
					NamespaceID: "shield/project",
					Name:        "editor",
				},
				{
					ID:          "shield/project:viewer",
					NamespaceID: "shield/project",
					Name:        "viewer",
				},
				{
					ID:          "shield/project:organization",
					NamespaceID: "shield/project",
					Name:        "organization",
				},
				{
					ID:          "shield/group:member",
					NamespaceID: "shield/group",
					Name:        "member",
				},
				{
					ID:          "shield/group:manager",
					NamespaceID: "shield/group",
					Name:        "manager",
				},
				{
					ID:          "shield/group:organization",
					NamespaceID: "shield/group",
					Name:        "organization",
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
			if !cmp.Equal(got, tc.ExpectedRoles, cmpopts.IgnoreFields(role.Role{},
				"ID", "Types", "CreatedAt", "UpdatedAt", "Metadata")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRoles)
			}
		})
	}
}

func (s *RoleRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description    string
		RoleToUpdate   role.Role
		ExpectedRoleID string
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should update a role",
			RoleToUpdate: role.Role{
				ID:          s.roleIDs[0],
				Name:        "role members",
				NamespaceID: "ns1",
				Metadata:    metadata.Metadata{},
				Types:       []string{"member", "user"},
			},
			ExpectedRoleID: s.roleIDs[0],
		},
		{
			Description: "should return error if namespace id does not exist",
			RoleToUpdate: role.Role{
				ID:          s.roleIDs[0],
				Name:        "role member",
				NamespaceID: "ns-random",
				Metadata:    metadata.Metadata{},
				Types:       []string{"member", "user"},
			},
			ExpectedRoleID: "",
			ErrString:      namespace.ErrNotExist.Error(),
		},
		{
			Description: "should return error if role not found",
			RoleToUpdate: role.Role{
				ID:          "ns:random",
				Name:        "role member",
				NamespaceID: "ns1",
				Metadata:    metadata.Metadata{},
				Types:       []string{"member", "user"},
			},
			ExpectedRoleID: "",
			ErrString:      role.ErrNotExist.Error(),
		},
		{
			Description: "should return error if policy id is empty",
			RoleToUpdate: role.Role{
				ID:          "",
				Name:        "role member",
				NamespaceID: "ns1",
				Metadata:    metadata.Metadata{},
				Types:       []string{"member", "user"},
			},
			ExpectedRoleID: "",
			ErrString:      role.ErrInvalidID.Error(),
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
			if !cmp.Equal(got, tc.ExpectedRoleID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRoleID)
			}
		})
	}
}

func TestRoleRepository(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}
