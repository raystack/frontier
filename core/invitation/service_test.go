package invitation_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/invitation/mocks"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	mocks2 "github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockService(t *testing.T) (*mocks2.Dialer, *mocks.Repository, *mocks.OrganizationService, *mocks.GroupService,
	*mocks.UserService, *mocks.RelationService, *mocks.PolicyService, *mocks.PreferencesService) {
	t.Helper()
	dialer := mocks2.NewDialer(t)
	repo := mocks.NewRepository(t)
	userService := mocks.NewUserService(t)
	orgService := mocks.NewOrganizationService(t)
	groupService := mocks.NewGroupService(t)
	relationService := mocks.NewRelationService(t)
	policyService := mocks.NewPolicyService(t)
	prefService := mocks.NewPreferencesService(t)
	return dialer, repo, orgService, groupService, userService, relationService, policyService, prefService
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *invitation.Service
		inviteToCreate invitation.Invitation
		want           invitation.Invitation
		err            error
	}{
		{
			name: "don't create an invite for already existing user in an organization",
			inviteToCreate: invitation.Invitation{
				UserEmailID: "test@example.com",
				OrgID:       "org-id",
			},
			err: invitation.ErrAlreadyMember,
			setup: func() *invitation.Service {
				dialer, repo, orgService, groupService, userService, relationService, policyService, prefService := mockService(t)

				prefService.EXPECT().LoadPlatformPreferences(mock.Anything).Return(map[string]string{}, nil)
				orgService.EXPECT().Get(mock.Anything, "org-id").Return(organization.Organization{
					ID: "org-id",
				}, nil)
				orgService.EXPECT().ListByUser(mock.Anything, "user-id", organization.Filter{}).Return([]organization.Organization{
					{
						ID: "org-id",
					},
				}, nil)

				userService.EXPECT().GetByID(context.Background(), "test@example.com").Return(user.User{
					ID:    "user-id",
					Email: "test@example.com",
				}, nil)

				return invitation.NewService(dialer, repo, orgService, groupService,
					userService, relationService, policyService, prefService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(context.Background(), tt.inviteToCreate)
			if (err != nil) && tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Create() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
