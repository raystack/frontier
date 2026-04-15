package userpat

import (
	"context"
	"errors"
	"testing"
	"time"

	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/userpat/mocks"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/pkg/db"
	mailerMock "github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newAlertMocks(t *testing.T) (
	*mocks.AlertRepository,
	*mocks.AlertUserService,
	*mocks.AlertOrgService,
	*mailerMock.Dialer,
	*mocks.Locker,
	*mocks.AlertAuditRepository,
) {
	t.Helper()
	return mocks.NewAlertRepository(t),
		mocks.NewAlertUserService(t),
		mocks.NewAlertOrgService(t),
		mailerMock.NewDialer(t),
		mocks.NewLocker(t),
		mocks.NewAlertAuditRepository(t)
}

func newAlertService(t *testing.T, cfg AlertConfig) (*AlertService, *mocks.AlertRepository, *mocks.AlertUserService, *mocks.AlertOrgService, *mailerMock.Dialer, *mocks.Locker, *mocks.AlertAuditRepository) {
	t.Helper()
	repo, userSvc, orgSvc, dialer, locker, auditRepo := newAlertMocks(t)
	svc := NewAlertService(repo, userSvc, orgSvc, dialer, locker, cfg, log.NewNoop(), auditRepo)
	return svc, repo, userSvc, orgSvc, dialer, locker, auditRepo
}

