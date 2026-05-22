package domain_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/domain/mocks"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_ListJoinableOrgsByDomain(t *testing.T) {
	ctx := context.Background()
	email := "alice@example.com"
	userID := "user-1"

	newService := func(t *testing.T) (*domain.Service, *mocks.Repository, *mocks.UserService, *mocks.MembershipService) {
		t.Helper()
		repo := mocks.NewRepository(t)
		userSvc := mocks.NewUserService(t)
		orgSvc := mocks.NewOrgService(t)
		memberSvc := mocks.NewMembershipService(t)
		svc := domain.NewService(slog.Default(), repo, userSvc, orgSvc, memberSvc)
		return svc, repo, userSvc, memberSvc
	}

	t.Run("returns verified-domain orgs the user is not a member of", func(t *testing.T) {
		svc, repo, userSvc, memberSvc := newService(t)

		repo.EXPECT().List(ctx, domain.Filter{Name: "example.com", State: domain.Verified}).
			Return([]domain.Domain{
				{OrgID: "org-1"},
				{OrgID: "org-2"},
				{OrgID: "org-3"},
			}, nil)

		userSvc.EXPECT().GetByID(ctx, email).Return(user.User{ID: userID}, nil)

		memberSvc.EXPECT().ListResourcesByPrincipal(ctx, authenticate.Principal{
			ID: userID, Type: schema.UserPrincipal,
		}, schema.OrganizationNamespace, mock.Anything).Return([]string{"org-2"}, nil)

		got, err := svc.ListJoinableOrgsByDomain(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, []string{"org-1", "org-3"}, got)
	})

	t.Run("excludes disabled-org policy-holder from joinable list", func(t *testing.T) {
		// A stale policy on a disabled org still counts as membership —
		// otherwise we'd offer the disabled org as joinable.
		svc, repo, userSvc, memberSvc := newService(t)

		repo.EXPECT().List(ctx, domain.Filter{Name: "example.com", State: domain.Verified}).
			Return([]domain.Domain{{OrgID: "org-disabled"}}, nil)

		userSvc.EXPECT().GetByID(ctx, email).Return(user.User{ID: userID}, nil)

		// Membership returns the disabled org because it's policy-based, not state-aware.
		memberSvc.EXPECT().ListResourcesByPrincipal(ctx, mock.Anything, schema.OrganizationNamespace, mock.Anything).
			Return([]string{"org-disabled"}, nil)

		got, err := svc.ListJoinableOrgsByDomain(ctx, email)
		assert.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("returns all verified-domain orgs when user has no memberships", func(t *testing.T) {
		svc, repo, userSvc, memberSvc := newService(t)

		repo.EXPECT().List(ctx, domain.Filter{Name: "example.com", State: domain.Verified}).
			Return([]domain.Domain{{OrgID: "org-1"}, {OrgID: "org-2"}}, nil)

		userSvc.EXPECT().GetByID(ctx, email).Return(user.User{ID: userID}, nil)

		memberSvc.EXPECT().ListResourcesByPrincipal(ctx, mock.Anything, schema.OrganizationNamespace, mock.Anything).
			Return(nil, nil)

		got, err := svc.ListJoinableOrgsByDomain(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, []string{"org-1", "org-2"}, got)
	})
}
