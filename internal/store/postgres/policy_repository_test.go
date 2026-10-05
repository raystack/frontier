package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
)

type PolicyRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.PolicyRepository
	policies   []policy.Policy
	userID     string
	orgID      string
	roles      []role.Role
}

func (s *PolicyRepositoryTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
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
	if err := closeTestClient(s.client); err != nil {
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

	s.Run("should create a new row when the same policy was soft-deleted", func() {
		pol := policy.Policy{
			RoleID:        s.roles[0].ID,
			ResourceID:    uuid.NewString(),
			ResourceType:  "ns1",
			PrincipalID:   s.userID,
			PrincipalType: schema.UserPrincipal,
		}
		first, err := s.repository.Upsert(s.ctx, pol)
		s.Require().NoError(err)
		if _, err := s.client.ExecContext(s.ctx, "UPDATE policies SET deleted_at = now() WHERE id = $1", first.ID); err != nil {
			s.T().Fatal(err)
		}

		second, err := s.repository.Upsert(s.ctx, pol)
		s.Assert().NoError(err)
		s.Assert().NotEqual(first.ID, second.ID)

		var firstStillDeleted bool
		if err := s.client.QueryRowxContext(s.ctx, "SELECT deleted_at IS NOT NULL FROM policies WHERE id = $1", first.ID).Scan(&firstStillDeleted); err != nil {
			s.T().Fatal(err)
		}
		s.Assert().True(firstStillDeleted)

		_, err = s.repository.Get(s.ctx, second.ID)
		s.Assert().NoError(err)
	})

	s.Run("should update the live policy in place when the same policy exists", func() {
		pol := policy.Policy{
			RoleID:        s.roles[0].ID,
			ResourceID:    uuid.NewString(),
			ResourceType:  "ns1",
			PrincipalID:   s.userID,
			PrincipalType: schema.UserPrincipal,
		}
		before, err := s.repository.Upsert(s.ctx, pol)
		s.Require().NoError(err)

		pol.Metadata = metadata.Metadata{"team": "maps"}
		got, err := s.repository.Upsert(s.ctx, pol)
		s.Assert().NoError(err)
		s.Assert().Equal(before.ID, got.ID)
		s.Assert().Equal(metadata.Metadata{"team": "maps"}, got.Metadata)
		s.Assert().True(got.UpdatedAt.After(before.UpdatedAt))
	})

	newRole := func(name string) role.Role {
		created, err := postgres.NewRoleRepository(s.client).Upsert(s.ctx, role.Role{
			Name:     name,
			OrgID:    s.orgID,
			Metadata: metadata.Metadata{},
		})
		s.Require().NoError(err)
		return created
	}

	s.Run("should return not found when the role is soft-deleted", func() {
		deleted := newRole("soft-deleted role for a policy create")
		_, err := s.client.ExecContext(s.ctx, "UPDATE roles SET deleted_at = now() WHERE id = $1", deleted.ID)
		s.Require().NoError(err)

		_, err = s.repository.Upsert(s.ctx, policy.Policy{
			RoleID:        deleted.ID,
			ResourceID:    uuid.NewString(),
			ResourceType:  "ns1",
			PrincipalID:   uuid.NewString(),
			PrincipalType: schema.UserPrincipal,
		})
		s.Assert().ErrorIs(err, role.ErrNotExist)

		var created int
		err = s.client.QueryRowxContext(s.ctx, "SELECT count(*) FROM policies WHERE role_id = $1", deleted.ID).Scan(&created)
		s.Require().NoError(err)
		s.Assert().Equal(0, created)
	})

	s.Run("should wait for a role delete that has not committed and then return not found", func() {
		target := newRole("role with a running delete")
		tx, err := s.client.BeginTxx(s.ctx, nil)
		s.Require().NoError(err)
		defer tx.Rollback() // nolint
		var lockedID string
		err = tx.QueryRowContext(s.ctx, "SELECT id FROM roles WHERE id = $1 AND deleted_at IS NULL FOR UPDATE", target.ID).Scan(&lockedID)
		s.Require().NoError(err)
		_, err = tx.ExecContext(s.ctx, "UPDATE roles SET deleted_at = now() WHERE id = $1", target.ID)
		s.Require().NoError(err)

		done := make(chan error, 1)
		go func() {
			_, err := s.repository.Upsert(s.ctx, policy.Policy{
				RoleID:        target.ID,
				ResourceID:    uuid.NewString(),
				ResourceType:  "ns1",
				PrincipalID:   uuid.NewString(),
				PrincipalType: schema.UserPrincipal,
			})
			done <- err
		}()

		s.Require().Eventually(func() bool {
			var waiting int
			err := s.client.QueryRowxContext(s.ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'").Scan(&waiting)
			return err == nil && waiting == 1
		}, time.Second, 10*time.Millisecond, "the create did not wait for the role delete")

		s.Require().NoError(tx.Commit())
		s.Assert().ErrorIs(<-done, role.ErrNotExist)
	})

	s.Run("should not wait for another policy create on the same role", func() {
		target := newRole("role with two policy creates")
		tx, err := s.client.BeginTxx(s.ctx, nil)
		s.Require().NoError(err)
		defer tx.Rollback() // nolint
		_, err = tx.ExecContext(s.ctx, "INSERT INTO policies (role_id, resource_id, resource_type, principal_id, principal_type) VALUES ($1, $2, 'ns1', $3, 'app/user')",
			target.ID, s.orgID, uuid.NewString())
		s.Require().NoError(err)

		_, err = s.repository.Upsert(s.ctx, policy.Policy{
			RoleID:        target.ID,
			ResourceID:    s.orgID,
			ResourceType:  "ns1",
			PrincipalID:   uuid.NewString(),
			PrincipalType: schema.UserPrincipal,
		})
		s.Assert().NoError(err)
	})
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
					RoleID:        s.roles[0].ID,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
					ResourceID:    s.orgID,
					ResourceType:  "ns1",
				},
				{
					RoleID:        s.roles[0].ID,
					PrincipalID:   s.userID,
					PrincipalType: schema.UserPrincipal,
					ResourceID:    s.orgID,
					ResourceType:  "ns2",
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
			if diff := cmp.Diff(tc.ExpectedPolicys, got,
				cmpopts.IgnoreFields(policy.Policy{}, "ID", "CreatedAt", "UpdatedAt"),
				cmpopts.SortSlices(func(a, b policy.Policy) bool { return a.ResourceType < b.ResourceType }),
			); diff != "" {
				s.T().Fatalf("mismatch (-want +got):\n%s", diff)
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

	newPolicy := func() policy.Policy {
		created, err := s.repository.Upsert(s.ctx, policy.Policy{
			RoleID:        s.roles[0].ID,
			ResourceID:    uuid.NewString(),
			ResourceType:  "ns1",
			PrincipalID:   s.userID,
			PrincipalType: schema.UserPrincipal,
		})
		s.Require().NoError(err)
		return created
	}
	kept := newPolicy()

	s.Run("should keep the row and mark it deleted", func() {
		s.Assert().NoError(s.repository.Delete(s.ctx, kept.ID))

		var markedDeleted bool
		if err := s.client.QueryRowxContext(s.ctx, "SELECT deleted_at IS NOT NULL FROM policies WHERE id = $1", kept.ID).Scan(&markedDeleted); err != nil {
			s.T().Fatal(err)
		}
		s.Assert().True(markedDeleted)

		_, err := s.repository.Get(s.ctx, kept.ID)
		s.Assert().ErrorIs(err, policy.ErrNotExist)
	})

	s.Run("should return not found when the policy is already deleted", func() {
		s.Assert().ErrorIs(s.repository.Delete(s.ctx, kept.ID), policy.ErrNotExist)
	})
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
			Description: "count only human users and not service users or groups",
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

func (s *PolicyRepositoryTestSuite) TestSkipsSoftDeletedPolicies() {
	deleted := s.policies[0]
	flt := policy.Filter{PrincipalID: s.userID}
	before, err := s.repository.List(s.ctx, flt)
	s.Require().NoError(err)
	countBefore, err := s.repository.Count(s.ctx, flt)
	s.Require().NoError(err)

	_, err = s.client.ExecContext(s.ctx, "UPDATE policies SET deleted_at = now() WHERE id = $1", deleted.ID)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = s.repository.Get(s.ctx, deleted.ID)
	s.Assert().ErrorIs(err, policy.ErrNotExist)

	got, err := s.repository.List(s.ctx, flt)
	s.Assert().NoError(err)
	s.Assert().Len(got, len(before)-1)
	for _, p := range got {
		s.Assert().NotEqual(deleted.ID, p.ID)
	}

	count, err := s.repository.Count(s.ctx, flt)
	s.Assert().NoError(err)
	s.Assert().Equal(countBefore-1, count)

	_, err = s.repository.Update(s.ctx, deleted)
	s.Assert().ErrorIs(err, policy.ErrNotExist)
}

func (s *PolicyRepositoryTestSuite) TestDeleteWithMinRoleGuardCountsLiveHoldersOnly() {
	guardRole := s.roles[0].ID
	resourceID := uuid.NewString()
	newHolder := func() policy.Policy {
		created, err := s.repository.Upsert(s.ctx, policy.Policy{
			RoleID:        guardRole,
			ResourceID:    resourceID,
			ResourceType:  schema.OrganizationNamespace,
			PrincipalID:   uuid.NewString(),
			PrincipalType: schema.UserPrincipal,
		})
		s.Require().NoError(err)
		return created
	}
	first, second, hidden := newHolder(), newHolder(), newHolder()

	_, err := s.client.ExecContext(s.ctx, "UPDATE policies SET deleted_at = now() WHERE id = $1", hidden.ID)
	s.Require().NoError(err)

	// two live holders remain, so removing one is allowed
	s.Assert().NoError(s.repository.DeleteWithMinRoleGuard(s.ctx, first.ID, guardRole))

	// only the soft-deleted holder is left besides this one, so it must not count
	err = s.repository.DeleteWithMinRoleGuard(s.ctx, second.ID, guardRole)
	s.Assert().ErrorIs(err, policy.ErrLastRoleGuard)

	_, err = s.repository.Get(s.ctx, second.ID)
	s.Assert().NoError(err)
}

func (s *PolicyRepositoryTestSuite) TestDeleteWithMinRoleGuardKeepsTheRow() {
	guardRole := s.roles[0].ID
	resourceID := uuid.NewString()
	newHolder := func() policy.Policy {
		created, err := s.repository.Upsert(s.ctx, policy.Policy{
			RoleID:        guardRole,
			ResourceID:    resourceID,
			ResourceType:  schema.OrganizationNamespace,
			PrincipalID:   uuid.NewString(),
			PrincipalType: schema.UserPrincipal,
		})
		s.Require().NoError(err)
		return created
	}
	first, second := newHolder(), newHolder()

	s.Run("should keep the row and mark it deleted", func() {
		s.Assert().NoError(s.repository.DeleteWithMinRoleGuard(s.ctx, first.ID, guardRole))

		var markedDeleted bool
		if err := s.client.QueryRowxContext(s.ctx, "SELECT deleted_at IS NOT NULL FROM policies WHERE id = $1", first.ID).Scan(&markedDeleted); err != nil {
			s.T().Fatal(err)
		}
		s.Assert().True(markedDeleted)

		_, err := s.repository.Get(s.ctx, first.ID)
		s.Assert().ErrorIs(err, policy.ErrNotExist)

		_, err = s.repository.Get(s.ctx, second.ID)
		s.Assert().NoError(err)
	})

	s.Run("should return not found instead of the guard error when the policy is already deleted", func() {
		err := s.repository.DeleteWithMinRoleGuard(s.ctx, first.ID, guardRole)
		s.Assert().ErrorIs(err, policy.ErrNotExist)
		s.Assert().NotErrorIs(err, policy.ErrLastRoleGuard)
	})
}
