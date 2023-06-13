package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/shield/internal/bootstrap/schema"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/core/group"
	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/internal/store/postgres"
	"github.com/raystack/shield/pkg/db"
	"github.com/raystack/shield/pkg/metadata"
	"github.com/stretchr/testify/suite"
)

type GroupRepositoryTestSuite struct {
	suite.Suite
	ctx                 context.Context
	client              *db.Client
	pool                *dockertest.Pool
	resource            *dockertest.Resource
	repository          *postgres.GroupRepository
	relationRepository  *postgres.RelationRepository
	namespaceRepository *postgres.NamespaceRepository
	roleRepository      *postgres.RoleRepository
	orgs                []organization.Organization
	groups              []group.Group
	users               []user.User
}

func (s *GroupRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewGroupRepository(s.client)

	s.users, err = bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.relationRepository = postgres.NewRelationRepository(s.client)
	s.namespaceRepository = postgres.NewNamespaceRepository(s.client)
	s.roleRepository = postgres.NewRoleRepository(s.client)

	s.orgs, err = bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *GroupRepositoryTestSuite) SetupTest() {
	var err error
	s.groups, err = bootstrapGroup(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapPermissions(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapRole(s.client, s.orgs[0].ID)
	if err != nil {
		s.T().Fatal(err)
	}

	for _, group := range s.groups {
		_, err = s.relationRepository.Upsert(context.Background(), relation.Relation{
			Subject: relation.Subject{
				ID:        s.users[0].ID,
				Namespace: schema.UserPrincipal,
			},
			Object: relation.Object{
				ID:        group.ID,
				Namespace: schema.GroupNamespace,
			},
			RelationName: schema.MemberRelationName,
		})
		if err != nil {
			s.T().Fatal(err)
		}
	}

	for _, user := range s.users {
		_, err = s.relationRepository.Upsert(context.Background(), relation.Relation{
			Subject: relation.Subject{
				ID:        user.ID,
				Namespace: schema.UserPrincipal,
			},
			Object: relation.Object{
				ID:        s.groups[0].ID,
				Namespace: schema.GroupNamespace,
			},
			RelationName: schema.MemberRelationName,
		})
		if err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *GroupRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *GroupRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *GroupRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_GROUPS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_RELATIONS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ROLES),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_NAMESPACES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *GroupRepositoryTestSuite) TestGetByID() {
	type testCase struct {
		Description   string
		SelectedID    string
		ExpectedGroup group.Group
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should get a group",
			SelectedID:  s.groups[0].ID,
			ExpectedGroup: group.Group{
				Name:           "group-1",
				OrganizationID: s.groups[0].OrganizationID,
				State:          group.Enabled,
			},
		},
		{
			Description: "should return error no exist if can't found group",
			SelectedID:  uuid.NewString(),
			ErrString:   group.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id empty",
			ErrString:   group.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "10000",
			ErrString:   group.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByID(s.ctx, tc.SelectedID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{},
				"ID",
				"Metadata",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestGetByIDs() {
	type testCase struct {
		Description    string
		SelectedIDs    []string
		Filter         group.Filter
		ExpectedGroups []group.Group
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should list groups from all orgs if no filter applied",
			SelectedIDs: []string{s.groups[0].ID, s.groups[1].ID, s.groups[2].ID},
			Filter:      group.Filter{},
			ExpectedGroups: []group.Group{{
				Name:           "group-1",
				OrganizationID: s.orgs[0].ID,
				State:          group.Enabled,
			}, {
				Name:           "group-2",
				OrganizationID: s.orgs[0].ID,
				State:          group.Enabled,
			}, {
				Name:           "group-3",
				OrganizationID: s.orgs[1].ID,
				State:          group.Enabled,
			},
			},
		},
		{
			Description: "should list groups from a single org with org filter",
			SelectedIDs: []string{s.groups[0].ID, s.groups[1].ID, s.groups[2].ID},
			Filter: group.Filter{
				OrganizationID: s.orgs[1].ID,
			},
			ExpectedGroups: []group.Group{{
				Name:           "group-3",
				OrganizationID: s.orgs[1].ID,
				State:          group.Enabled,
			},
			},
		},
		{
			Description: "should return error if id empty",
			SelectedIDs: []string{s.groups[0].ID, ""},
			ErrString:   group.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedIDs: []string{s.groups[0].ID, "10000"},
			ErrString:   group.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByIDs(s.ctx, tc.SelectedIDs, tc.Filter)
			if tc.ErrString != "" && err != nil {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}

			s.Assert().Len(got, len(tc.ExpectedGroups))
			for i, grp := range got {
				if !cmp.Equal(grp, tc.ExpectedGroups[i], cmpopts.IgnoreFields(group.Group{},
					"ID",
					"Metadata",
					"CreatedAt",
					"UpdatedAt")) {
					s.T().Fatalf("got result %+v, expected was %+v", grp, tc.ExpectedGroups[i])
				}
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description   string
		GroupToCreate group.Group
		ExpectedGroup group.Group
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should create a group",
			GroupToCreate: group.Group{
				Name:           "new-group",
				OrganizationID: s.orgs[0].ID,
				Metadata:       metadata.Metadata{},
			},
			ExpectedGroup: group.Group{
				Name:           "new-group",
				OrganizationID: s.orgs[0].ID,
				State:          group.Enabled,
				Metadata:       metadata.Metadata{},
			},
		},
		{
			Description: "should not return error if group name already exist in different org",
			GroupToCreate: group.Group{
				Name:           "group-2",
				OrganizationID: s.orgs[1].ID,
				Metadata:       metadata.Metadata{},
			},
			ExpectedGroup: group.Group{
				Name:           "group-2",
				OrganizationID: s.orgs[1].ID,
				State:          group.Enabled,
				Metadata:       metadata.Metadata{},
			},
		},
		{
			Description: "should return error if group name already exist in same org",
			GroupToCreate: group.Group{
				Name:           "group-2",
				OrganizationID: s.orgs[0].ID,
				Metadata:       metadata.Metadata{},
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if org id not an uuid",
			GroupToCreate: group.Group{
				Name:           "newslug",
				OrganizationID: "some-id",
				Metadata:       metadata.Metadata{},
			},
			ErrString: group.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id does not exist",
			GroupToCreate: group.Group{
				Name:           "newslug",
				OrganizationID: uuid.NewString(),
				Metadata:       metadata.Metadata{},
			},
			ErrString: organization.ErrNotExist.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.GroupToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestList() {
	type testCase struct {
		Description    string
		Filter         group.Filter
		ExpectedGroups []group.Group
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should get all groups",
			ExpectedGroups: []group.Group{
				{
					Name:           "group-1",
					OrganizationID: s.orgs[0].ID,
					State:          group.Enabled,
					Metadata:       metadata.Metadata{},
				},
				{
					Name:           "group-2",
					OrganizationID: s.orgs[0].ID,
					State:          group.Enabled,
					Metadata:       metadata.Metadata{},
				},
				{
					Name:           "group-3",
					OrganizationID: s.orgs[1].ID,
					State:          group.Enabled,
					Metadata:       metadata.Metadata{},
				},
			},
		},
		{
			Description: "should get filtered groups",
			Filter: group.Filter{
				OrganizationID: s.orgs[1].ID,
			},
			ExpectedGroups: []group.Group{
				{
					Name:           "group-3",
					OrganizationID: s.orgs[1].ID,
					State:          group.Enabled,
					Metadata:       metadata.Metadata{},
				},
			},
		},
		{
			Description: "should get filtered groups for disabled state",
			Filter: group.Filter{
				State: group.Disabled,
			},
			ExpectedGroups: []group.Group{
				{
					Name:           "group-4",
					OrganizationID: s.orgs[1].ID,
					State:          group.Disabled,
					Metadata:       metadata.Metadata{},
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, tc.Filter)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			s.Assert().Len(got, len(tc.ExpectedGroups))
			if !cmp.Equal(got, tc.ExpectedGroups, cmpopts.IgnoreFields(group.Group{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroups)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestUpdateByID() {
	type testCase struct {
		Description   string
		GroupToUpdate group.Group
		ExpectedGroup group.Group
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should update a group",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Name:           "new-group-update",
				Title:          "new-group-update-name",
				OrganizationID: s.orgs[0].ID,
				Metadata:       metadata.Metadata{},
			},
			ExpectedGroup: group.Group{
				ID:             s.groups[0].ID,
				Name:           "new-group-update",
				Title:          "new-group-update-name",
				OrganizationID: s.orgs[0].ID,
				State:          group.Enabled,
				Metadata:       metadata.Metadata{},
			},
		},
		{
			Description: "should return error if group slug already exist",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Name:           "group-2",
				OrganizationID: s.orgs[0].ID,
				Metadata:       metadata.Metadata{},
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if group not found",
			GroupToUpdate: group.Group{
				ID:             uuid.NewString(),
				Name:           "not-exist",
				OrganizationID: s.orgs[0].ID,
				Metadata:       metadata.Metadata{},
			},
			ErrString: group.ErrNotExist.Error(),
		},
		{
			Description: "should return error if group id is not uuid",
			GroupToUpdate: group.Group{
				ID:             "12345",
				Name:           "not-exist",
				OrganizationID: s.orgs[0].ID,
				Metadata:       metadata.Metadata{},
			},
			ErrString: group.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Name:           "not-exist",
				OrganizationID: "not-uuid",
				Metadata:       metadata.Metadata{},
			},
			ErrString: organization.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id not exist",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Name:           "not-exist",
				OrganizationID: uuid.NewString(),
				Metadata:       metadata.Metadata{},
			},
			ErrString: organization.ErrNotExist.Error(),
		},
		{
			Description: "should return error if group id is empty",
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByID(s.ctx, tc.GroupToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			} else {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestListGroupRelations() {
	type testCase struct {
		Description       string
		ObjectID          string
		SubjectType       string
		Role              string
		ExpectedRelations []relation.Relation
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should get a list of relations",
			ObjectID:    s.groups[0].ID,
			SubjectType: "user",
			Role:        "member",
			ExpectedRelations: []relation.Relation{
				{
					Object: relation.Object{
						ID:        s.groups[0].ID,
						Namespace: schema.GroupNamespace,
					},
					Subject: relation.Subject{
						ID:              s.users[0].ID,
						Namespace:       schema.UserPrincipal,
						SubRelationName: "shield/group:member",
					},
				},
				{
					Object: relation.Object{
						ID:        s.groups[0].ID,
						Namespace: schema.GroupNamespace,
					},
					Subject: relation.Subject{
						ID:              s.users[1].ID,
						Namespace:       schema.UserPrincipal,
						SubRelationName: "shield/group:member",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.ListGroupRelations(s.ctx, tc.ObjectID, tc.SubjectType, tc.Role)
			if tc.ErrString != "" && err != nil {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}

			for _, rel := range got {
				found := false
				for _, expectedRel := range tc.ExpectedRelations {
					if cmp.Equal(rel, expectedRel, cmpopts.IgnoreFields(relation.Relation{},
						"ID",
						"CreatedAt",
						"UpdatedAt")) {
						found = true
						break
					}
				}
				if !found {
					s.T().Fatalf("can't find relation %+v", rel)
				}
			}
		})
	}
}

func TestGroupRepository(t *testing.T) {
	suite.Run(t, new(GroupRepositoryTestSuite))
}