func TestAlertService_Init(t *testing.T) {
	t.Run("disabled config does nothing", func(t *testing.T) {
		svc, _, _, _, _, _, _ := newAlertService(t, AlertConfig{Enabled: false})
		assert.NoError(t, svc.Init(context.Background()))
		assert.NoError(t, svc.Close())
	})

	t.Run("valid schedule starts and stops", func(t *testing.T) {
		svc, _, _, _, _, _, _ := newAlertService(t, AlertConfig{Enabled: true, Schedule: "@every 1h"})
		assert.NoError(t, svc.Init(context.Background()))
		assert.NoError(t, svc.Close())
	})

	t.Run("invalid schedule returns error", func(t *testing.T) {
		svc, _, _, _, _, _, _ := newAlertService(t, AlertConfig{Enabled: true, Schedule: "bad-schedule"})
		err := svc.Init(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to schedule")
	})
}

func TestAlertService_Run(t *testing.T) {
	t.Run("lock busy returns nil", func(t *testing.T) {
		svc, _, _, _, _, locker, _ := newAlertService(t, AlertConfig{DaysBefore: 7})
		locker.EXPECT().TryLock(mock.Anything, "pat-expiry-alerts").
			Return(nil, db.ErrLockBusy)

		assert.NoError(t, svc.Run(context.Background()))
	})

	t.Run("lock error propagates", func(t *testing.T) {
		svc, _, _, _, _, locker, _ := newAlertService(t, AlertConfig{DaysBefore: 7})
		locker.EXPECT().TryLock(mock.Anything, "pat-expiry-alerts").
			Return(nil, errors.New("connection refused"))

		err := svc.Run(context.Background())
		assert.ErrorContains(t, err, "connection refused")
	})
}

func TestAlertService_sendExpiryReminders(t *testing.T) {
	t.Run("no pending PATs does nothing", func(t *testing.T) {
		svc, repo, _, _, _, _, _ := newAlertService(t, AlertConfig{DaysBefore: 7})
		repo.EXPECT().ListExpiryReminderPending(mock.Anything, 7).
			Return([]models.PAT{}, nil)

		svc.sendExpiryReminders(context.Background())
	})

	t.Run("list error logs and returns", func(t *testing.T) {
		svc, repo, _, _, _, _, _ := newAlertService(t, AlertConfig{DaysBefore: 7})
		repo.EXPECT().ListExpiryReminderPending(mock.Anything, 7).
			Return(nil, errors.New("db error"))

		svc.sendExpiryReminders(context.Background())
	})

	t.Run("sends email marks metadata and creates audit record", func(t *testing.T) {
		svc, repo, userSvc, orgSvc, dialer, _, auditRepo := newAlertService(t, AlertConfig{DaysBefore: 7})

		pat := models.PAT{
			ID:        "pat-1",
			UserID:    "user-1",
			OrgID:     "org-1",
			Title:     "My Token",
			ExpiresAt: time.Now().Add(3 * 24 * time.Hour),
		}

		repo.EXPECT().ListExpiryReminderPending(mock.Anything, 7).
			Return([]models.PAT{pat}, nil)
		userSvc.EXPECT().GetByID(mock.Anything, "user-1").
			Return(user.User{ID: "user-1", Name: "jdoe", Email: "jdoe@example.com"}, nil)
		orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1", Title: "Acme Corp"}, nil)
		dialer.EXPECT().FromHeader().Return("noreply@frontier.io")
		dialer.EXPECT().DialAndSend(mock.AnythingOfType("*mail.Message")).Return(nil)
		repo.EXPECT().SetAlertSentMetadata(mock.Anything, "pat-1", "expiry_reminder_sent_at").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		svc.sendExpiryReminders(context.Background())
	})

	t.Run("email failure logs error and continues", func(t *testing.T) {
		svc, repo, userSvc, orgSvc, dialer, _, _ := newAlertService(t, AlertConfig{DaysBefore: 7})

		pat := models.PAT{
			ID: "pat-fail", UserID: "user-1", OrgID: "org-1",
			Title: "Fail Token", ExpiresAt: time.Now().Add(2 * 24 * time.Hour),
		}

		repo.EXPECT().ListExpiryReminderPending(mock.Anything, 7).
			Return([]models.PAT{pat}, nil)
		userSvc.EXPECT().GetByID(mock.Anything, "user-1").
			Return(user.User{ID: "user-1", Email: "jdoe@example.com"}, nil)
		orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1", Title: "Acme"}, nil)
		dialer.EXPECT().FromHeader().Return("noreply@frontier.io")
		dialer.EXPECT().DialAndSend(mock.AnythingOfType("*mail.Message")).
			Return(errors.New("smtp timeout"))

		svc.sendExpiryReminders(context.Background())
		// No SetAlertSentMetadata or audit Create expected — email failed
	})

	t.Run("uses custom templates from config", func(t *testing.T) {
		cfg := AlertConfig{
			DaysBefore:            7,
			ExpiryReminderSubject: "Custom: {{.Title}}",
			ExpiryReminderBody:    "<b>{{.Title}}</b> expires on {{.ExpiresAt}}",
		}
		svc, repo, userSvc, orgSvc, dialer, _, auditRepo := newAlertService(t, cfg)

		pat := models.PAT{
			ID: "pat-custom", UserID: "user-1", OrgID: "org-1",
			Title: "Custom Token", ExpiresAt: time.Now().Add(5 * 24 * time.Hour),
		}

		repo.EXPECT().ListExpiryReminderPending(mock.Anything, 7).
			Return([]models.PAT{pat}, nil)
		userSvc.EXPECT().GetByID(mock.Anything, "user-1").
			Return(user.User{ID: "user-1", Email: "user@test.com"}, nil)
		orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1", Title: "TestOrg"}, nil)
		dialer.EXPECT().FromHeader().Return("noreply@frontier.io")
		dialer.EXPECT().DialAndSend(mock.AnythingOfType("*mail.Message")).Return(nil)
		repo.EXPECT().SetAlertSentMetadata(mock.Anything, "pat-custom", "expiry_reminder_sent_at").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		svc.sendExpiryReminders(context.Background())
	})
}

func TestAlertService_sendExpiredNotices(t *testing.T) {
	t.Run("no pending PATs does nothing", func(t *testing.T) {
		svc, repo, _, _, _, _, _ := newAlertService(t, AlertConfig{})
		repo.EXPECT().ListExpiredNoticePending(mock.Anything).
			Return([]models.PAT{}, nil)

		svc.sendExpiredNotices(context.Background())
	})

	t.Run("sends email marks metadata and creates audit record", func(t *testing.T) {
		svc, repo, userSvc, orgSvc, dialer, _, auditRepo := newAlertService(t, AlertConfig{})

		pat := models.PAT{
			ID: "pat-expired", UserID: "user-1", OrgID: "org-1",
			Title: "Expired Token", ExpiresAt: time.Now().Add(-2 * time.Hour),
		}

		repo.EXPECT().ListExpiredNoticePending(mock.Anything).
			Return([]models.PAT{pat}, nil)
		userSvc.EXPECT().GetByID(mock.Anything, "user-1").
			Return(user.User{ID: "user-1", Email: "jdoe@example.com"}, nil)
		orgSvc.EXPECT().Get(mock.Anything, "org-1").
			Return(organization.Organization{ID: "org-1", Title: "Acme"}, nil)
		dialer.EXPECT().FromHeader().Return("noreply@frontier.io")
		dialer.EXPECT().DialAndSend(mock.AnythingOfType("*mail.Message")).Return(nil)
		repo.EXPECT().SetAlertSentMetadata(mock.Anything, "pat-expired", "expired_notice_sent_at").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		svc.sendExpiredNotices(context.Background())
	})
}
