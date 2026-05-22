package invitation_test

import (
	"context"
	"testing"
	"time"

	auditMocks "github.com/raystack/frontier/core/auditrecord/mocks"
	auditModels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/invitation/mocks"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	mocks2 "github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockService(t *testing.T) (*mocks2.Dialer, *mocks.Repository, *mocks.OrganizationService, *mocks.GroupService,
	*mocks.UserService, *mocks.RelationService, *mocks.PreferencesService, *auditMocks.Repository) {
	t.Helper()
	dialer := mocks2.NewDialer(t)
	repo := mocks.NewRepository(t)
	userService := mocks.NewUserService(t)
	orgService := mocks.NewOrganizationService(t)
	groupService := mocks.NewGroupService(t)
	relationService := mocks.NewRelationService(t)
	prefService := mocks.NewPreferencesService(t)
	auditRecordRepo := auditMocks.NewRepository(t)
	return dialer, repo, orgService, groupService, userService, relationService, prefService, auditRecordRepo
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
				dialer, repo, orgService, groupService, userService, relationService, prefService, auditRecordRepo := mockService(t)

				prefService.EXPECT().LoadPlatformPreferences(mock.Anything).Return(map[string]string{}, nil)
				orgService.EXPECT().Get(mock.Anything, "org-id").Return(organization.Organization{
					ID: "org-id",
				}, nil)

				userService.EXPECT().GetByID(context.Background(), "test@example.com").Return(user.User{
					ID:    "user-id",
					Email: "test@example.com",
				}, nil)

				membershipSvc := mocks.NewMembershipService(t)
				membershipSvc.EXPECT().ListResourcesByPrincipal(mock.Anything, authenticate.Principal{
					ID:   "user-id",
					Type: schema.UserPrincipal,
				}, schema.OrganizationNamespace, membership.ResourceFilter{}).Return([]string{"org-id"}, nil)
				return invitation.NewService(dialer, repo, orgService, groupService,
					userService, relationService, prefService, auditRecordRepo, membershipSvc)
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

func TestService_Accept_DedupesExistingGroupMembers(t *testing.T) {
	ctx := context.Background()
	inviteID := uuid.New()
	userID := "user-id"
	userEmail := "test@example.com"
	orgID := "org-id"

	dialer, repo, orgService, groupService, userService, relationService, prefService, auditRecordRepo := mockService(t)
	membershipSvc := mocks.NewMembershipService(t)

	repo.EXPECT().Get(ctx, inviteID).Return(invitation.Invitation{
		ID:          inviteID,
		UserEmailID: userEmail,
		OrgID:       orgID,
		GroupIDs:    []string{"g-alpha", "g-gamma"},
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil)

	userOb := user.User{ID: userID, Email: userEmail, Title: "Test User"}
	userPrincipal := authenticate.Principal{ID: userID, Type: schema.UserPrincipal}

	// isUserOrgMember — already a member, so AddOrganizationMember is skipped
	userService.EXPECT().GetByID(ctx, userEmail).Return(userOb, nil)
	membershipSvc.EXPECT().ListResourcesByPrincipal(ctx, userPrincipal, schema.OrganizationNamespace, membership.ResourceFilter{}).
		Return([]string{orgID}, nil)

	prefService.EXPECT().LoadPlatformPreferences(ctx).Return(map[string]string{}, nil)

	// User is already a member of g-alpha
	groupService.EXPECT().List(ctx, group.Filter{Principal: &userPrincipal}).
		Return([]group.Group{{ID: "g-alpha"}}, nil)

	// Both invite groups get looked up
	groupService.EXPECT().Get(ctx, "g-alpha").Return(group.Group{ID: "g-alpha"}, nil)
	groupService.EXPECT().Get(ctx, "g-gamma").Return(group.Group{ID: "g-gamma"}, nil)

	// Only g-gamma is added; g-alpha is skipped (no SetGroupMemberRole expectation for it,
	// so the mock would fail if the code called it)
	membershipSvc.EXPECT().
		SetGroupMemberRole(ctx, "g-gamma", userID, schema.UserPrincipal, schema.GroupMemberRole).
		Return(nil)

	// Audit + delete tail
	orgService.EXPECT().Get(ctx, orgID).Return(organization.Organization{ID: orgID, Title: "Test Org"}, nil)
	auditRecordRepo.EXPECT().Create(ctx, mock.AnythingOfType("models.AuditRecord")).
		Return(auditModels.AuditRecord{}, nil)
	relationService.EXPECT().Delete(ctx, mock.AnythingOfType("relation.Relation")).Return(nil)
	repo.EXPECT().Delete(ctx, inviteID).Return(nil)

	svc := invitation.NewService(dialer, repo, orgService, groupService,
		userService, relationService, prefService, auditRecordRepo, membershipSvc)

	err := svc.Accept(ctx, inviteID)
	assert.NoError(t, err)
}
