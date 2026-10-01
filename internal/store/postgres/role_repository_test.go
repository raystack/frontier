package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/uuid"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/suite"
)

type RoleRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.RoleRepository
	roles      []role.Role
	orgID      string
}

func (s *RoleRepositoryTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
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
	if err := closeTestClient(s.client); err != nil {
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
				OrgID:  s.orgID,
				Scopes: []string{},
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

	s.Run("should create a new row when the name belongs to a soft-deleted role", func() {
		deleted := s.roles[3]
		if _, err := s.client.ExecContext(s.ctx, "UPDATE roles SET deleted_at = now() WHERE id = $1", deleted.ID); err != nil {
			s.T().Fatal(err)
		}

		got, err := s.repository.Upsert(s.ctx, role.Role{
			Name:        deleted.Name,
			Title:       "Test Title",
			Permissions: []string{"user"},
			OrgID:       s.orgID,
			Metadata:    metadata.Metadata{},
		})
		s.Assert().NoError(err)
		if got.ID == deleted.ID {
			s.T().Fatalf("got the deleted row %s back, expected a new row", deleted.ID)
		}

		var rows int
		if err := s.client.QueryRowxContext(s.ctx, "SELECT count(*) FROM roles WHERE org_id = $1 AND name = $2", s.orgID, deleted.Name).Scan(&rows); err != nil {
			s.T().Fatal(err)
		}
		if rows != 2 {
			s.T().Fatalf("got %d rows named %s, expected the deleted row and the new one", rows, deleted.Name)
		}
	})

	s.Run("should return conflict when the id belongs to a soft-deleted role", func() {
		deleted := s.roles[4]
		if _, err := s.client.ExecContext(s.ctx, "UPDATE roles SET deleted_at = now() WHERE id = $1", deleted.ID); err != nil {
			s.T().Fatal(err)
		}

		_, err := s.repository.Upsert(s.ctx, role.Role{
			ID:       deleted.ID,
			Name:     "role with a reused id",
			OrgID:    s.orgID,
			Metadata: metadata.Metadata{},
		})
		s.Assert().ErrorIs(err, role.ErrConflict)
	})

	s.Run("should update the live role and move updated_at when the name is taken", func() {
		before, err := s.repository.Get(s.ctx, s.roles[2].ID)
		if err != nil {
			s.T().Fatal(err)
		}

		got, err := s.repository.Upsert(s.ctx, role.Role{
			Name:        before.Name,
			Title:       "changed",
			Permissions: before.Permissions,
			Scopes:      before.Scopes,
			OrgID:       s.orgID,
			Metadata:    metadata.Metadata{},
		})
		s.Assert().NoError(err)
		s.Assert().Equal(before.ID, got.ID)
		s.Assert().Equal("changed", got.Title)
		s.Assert().True(got.UpdatedAt.After(before.UpdatedAt))
	})
}

func (s *RoleRepositoryTestSuite) TestList() {
	type testCase struct {
		Description      string
		ExpectedRolesLen int
		ErrString        string
		Filter           role.Filter
	}

	var testCases = []testCase{
		{
			Description:      "should get all roles",
			ExpectedRolesLen: 8,
			Filter:           role.Filter{},
		},
		{
			Description:      "should get all roles org scope",
			ExpectedRolesLen: 2,
			Filter: role.Filter{
				Scopes: []string{
					schema.OrganizationNamespace,
				},
			},
		},
		{
			Description:      "should get all roles with either org or group scope",
			ExpectedRolesLen: 3,
			Filter: role.Filter{
				Scopes: []string{
					schema.OrganizationNamespace,
					schema.GroupPrincipal,
				},
			},
		},
		{
			Description:      "should get all roles group scope",
			ExpectedRolesLen: 1,
			Filter: role.Filter{
				Scopes: []string{
					schema.GroupNamespace,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, tc.Filter)
			if tc.ErrString != "" || err != nil {
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

	s.Run("should keep the row and mark it deleted", func() {
		target := s.roles[1]
		s.Require().NoError(s.repository.Delete(s.ctx, target.ID))

		var deleted bool
		err := s.client.QueryRowxContext(s.ctx, "SELECT deleted_at IS NOT NULL FROM roles WHERE id = $1", target.ID).Scan(&deleted)
		s.Require().NoError(err)
		s.Assert().True(deleted)
	})

	s.Run("should return not found when the role is already deleted", func() {
		target := s.roles[2]
		_, err := s.client.ExecContext(s.ctx, "UPDATE roles SET deleted_at = now() WHERE id = $1", target.ID)
		s.Require().NoError(err)

		s.Assert().ErrorIs(s.repository.Delete(s.ctx, target.ID), role.ErrNotExist)
	})

	s.Run("should return in use and keep the role when a live policy uses it", func() {
		target := s.roles[3]
		_, err := bootstrapPolicy(s.client, s.orgID, target, uuid.NewString())
		s.Require().NoError(err)

		s.Assert().ErrorIs(s.repository.Delete(s.ctx, target.ID), role.ErrRoleInUse)
		_, err = s.repository.Get(s.ctx, target.ID)
		s.Assert().NoError(err)
	})

	s.Run("should delete the role when only deleted policies use it", func() {
		target := s.roles[4]
		_, err := bootstrapPolicy(s.client, s.orgID, target, uuid.NewString())
		s.Require().NoError(err)
		_, err = s.client.ExecContext(s.ctx, "UPDATE policies SET deleted_at = now() WHERE role_id = $1", target.ID)
		s.Require().NoError(err)

		s.Assert().NoError(s.repository.Delete(s.ctx, target.ID))
	})

	s.Run("should wait for a policy insert that has not committed and return in use", func() {
		target, err := s.repository.Upsert(s.ctx, role.Role{
			Name:     "role with a running policy insert",
			OrgID:    s.orgID,
			Metadata: metadata.Metadata{},
		})
		s.Require().NoError(err)
		tx, err := s.client.BeginTxx(s.ctx, nil)
		s.Require().NoError(err)
		defer tx.Rollback() // nolint
		_, err = tx.ExecContext(s.ctx, "INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type) VALUES ($1, $2, 'ns1', $3, 'app/user')",
			target.ID, s.orgID, uuid.NewString())
		s.Require().NoError(err)

		done := make(chan error, 1)
		go func() { done <- s.repository.Delete(s.ctx, target.ID) }()

		s.Require().Eventually(func() bool {
			var waiting int
			err := s.client.QueryRowxContext(s.ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'").Scan(&waiting)
			return err == nil && waiting == 1
		}, time.Second, 10*time.Millisecond, "the delete did not wait for the policy insert")

		s.Require().NoError(tx.Commit())
		s.Assert().ErrorIs(<-done, role.ErrRoleInUse)
	})
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
				s.Assert().Equal(tc.ExpectedRole.ID, got.ID)
				s.Assert().Equal(tc.ExpectedRole.Name, got.Name)
				s.Assert().Equal(tc.ExpectedRole.Permissions, got.Permissions)
			}
		})
	}
}

func TestRoleRepository(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}

func (s *RoleRepositoryTestSuite) TestSkipsSoftDeletedRoles() {
	deleted := s.roles[0]
	flt := role.Filter{OrgID: s.orgID}
	before, err := s.repository.List(s.ctx, flt)
	s.Require().NoError(err)

	_, err = s.client.ExecContext(s.ctx, "UPDATE roles SET deleted_at = now() WHERE id = $1", deleted.ID)
	s.Require().NoError(err)

	_, err = s.repository.Get(s.ctx, deleted.ID)
	s.Assert().ErrorIs(err, role.ErrNotExist)

	_, err = s.repository.GetByName(s.ctx, deleted.OrgID, deleted.Name)
	s.Assert().ErrorIs(err, role.ErrNotExist)

	got, err := s.repository.List(s.ctx, flt)
	s.Assert().NoError(err)
	s.Assert().Len(got, len(before)-1)
	for _, r := range got {
		s.Assert().NotEqual(deleted.ID, r.ID)
	}

	updated := deleted
	updated.Title = "changed"
	_, err = s.repository.Update(s.ctx, updated)
	s.Assert().ErrorIs(err, role.ErrNotExist)
}
