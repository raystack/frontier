package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/odpf/shield/core/permission"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

type PermissionRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.PermissionRepository
	permsInDB  []permission.Permission
}

func (s *PermissionRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewPermissionRepository(s.client)

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *PermissionRepositoryTestSuite) SetupTest() {
	var err error
	s.permsInDB, err = bootstrapPermissions(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *PermissionRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *PermissionRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *PermissionRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_PERMISSIONS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *PermissionRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description        string
		SelectedID         string
		ExpectedPermission permission.Permission
		ErrString          string
	}

	var testCases = []testCase{
		{
			Description: "should get a permission",
			SelectedID:  s.permsInDB[0].ID,
			ExpectedPermission: permission.Permission{
				ID:          s.permsInDB[0].ID,
				Name:        s.permsInDB[0].Name,
				NamespaceID: s.permsInDB[0].NamespaceID,
				Slug:        s.permsInDB[0].Slug,
			},
		},
		{
			Description: "should return error no exist if can't found permission",
			SelectedID:  uuid.New().String(),
			ErrString:   permission.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id is empty",
			ErrString:   permission.ErrInvalidID.Error(),
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
			if !cmp.Equal(got, tc.ExpectedPermission, cmpopts.IgnoreFields(permission.Permission{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPermission)
			}
		})
	}
}

func (s *PermissionRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description        string
		PermissionToCreate permission.Permission
		ExpectedPermission permission.Permission
		ErrString          string
	}
	id1 := uuid.NewString()
	var testCases = []testCase{
		{
			Description: "should create an permission",
			PermissionToCreate: permission.Permission{
				ID:          id1,
				Name:        "permission-123",
				NamespaceID: "ns2",
				Slug:        "ns_2_perm",
			},
			ExpectedPermission: permission.Permission{
				ID:          id1,
				Name:        "permission-123",
				NamespaceID: "ns2",
				Slug:        "ns_2_perm",
			},
		},
		{
			Description: "should return error if namespace id not exist",
			PermissionToCreate: permission.Permission{
				ID:          uuid.NewString(),
				Name:        "permission-123",
				NamespaceID: "random-ns",
				Slug:        "ns_3_perm",
			},
			ErrString: namespace.ErrNotExist.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Upsert(s.ctx, tc.PermissionToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedPermission, cmpopts.IgnoreFields(permission.Permission{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPermission)
			}
		})
	}
}

func (s *PermissionRepositoryTestSuite) TestList() {
	type testCase struct {
		Description         string
		ExpectedPermissions []permission.Permission
		ErrString           string
	}

	var testCases = []testCase{
		{
			Description:         "should get all permissions",
			ExpectedPermissions: s.permsInDB,
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
			if !cmp.Equal(got, tc.ExpectedPermissions, cmpopts.IgnoreFields(permission.Permission{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPermissions)
			}
		})
	}
}

func (s *PermissionRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description        string
		PermissionToUpdate permission.Permission
		ExpectedPermission permission.Permission
		ErrString          string
	}

	var testCases = []testCase{
		{
			Description: "should update a permission",
			PermissionToUpdate: permission.Permission{
				ID:          s.permsInDB[0].ID,
				Name:        "permission-get-updated",
				NamespaceID: "ns2",
			},
			ExpectedPermission: permission.Permission{
				ID:          s.permsInDB[0].ID,
				Name:        "permission-get-updated",
				NamespaceID: "ns2",
				Slug:        s.permsInDB[0].Slug,
			},
		},
		{
			Description: "should return error if namespace id does not exist",
			PermissionToUpdate: permission.Permission{
				ID:          s.permsInDB[0].ID,
				Name:        "permission-get-updated",
				NamespaceID: "random-ns2",
			},
			ErrString: namespace.ErrNotExist.Error(),
		},
		{
			Description: "should return error if permission not found",
			PermissionToUpdate: permission.Permission{
				ID:   uuid.NewString(),
				Name: "not-exist",
			},
			ErrString: "permission doesn't exist",
		},
		{
			Description: "should return error if permission id is empty",
			ErrString:   "permission id is invalid",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.PermissionToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedPermission, cmpopts.IgnoreFields(permission.Permission{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedPermission)
			}
		})
	}
}

func TestPermissionRepository(t *testing.T) {
	suite.Run(t, new(PermissionRepositoryTestSuite))
}
