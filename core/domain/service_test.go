package domain_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/domain/mocks"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgauditrecord "github.com/raystack/frontier/pkg/auditrecord"
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
		svc := domain.NewService(slog.Default(), repo, userSvc, orgSvc, memberSvc, mocks.NewAuditRecordRepository(t))
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

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	dmn := domain.Domain{ID: "3f1c9a2e-8b4d-4e6f-9a1b-2c3d4e5f6a7b", Name: "acme.com", OrgID: "org-1"}
	org := organization.Organization{ID: "org-1", Name: "acme", Title: "Acme Inc"}
	errDB := errors.New("connection reset")

	newService := func(t *testing.T) (*domain.Service, *mocks.Repository, *mocks.OrgService, *mocks.AuditRecordRepository) {
		t.Helper()
		repo := mocks.NewRepository(t)
		orgSvc := mocks.NewOrgService(t)
		auditRepo := mocks.NewAuditRecordRepository(t)
		svc := domain.NewService(slog.Default(), repo, mocks.NewUserService(t), orgSvc, mocks.NewMembershipService(t), auditRepo)
		return svc, repo, orgSvc, auditRepo
	}

	t.Run("marks the domain deleted and writes a domain.deleted audit record", func(t *testing.T) {
		svc, repo, orgSvc, auditRepo := newService(t)
		repo.EXPECT().Get(ctx, dmn.ID).Return(dmn, nil)
		orgSvc.EXPECT().GetRaw(ctx, org.ID).Return(org, nil)
		repo.EXPECT().Delete(ctx, dmn.ID).Return(nil)

		var got auditmodels.AuditRecord
		auditRepo.EXPECT().Create(ctx, mock.Anything).
			Run(func(_ context.Context, record auditmodels.AuditRecord) { got = record }).
			Return(auditmodels.AuditRecord{}, nil)

		assert.NoError(t, svc.Delete(ctx, dmn.ID))

		assert.False(t, got.OccurredAt.IsZero())
		got.OccurredAt = time.Time{}
		assert.Equal(t, auditmodels.AuditRecord{
			Event:    pkgauditrecord.DomainDeletedEvent,
			Resource: auditmodels.Resource{ID: org.ID, Type: pkgauditrecord.OrganizationType, Name: org.Title},
			Target:   &auditmodels.Target{ID: dmn.ID, Type: pkgauditrecord.DomainType, Name: dmn.Name},
			OrgID:    org.ID,
			OrgName:  org.Title,
		}, got)
	})

	t.Run("an unknown domain writes no audit record", func(t *testing.T) {
		svc, repo, _, _ := newService(t)
		repo.EXPECT().Get(ctx, dmn.ID).Return(domain.Domain{}, domain.ErrNotExist)

		assert.ErrorIs(t, svc.Delete(ctx, dmn.ID), domain.ErrNotExist)
	})

	t.Run("a failed org lookup leaves the domain alone", func(t *testing.T) {
		svc, repo, orgSvc, _ := newService(t)
		repo.EXPECT().Get(ctx, dmn.ID).Return(dmn, nil)
		orgSvc.EXPECT().GetRaw(ctx, org.ID).Return(organization.Organization{}, errDB)

		assert.ErrorIs(t, svc.Delete(ctx, dmn.ID), errDB)
	})

	t.Run("a failed delete writes no audit record", func(t *testing.T) {
		svc, repo, orgSvc, _ := newService(t)
		repo.EXPECT().Get(ctx, dmn.ID).Return(dmn, nil)
		orgSvc.EXPECT().GetRaw(ctx, org.ID).Return(org, nil)
		repo.EXPECT().Delete(ctx, dmn.ID).Return(errDB)

		assert.ErrorIs(t, svc.Delete(ctx, dmn.ID), errDB)
	})

	t.Run("a failed audit write does not fail the delete", func(t *testing.T) {
		svc, repo, orgSvc, auditRepo := newService(t)
		repo.EXPECT().Get(ctx, dmn.ID).Return(dmn, nil)
		orgSvc.EXPECT().GetRaw(ctx, org.ID).Return(org, nil)
		repo.EXPECT().Delete(ctx, dmn.ID).Return(nil)
		auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditmodels.AuditRecord{}, errDB)

		assert.NoError(t, svc.Delete(ctx, dmn.ID))
	})
}
