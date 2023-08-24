package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type RoleRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.RoleRepository
	roles      []role.Role
	orgID      string
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

	orgs, err := bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.orgID = orgs[0].ID
}

func (s *RoleRepositoryTestSuite) SetupTest() {
	var err error
	s.roles, err = bootstrapRole(s.client, s.orgID)
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
			SelectedID:  s.roles[3].ID,
			ExpectedRole: role.Role{
				ID:    s.roles[3].ID,
				Name:  "editor",
				Title: "Test Title",
				Permissions: []string{
					"user",
					"group",
				},
				OrgID: s.orgID,
			},
		},
		{
			Description: "should return error if id is empty",
			ErrString:   role.ErrInvalidID.Error(),
		},
		{
			Description: "should return error no exist if can't found role",
			SelectedID:  uuid.NewString(),
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
			if tc.ErrString == "" {
				s.Assert().NoError(err)
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
		ErrString    error
	}
	roleID1 := uuid.New().String()
	roleID2 := uuid.New().String()
	var testCases = []testCase{
		{
			Description: "should create a role",
			RoleToCreate: role.Role{
				ID:    roleID1,
				Name:  "role other",
				Title: "Test Title",
				Permissions: []string{
					"some-type1",
					"some-type2",
				},
				OrgID:    s.orgID,
				Metadata: metadata.Metadata{},
			},
			ExpectedID: roleID1,
		},
		{
			Description: "should return error if org id does not exist",
			RoleToCreate: role.Role{
				ID:    roleID2,
				Name:  "role other new",
				Title: "Test Title",
				Permissions: []string{
					"some-type1",
					"some-type2",
				},
				OrgID:    "random-ns",
				Metadata: metadata.Metadata{},
			},
			ErrString: postgres.ErrInvalidTextRepresentation,
		},
		{
			Description: "should return error if org id is empty",
			ErrString:   role.ErrInvalidDetail,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Upsert(s.ctx, tc.RoleToCreate)
			if tc.ErrString != nil {
				if !errors.Is(err, tc.ErrString) {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ExpectedID != "" && (got.ID != tc.ExpectedID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedID)
			}
		})
	}
}

func (s *RoleRepositoryTestSuite) TestList() {
	type testCase struct {
		Description      string
		ExpectedRolesLen int
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description:      "should get all roles",
			ExpectedRolesLen: 8,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, role.Filter{})
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if len(got) != tc.ExpectedRolesLen {
				s.T().Fatalf("got result %+v, expected was %+v", len(got), tc.ExpectedRolesLen)
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
				ID:          s.roles[0].ID,
				Name:        "role members",
				Title:       "Test Title",
				OrgID:       s.orgID,
				Metadata:    metadata.Metadata{},
				Permissions: []string{"member", "user"},
			},
			ExpectedRoleID: s.roles[0].ID,
		},
		{
			Description: "should return error if role not found",
			RoleToUpdate: role.Role{
				ID:          uuid.NewString(),
				Name:        "role member",
				Title:       "Test Title",
				OrgID:       "ns1",
				Metadata:    metadata.Metadata{},
				Permissions: []string{"member", "user"},
			},
			ExpectedRoleID: "",
			ErrString:      role.ErrNotExist.Error(),
		},
		{
			Description: "should return error if policy id is empty",
			RoleToUpdate: role.Role{
				ID:          "",
				Name:        "role member",
				Title:       "Test Title",
				OrgID:       "ns1",
				Metadata:    metadata.Metadata{},
				Permissions: []string{"member", "user"},
			},
			ExpectedRoleID: "",
			ErrString:      role.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if role name is empty",
			RoleToUpdate: role.Role{
				ID:   s.roles[0].ID,
				Name: "",
			},
			ExpectedRoleID: "",
			ErrString:      role.ErrInvalidDetail.Error(),
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
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got.ID, tc.ExpectedRoleID) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRoleID)
			}
		})
	}
}

func (s *RoleRepositoryTestSuite) TestDelete() {
	type testCase struct {
		Description  string
		RoleToDelete string
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description:  "should return error if role not found",
			RoleToDelete: uuid.NewString(),
			ErrString:    role.ErrNotExist.Error(),
		},
		{
			Description:  "should delete a role and return no error on success",
			RoleToDelete: s.roles[0].ID,
			ErrString:    "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			err := s.repository.Delete(s.ctx, tc.RoleToDelete)
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

func (s *RoleRepositoryTestSuite) TestGetByName() {
	type testCase struct {
		Description  string
		RoleName     string
		ExpectedRole role.Role
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description:  "should return error if role name is empty",
			RoleName:     "",
			ExpectedRole: role.Role{},
			ErrString:    role.ErrInvalidDetail.Error(),
		},
		{
			Description:  "should return error if role not found",
			RoleName:     "role not found",
			ExpectedRole: role.Role{},
			ErrString:    role.ErrNotExist.Error(),
		},
		{
			Description: "should get a role by name",
			RoleName:    "editor",
			ExpectedRole: role.Role{
				ID:    s.roles[3].ID,
				Name:  "editor",
				Title: "Test Title",
				Permissions: []string{
					"user",
					"group",
				},
				OrgID: s.orgID,
			},
			ErrString: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByName(s.ctx, tc.ExpectedRole.OrgID, tc.RoleName)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
				s.Assert().Equal(tc.ExpectedRole, got)
			}
		})
	}
}

func TestRoleRepository(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}
