package userpat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"math"
	texttemplate "text/template"
	"time"

	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/userpat/models"
	pkgauditrecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/salt/log"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	mail "gopkg.in/mail.v2"
)

const (
	alertLockKey = "pat-expiry-alerts"

	defaultExpiryReminderBody = `Your personal access token "{{.Title}}" in organization "{{.OrgName}}" will expire on {{.ExpiresAt}}. Regenerate it to avoid service disruption.`

	defaultExpiredNoticeBody = `Your personal access token "{{.Title}}" in organization "{{.OrgName}}" expired on {{.ExpiresAt}}. Any services using this token will no longer authenticate. Please create a new token.`

	defaultExpiryReminderSubject = `Your personal access token "{{.Title}}" expires in {{.DaysLeft}} days`
	defaultExpiredNoticeSubject  = `Your personal access token "{{.Title}}" has expired`

	expiryReminderMetadataKey = "expiry_reminder_sent_at"
	expiredNoticeMetadataKey  = "expired_notice_sent_at"
)

// AlertUserService resolves user details from user ID.
type AlertUserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

// AlertOrgService resolves org details from org ID.
type AlertOrgService interface {
	Get(ctx context.Context, id string) (organization.Organization, error)
}

// AlertRepository is the subset of Repository needed by the alert service.
type AlertRepository interface {
	ListExpiryReminderPending(ctx context.Context, days int) ([]models.PAT, error)
	ListExpiredNoticePending(ctx context.Context) ([]models.PAT, error)
	SetAlertSentMetadata(ctx context.Context, id string, key string) error
}

// AlertAuditRepository creates audit records for alert events.
type AlertAuditRepository interface {
	Create(ctx context.Context, auditRecord auditmodels.AuditRecord) (auditmodels.AuditRecord, error)
}

// Locker acquires distributed locks via Postgres advisory locks.
type Locker interface {
	TryLock(ctx context.Context, id string) (*db.Lock, error)
}

type AlertService struct {
	repo      AlertRepository
	userSvc   AlertUserService
	orgSvc    AlertOrgService
	auditRepo AlertAuditRepository

	dialer mailer.Dialer
	locker Locker
	config AlertConfig
	logger log.Logger
	cron   *cron.Cron
}

func NewAlertService(
	repo AlertRepository,
	userSvc AlertUserService,
	orgSvc AlertOrgService,
	dialer mailer.Dialer,
	locker Locker,
	config AlertConfig,
	logger log.Logger,
	auditRepo AlertAuditRepository,
) *AlertService {
	return &AlertService{
		repo:      repo,
		userSvc:   userSvc,
		orgSvc:    orgSvc,
		auditRepo: auditRepo,
		dialer:    dialer,
		locker:    locker,
		config:    config,
		logger:    logger,
	}
}

func (s *AlertService) Init(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	s.cron = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))
	_, err := s.cron.AddFunc(s.config.Schedule, func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		if err := s.Run(ctx); err != nil {
			s.logger.Error("PAT expiry alert run failed", zap.Error(err))
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule PAT alert job: %w", err)
	}
	s.cron.Start()
	return nil
}

func (s *AlertService) Close() error {
	if s.cron != nil {
		<-s.cron.Stop().Done()
	}
	return nil
}

func (s *AlertService) Run(ctx context.Context) error {
	lock, err := s.locker.TryLock(ctx, alertLockKey)
	if err != nil {
		if errors.Is(err, db.ErrLockBusy) {
			return nil
		}
		return err
	}
	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			s.logger.Error("failed to unlock PAT alert lock", zap.Error(unlockErr))
		}
	}()

	s.logger.Info("running PAT expiry alert check")
	s.sendExpiryReminders(ctx)
	s.sendExpiredNotices(ctx)
	return nil
}

func (s *AlertService) sendExpiryReminders(ctx context.Context) {
	pats, err := s.repo.ListExpiryReminderPending(ctx, s.config.DaysBefore)
	if err != nil {
		s.logger.Error("failed to list pre-expiry PATs", zap.Error(err))
		return
	}

	subjectTpl := s.config.ExpiryReminderSubject
	if subjectTpl == "" {
		subjectTpl = defaultExpiryReminderSubject
	}
	bodyTpl := s.config.ExpiryReminderBody
	if bodyTpl == "" {
		bodyTpl = defaultExpiryReminderBody
	}

	for _, pat := range pats {
		if err := s.sendAlert(ctx, pat, subjectTpl, bodyTpl, expiryReminderMetadataKey, pkgauditrecord.PATExpiryReminderEvent); err != nil {
			s.logger.Error("failed to send expiry reminder",
				zap.String("pat_id", pat.ID), zap.Error(err))
		}
	}
}

