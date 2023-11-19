package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/raystack/frontier/core/organization"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InvitationRespositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.InvitationRepository
	users      []user.User
	groups     []group.Group
	invites    []invitation.Invitation
	orgs       []organization.Organization
}

func (s *InvitationRespositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewInvitationRepository(logger, s.client)

	s.users, err = bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.orgs, err = bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.groups, err = bootstrapGroup(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *InvitationRespositoryTestSuite) SetupTest() {
	var err error
	s.invites, err = bootstrapInvitation(s.client, s.users, s.orgs, s.groups)
	require.NoError(s.T(), err)
}

func (s *InvitationRespositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *InvitationRespositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *InvitationRespositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_INVITATIONS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *InvitationRespositoryTestSuite) TestSet() {
	type testCase struct {
		name          string
		inviteToSet   invitation.Invitation
		ExpectedError error
	}

	var testCases = []testCase{
		{
			name: "set invitation",
			inviteToSet: invitation.Invitation{
				ID:        s.invites[0].ID,
				UserID:    s.users[0].ID,
				OrgID:     s.orgs[0].ID,
				GroupIDs:  []string{s.groups[0].ID},
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
			ExpectedError: nil,
		},
		{
			name: "set invitation with invalid id",
			inviteToSet: invitation.Invitation{
				ID:        uuid.Nil,
				UserID:    s.users[0].ID,
				OrgID:     s.orgs[0].ID,
				GroupIDs:  []string{s.groups[0].ID},
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
			ExpectedError: postgres.ErrInvalidID,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.repository.Set(s.ctx, tc.inviteToSet)
			if tc.ExpectedError != nil {
				require.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (s *InvitationRespositoryTestSuite) TestGet() {
	type testCase struct {
		Name           string
		SelectedID     uuid.UUID
		ExpectedInvite invitation.Invitation
		ExpectedError  error
	}

	var testCases = []testCase{
		{
			Name:       "get invitation",
			SelectedID: s.invites[0].ID,
			ExpectedInvite: invitation.Invitation{
				ID:        s.invites[0].ID,
				UserID:    s.invites[0].UserID,
				OrgID:     s.invites[0].OrgID,
				GroupIDs:  s.invites[0].GroupIDs,
				Metadata:  s.invites[0].Metadata,
				CreatedAt: s.invites[0].CreatedAt,
				ExpiresAt: s.invites[0].ExpiresAt,
			},
			ExpectedError: nil,
		},
		{
			Name:          "get invitation with non-existing id",
			SelectedID:    uuid.New(),
			ExpectedError: invitation.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			invite, err := s.repository.Get(s.ctx, tc.SelectedID)
			if tc.ExpectedError != nil {
				require.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.ExpectedInvite, invite)
			}
		})
	}
}

func (s *InvitationRespositoryTestSuite) TestList() {
	type testCase struct {
		Name           string
		SelectedUserID string
		OrgIDFilter    string
		ExpectedList   []invitation.Invitation
		ExpectedError  error
	}

	var testCases = []testCase{
		{
			Name:           "list invitations",
			SelectedUserID: s.invites[0].UserID,
			OrgIDFilter:    s.invites[0].OrgID,
			ExpectedList: []invitation.Invitation{
				{
					ID:        s.invites[0].ID,
					UserID:    s.invites[0].UserID,
					OrgID:     s.invites[0].OrgID,
					GroupIDs:  s.invites[0].GroupIDs,
					Metadata:  s.invites[0].Metadata,
					CreatedAt: s.invites[0].CreatedAt,
					ExpiresAt: s.invites[0].ExpiresAt,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "list invitations without userid and orgid filter",
			ExpectedList: []invitation.Invitation{
				{ID: s.invites[0].ID,
					UserID:    s.invites[0].UserID,
					OrgID:     s.invites[0].OrgID,
					GroupIDs:  s.invites[0].GroupIDs,
					Metadata:  s.invites[0].Metadata,
					CreatedAt: s.invites[0].CreatedAt,
					ExpiresAt: s.invites[0].ExpiresAt,
				},
				{
					ID:        s.invites[1].ID,
					UserID:    s.invites[1].UserID,
					OrgID:     s.invites[1].OrgID,
					GroupIDs:  s.invites[1].GroupIDs,
					Metadata:  s.invites[1].Metadata,
					CreatedAt: s.invites[1].CreatedAt,
					ExpiresAt: s.invites[1].ExpiresAt,
				},
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			invites, err := s.repository.List(s.ctx, invitation.Filter{
				UserID: tc.SelectedUserID,
				OrgID:  tc.OrgIDFilter,
			})
			if tc.ExpectedError != nil {
				require.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.ExpectedList, invites)
			}
		})
	}
}

func (s *InvitationRespositoryTestSuite) TestListByUser() {
	type testCase struct {
		Name           string
		SelectedUserID string
		ExpectedList   []invitation.Invitation
		ExpectedError  error
	}

	var testCases = []testCase{
		{
			Name:           "list invitations by user",
			SelectedUserID: s.invites[0].UserID,
			ExpectedList: []invitation.Invitation{
				{
					ID:        s.invites[0].ID,
					UserID:    s.invites[0].UserID,
					OrgID:     s.invites[0].OrgID,
					GroupIDs:  s.invites[0].GroupIDs,
					Metadata:  s.invites[0].Metadata,
					CreatedAt: s.invites[0].CreatedAt,
					ExpiresAt: s.invites[0].ExpiresAt,
				},
			},
			ExpectedError: nil,
		},
		{
			Name:           "list invitations by user with non-existing user id",
			SelectedUserID: uuid.New().String(),
			ExpectedList:   nil,
			ExpectedError:  nil,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			invites, err := s.repository.ListByUser(s.ctx, tc.SelectedUserID)
			if tc.ExpectedError != nil {
				require.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.ExpectedList, invites)
			}
		})
	}
}

func (s *InvitationRespositoryTestSuite) TestDelete() {
	type testCase struct {
		Name          string
		SelectedID    uuid.UUID
		ExpectedError error
	}

	var testCases = []testCase{
		{
			Name:          "delete invitation",
			SelectedID:    s.invites[0].ID,
			ExpectedError: nil,
		},
		{
			Name:          "delete invitation with non-existing id",
			SelectedID:    uuid.New(),
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			err := s.repository.Delete(s.ctx, tc.SelectedID)
			if tc.ExpectedError != nil {
				require.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (s *InvitationRespositoryTestSuite) TestGarbageCollect() {
	type testCase struct {
		Name          string
		ExpectedError error
	}

	var testCases = []testCase{
		{
			Name:          "garbage collect",
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			err := s.repository.GarbageCollect(s.ctx)
			if tc.ExpectedError != nil {
				require.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInvitationRespository(t *testing.T) {
	suite.Run(t, new(InvitationRespositoryTestSuite))
}
