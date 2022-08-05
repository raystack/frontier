package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
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

	// Add admin relation
	_, err = s.namespaceRepository.Create(context.Background(), namespace.DefinitionUser)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = s.namespaceRepository.Create(context.Background(), namespace.DefinitionTeam)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = s.roleRepository.Create(context.Background(), role.DefinitionTeamMember)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = s.relationRepository.Create(context.Background(), relation.Relation{
		SubjectNamespaceID: namespace.DefinitionUser.ID,
		SubjectID:          s.users[0].ID,
		ObjectNamespaceID:  namespace.DefinitionTeam.ID,
		ObjectID:           s.groups[0].ID,
		RoleID:             role.DefinitionTeamMember.ID,
		RelationType:       relation.RelationTypes.Role,
	})
	if err != nil {
		s.T().Fatal(err)
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
				Name:           "group1",
				Slug:           "group-1",
				OrganizationID: s.groups[0].OrganizationID,
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
				"Organization", // TODO need to do deeper comparison
				"Metadata",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestGetBySlug() {
	type testCase struct {
		Description   string
		SelectedSlug  string
		ExpectedGroup group.Group
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description:  "should get a group",
			SelectedSlug: "group-1",
			ExpectedGroup: group.Group{
				Name:           "group1",
				Slug:           "group-1",
				OrganizationID: s.groups[0].OrganizationID,
			},
		},
		{
			Description:  "should return error no exist if can't found group",
			SelectedSlug: "randomslug",
			ErrString:    group.ErrNotExist.Error(),
		},
		{
			Description: "should return error if slug empty",
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetBySlug(s.ctx, tc.SelectedSlug)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
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
				Slug:           "new-group-slug",
				OrganizationID: s.orgs[0].ID,
			},
			ExpectedGroup: group.Group{
				Name:           "new-group",
				Slug:           "new-group-slug",
				OrganizationID: s.orgs[0].ID,
			},
		},
		{
			Description: "should return error if group name already exist",
			GroupToCreate: group.Group{
				Name:           "group2",
				Slug:           "new-slug",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if group slug already exist",
			GroupToCreate: group.Group{
				Name:           "newslug",
				Slug:           "group-2",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if org id not an uuid",
			GroupToCreate: group.Group{
				Name:           "newslug",
				Slug:           "groupnewslug",
				OrganizationID: "some-id",
			},
			ErrString: group.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id does not exist",
			GroupToCreate: group.Group{
				Name:           "newslug",
				Slug:           "groupnewslug",
				OrganizationID: uuid.NewString(),
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
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
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
					Name:           "group1",
					Slug:           "group-1",
					OrganizationID: s.orgs[0].ID,
				},
				{
					Name:           "group2",
					Slug:           "group-2",
					OrganizationID: s.orgs[0].ID,
				},
				{
					Name:           "group3",
					Slug:           "group-3",
					OrganizationID: s.orgs[1].ID,
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
					Name:           "group3",
					Slug:           "group-3",
					OrganizationID: s.orgs[1].ID,
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
			if !cmp.Equal(got, tc.ExpectedGroups, cmpopts.IgnoreFields(group.Group{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
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
				Name:           "new group update",
				Slug:           "new-group-update",
				OrganizationID: s.orgs[0].ID,
			},
			ExpectedGroup: group.Group{
				Name:           "new group update",
				Slug:           "new-group-update",
				OrganizationID: s.orgs[0].ID,
			},
		},
		{
			Description: "should return error if group name already exist",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Name:           "group2",
				Slug:           "new-slug",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if group slug already exist",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Name:           "new-group-2",
				Slug:           "group-2",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if group not found",
			GroupToUpdate: group.Group{
				ID:             uuid.NewString(),
				Name:           "not-exist",
				Slug:           "some-slug",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrNotExist.Error(),
		},
		{
			Description: "should return error if group id is not uuid",
			GroupToUpdate: group.Group{
				ID:             "12345",
				Name:           "not-exist",
				Slug:           "some-slug",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Slug:           "new-prj",
				Name:           "not-exist",
				OrganizationID: "not-uuid",
			},
			ErrString: organization.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id not exist",
			GroupToUpdate: group.Group{
				ID:             s.groups[0].ID,
				Slug:           "new-prj",
				Name:           "not-exist",
				OrganizationID: uuid.NewString(),
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
			}
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestUpdateBySlug() {
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
				Name:           "new group update",
				Slug:           "group-1",
				OrganizationID: s.orgs[0].ID,
			},
			ExpectedGroup: group.Group{
				Name:           "new group update",
				Slug:           "group-1",
				OrganizationID: s.orgs[0].ID,
			},
		},
		{
			Description: "should return error if group name already exist",
			GroupToUpdate: group.Group{
				Name:           "group2",
				Slug:           "group-1",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrConflict.Error(),
		},
		{
			Description: "should return error if group not found",
			GroupToUpdate: group.Group{
				Slug:           "slug",
				Name:           "not-exist",
				OrganizationID: s.orgs[0].ID,
			},
			ErrString: group.ErrNotExist.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			GroupToUpdate: group.Group{
				Slug:           "group-1",
				Name:           "not-exist",
				OrganizationID: "not-uuid",
			},
			ErrString: organization.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id not exist",
			GroupToUpdate: group.Group{
				Slug:           "group-1",
				Name:           "not-exist",
				OrganizationID: uuid.NewString(),
			},
			ErrString: organization.ErrNotExist.Error(),
		},
		{
			Description: "should return error if group slug is empty",
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateBySlug(s.ctx, tc.GroupToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedGroup, cmpopts.IgnoreFields(group.Group{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroup)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestListUsersByGroupID() {
	type testCase struct {
		Description   string
		RoleID        string
		GroupID       string
		ExpectedUsers []user.User
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should return list of users",
			RoleID:      role.DefinitionTeamMember.ID,
			GroupID:     s.groups[0].ID,
			ExpectedUsers: []user.User{
				{
					Name:  "John Doe",
					Email: "john.doe@odpf.io",
				},
			},
		},
		{
			Description: "should get empty users if group does not have users",
			RoleID:      role.DefinitionTeamMember.ID,
			GroupID:     s.groups[1].ID,
		},
		{
			Description: "should not return error if role id is empty",
			GroupID:     s.groups[0].ID,
			ExpectedUsers: []user.User{
				{
					Name:  "John Doe",
					Email: "john.doe@odpf.io",
				},
			},
		},
		{
			Description: "should get error if group id is empty",
			RoleID:      role.DefinitionTeamMember.ID,
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.ListUsersByGroupID(s.ctx, tc.GroupID, tc.RoleID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUsers, cmpopts.IgnoreFields(user.User{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUsers)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestListUsersByGroupSlug() {
	type testCase struct {
		Description   string
		RoleID        string
		GroupSlug     string
		ExpectedUsers []user.User
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should return list of users",
			RoleID:      role.DefinitionTeamMember.ID,
			GroupSlug:   s.groups[0].Slug,
			ExpectedUsers: []user.User{
				{
					Name:  "John Doe",
					Email: "john.doe@odpf.io",
				},
			},
		},
		{
			Description: "should get empty users if group does not have users",
			RoleID:      role.DefinitionTeamMember.ID,
			GroupSlug:   s.groups[1].Slug,
		},
		{
			Description: "should not return error if role id is empty",
			GroupSlug:   s.groups[0].Slug,
			ExpectedUsers: []user.User{
				{
					Name:  "John Doe",
					Email: "john.doe@odpf.io",
				},
			},
		},
		{
			Description: "should get error if group id is empty",
			RoleID:      role.DefinitionTeamMember.ID,
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.ListUsersByGroupSlug(s.ctx, tc.GroupSlug, tc.RoleID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUsers, cmpopts.IgnoreFields(user.User{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUsers)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestListUserGroupIDRelations() {
	type testCase struct {
		Description       string
		UserID            string
		GroupID           string
		ExpectedRelations []relation.Relation
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should return list of relations if any",
			UserID:      s.users[0].ID,
			GroupID:     s.groups[0].ID,
			ExpectedRelations: []relation.Relation{
				{
					SubjectNamespaceID: namespace.DefinitionUser.ID,
					ObjectNamespaceID:  namespace.DefinitionTeam.ID,
					SubjectID:          s.users[0].ID,
					ObjectID:           s.groups[0].ID,
					RoleID:             role.DefinitionTeamMember.ID,
				},
			},
		},
		{
			Description: "should get empty relations if there is none",
			UserID:      s.users[0].ID,
			GroupID:     s.groups[1].ID,
		},
		{
			Description: "should get error if user id is empty",
			GroupID:     s.groups[0].ID,
			ErrString:   group.ErrInvalidID.Error(),
		},
		{
			Description: "should get error if group id is empty",
			UserID:      s.users[0].ID,
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.ListUserGroupIDRelations(s.ctx, tc.UserID, tc.GroupID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedRelations, cmpopts.IgnoreFields(relation.Relation{},
				"ID",
				"Role",
				"RelationType",
				"SubjectNamespace",
				"ObjectNamespace",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRelations)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestListUserGroupSlugRelations() {
	type testCase struct {
		Description       string
		UserID            string
		GroupSlug         string
		ExpectedRelations []relation.Relation
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should return list of relations if any",
			UserID:      s.users[0].ID,
			GroupSlug:   s.groups[0].Slug,
			ExpectedRelations: []relation.Relation{
				{
					SubjectNamespaceID: namespace.DefinitionUser.ID,
					ObjectNamespaceID:  namespace.DefinitionTeam.ID,
					SubjectID:          s.users[0].ID,
					ObjectID:           s.groups[0].ID,
					RoleID:             role.DefinitionTeamMember.ID,
				},
			},
		},
		{
			Description: "should get empty relations if there is none",
			UserID:      s.users[0].ID,
			GroupSlug:   s.groups[1].Slug,
		},
		{
			Description: "should get error if user id is empty",
			GroupSlug:   s.groups[1].Slug,
			ErrString:   group.ErrInvalidID.Error(),
		},
		{
			Description: "should get error if group id is empty",
			UserID:      s.users[0].ID,
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.ListUserGroupSlugRelations(s.ctx, tc.UserID, tc.GroupSlug)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedRelations, cmpopts.IgnoreFields(relation.Relation{},
				"ID",
				"Role",
				"RelationType",
				"SubjectNamespace",
				"ObjectNamespace",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRelations)
			}
		})
	}
}

func (s *GroupRepositoryTestSuite) TestListUserGroups() {
	type testCase struct {
		Description    string
		UserID         string
		RoleID         string
		ExpectedGroups []group.Group
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should return list of groups if any",
			UserID:      s.users[0].ID,
			RoleID:      role.DefinitionTeamMember.ID,
			ExpectedGroups: []group.Group{
				{
					Name:           "group1",
					Slug:           "group-1",
					OrganizationID: s.orgs[0].ID,
				},
			},
		},
		{
			Description: "should not return error if role id is empty",
			UserID:      s.users[0].ID,
			ExpectedGroups: []group.Group{
				{
					Name:           "group1",
					Slug:           "group-1",
					OrganizationID: s.orgs[0].ID,
				},
			},
		},
		{
			Description: "should get empty groups if there is none",
			UserID:      s.users[1].ID,
			RoleID:      role.DefinitionTeamMember.ID,
		},
		{
			Description: "should not return error if role id is empty",
			UserID:      s.users[0].ID,
			ExpectedGroups: []group.Group{
				{
					Name:           "group1",
					Slug:           "group-1",
					OrganizationID: s.orgs[0].ID,
				},
			},
		},
		{
			Description: "should get error if user id is empty",
			RoleID:      role.DefinitionTeamMember.ID,
			ErrString:   group.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.ListUserGroups(s.ctx, tc.UserID, tc.RoleID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedGroups, cmpopts.IgnoreFields(group.Group{},
				"ID",
				"Organization",
				"Metadata",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedGroups)
			}
		})
	}
}

func TestGroupRepository(t *testing.T) {
	suite.Run(t, new(GroupRepositoryTestSuite))
}