func (s *AlertService) sendExpiredNotices(ctx context.Context) {
	pats, err := s.repo.ListExpiredNoticePending(ctx)
	if err != nil {
		s.logger.Error("failed to list post-expiry PATs", zap.Error(err))
		return
	}

	subjectTpl := s.config.ExpiredNoticeSubject
	if subjectTpl == "" {
		subjectTpl = defaultExpiredNoticeSubject
	}
	bodyTpl := s.config.ExpiredNoticeBody
	if bodyTpl == "" {
		bodyTpl = defaultExpiredNoticeBody
	}

	for _, pat := range pats {
		if err := s.sendAlert(ctx, pat, subjectTpl, bodyTpl, expiredNoticeMetadataKey, pkgauditrecord.PATExpiredNoticeEvent); err != nil {
			s.logger.Error("failed to send expired notice",
				zap.String("pat_id", pat.ID), zap.Error(err))
		}
	}
}

type alertTemplateData struct {
	Title     string
	OrgName   string
	ExpiresAt string
	DaysLeft  int
	UserName  string
	UserEmail string
}

func (s *AlertService) sendAlert(ctx context.Context, pat models.PAT, subjectTpl, bodyTpl, metadataKey string, auditEvent pkgauditrecord.Event) error {
	usr, err := s.userSvc.GetByID(ctx, pat.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	org, err := s.orgSvc.Get(ctx, pat.OrgID)
	if err != nil {
		return fmt.Errorf("failed to get org: %w", err)
	}

	daysLeft := max(0, int(math.Ceil(time.Until(pat.ExpiresAt).Hours()/24)))

	data := alertTemplateData{
		Title:     pat.Title,
		OrgName:   org.Title,
		ExpiresAt: pat.ExpiresAt.Format("January 2, 2006 at 3:04 PM UTC"),
		DaysLeft:  daysLeft,
		UserName:  usr.Name,
		UserEmail: usr.Email,
	}

	subject, err := renderTextTemplate(subjectTpl, data)
	if err != nil {
		return fmt.Errorf("failed to render subject: %w", err)
	}

	body, err := renderHTMLTemplate(bodyTpl, data)
	if err != nil {
		return fmt.Errorf("failed to render body: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", s.dialer.FromHeader())
	msg.SetHeader("To", usr.Email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	if err := s.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("sent PAT expiry alert",
		zap.String("pat_id", pat.ID),
		zap.String("pat_title", pat.Title),
		zap.String("user_email", usr.Email),
		zap.String("alert_type", metadataKey))

	if err := s.repo.SetAlertSentMetadata(ctx, pat.ID, metadataKey); err != nil {
		s.logger.Error("alert sent but failed to mark metadata",
			zap.String("pat_id", pat.ID), zap.String("key", metadataKey), zap.Error(err))
	}

	s.createAlertAuditRecord(ctx, pat, usr, org.Title, auditEvent)
	return nil
}

func (s *AlertService) createAlertAuditRecord(ctx context.Context, pat models.PAT, usr user.User, orgName string, event pkgauditrecord.Event) {
	if _, err := s.auditRepo.Create(ctx, auditmodels.AuditRecord{
		Event: event,
		Resource: auditmodels.Resource{
			ID:   pat.OrgID,
			Type: pkgauditrecord.OrganizationType,
			Name: orgName,
		},
		Target: &auditmodels.Target{
			ID:   pat.ID,
			Type: pkgauditrecord.PATType,
			Name: pat.Title,
			Metadata: map[string]any{
				"user_id":    pat.UserID,
				"user_email": usr.Email,
				"expires_at": pat.ExpiresAt.Format(time.RFC3339),
			},
		},
		OrgID:      pat.OrgID,
		OccurredAt: time.Now(),
	}); err != nil {
		s.logger.Error("failed to create audit record for PAT alert",
			zap.String("pat_id", pat.ID), zap.Error(err))
	}
}

func renderTextTemplate(tpl string, data alertTemplateData) (string, error) {
	t, err := texttemplate.New("subject").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderHTMLTemplate(tpl string, data alertTemplateData) (string, error) {
	t, err := htmltemplate.New("body").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
