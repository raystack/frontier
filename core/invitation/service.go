package invitation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/raystack/frontier/core/organization"

	"github.com/mcuadros/go-defaults"
	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/preference"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/robfig/cron/v3"
	"gopkg.in/mail.v2"

	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/frontier/pkg/str"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type Repository interface {
	Set(ctx context.Context, invite Invitation) (Invitation, error)
	List(ctx context.Context, flt Filter) ([]Invitation, error)
	ListByUser(ctx context.Context, id string) ([]Invitation, error)
	Get(ctx context.Context, id uuid.UUID) (Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListExpired(ctx context.Context) ([]Invitation, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

type OrganizationService interface {
	Get(ctx context.Context, id string) (organization.Organization, error)
}

type MembershipService interface {
	AddOrganizationMember(ctx context.Context, orgID, principalID, principalType, roleID string) error
	SetGroupMemberRole(ctx context.Context, groupID, principalID, principalType, roleID string) error
	ListResourcesByPrincipal(ctx context.Context, principal authenticate.Principal, resourceType string, filter membership.ResourceFilter) ([]string, error)
}

type GroupService interface {
	Get(ctx context.Context, id string) (group.Group, error)
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type PreferencesService interface {
	LoadPlatformPreferences(ctx context.Context) (map[string]string, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	dialer                mailer.Dialer
	repo                  Repository
	orgSvc                OrganizationService
	groupSvc              GroupService
	userService           UserService
	relationService       RelationService
	prefService           PreferencesService
	auditRecordRepository AuditRecordRepository
	membershipSvc         MembershipService
	cron                  *cron.Cron
}

// invitationCleanupSchedule runs the expired-invitation cleanup once a day at
// midnight (UTC).
const invitationCleanupSchedule = "0 0 * * *"

func NewService(dialer mailer.Dialer, repo Repository,
	orgSvc OrganizationService, grpSvc GroupService,
	userService UserService, relService RelationService,
	prefService PreferencesService,
	auditRecordRepository AuditRecordRepository,
	membershipSvc MembershipService) *Service {
	return &Service{
		dialer:                dialer,
		repo:                  repo,
		orgSvc:                orgSvc,
		groupSvc:              grpSvc,
		userService:           userService,
		relationService:       relService,
		prefService:           prefService,
		auditRecordRepository: auditRecordRepository,
		membershipSvc:         membershipSvc,
		// Recover so a panic inside the cleanup job can't take down the scheduler
		// (or the server) — the job just logs and the next tick runs as usual.
		cron: cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
	}
}

func (s Service) List(ctx context.Context, flt Filter) ([]Invitation, error) {
	return s.repo.List(ctx, flt)
}

func (s Service) ListByUser(ctx context.Context, id string) ([]Invitation, error) {
	return s.repo.ListByUser(ctx, id)
}

func (s Service) Get(ctx context.Context, id uuid.UUID) (Invitation, error) {
	return s.repo.Get(ctx, id)
}

func (s Service) getConfig(ctx context.Context) *Config {
	c := &Config{}
	defaults.SetDefaults(c)
	prefs, err := s.prefService.LoadPlatformPreferences(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load platform preferences for invitation", "error", err)
		// don't fail
	}
	c.WithRoles = strings.EqualFold(prefs[preference.PlatformInviteWithRoles], "true")
	c.MailTemplate.Subject = prefs[preference.PlatformInviteMailSubject]
	c.MailTemplate.Body = prefs[preference.PlatformInviteMailBody]
	return c
}

func (s Service) Create(ctx context.Context, inviteToCreate Invitation) (Invitation, error) {
	if inviteToCreate.ID == uuid.Nil {
		inviteToCreate.ID = uuid.New()
	}
	conf := s.getConfig(ctx)
	if !conf.WithRoles {
		// clear roles if not allowed at instance level
		inviteToCreate.RoleIDs = nil
	}

	org, err := s.orgSvc.Get(ctx, inviteToCreate.OrgID)
	if err != nil {
		return Invitation{}, fmt.Errorf("invalid organization: %w", err)
	}
	// populate inviteToCreate with its uuid just in case it was passed as name
	inviteToCreate.OrgID = org.ID

	// check if user is already a member of the organization
	// if yes, we don't invite the user to the same organization again
	_, userOrgMember, err := s.isUserOrgMember(ctx, inviteToCreate.OrgID, inviteToCreate.UserEmailID)
	if err != nil {
		return Invitation{}, err
	}
	if userOrgMember {
		return Invitation{}, fmt.Errorf("%w: user: %s, organization: %s", ErrAlreadyMember, inviteToCreate.UserEmailID, inviteToCreate.OrgID)
	}

	// before creating a new invite check if user has already an active invite
	invites, err := s.repo.List(ctx, Filter{
		OrgID:  inviteToCreate.OrgID,
		UserID: inviteToCreate.UserEmailID,
	})
	if err != nil {
		return Invitation{}, err
	}
	for _, inv := range invites {
		if inv.ExpiresAt.After(time.Now()) {
			// if invite is not expired, return the existing invite
			inviteToCreate.ID = inv.ID
			break
		}
	}

	// update or create the invitation
	createdInvitation, err := s.repo.Set(ctx, inviteToCreate)
	if err != nil {
		return Invitation{}, err
	}
	// create relations for authz
	if err = s.createRelations(ctx, createdInvitation.ID, org.ID, createdInvitation.UserEmailID); err != nil {
		return Invitation{}, err
	}

	// notify user
	t, err := template.New("body").Parse(conf.MailTemplate.Body)
	if err != nil {
		return Invitation{}, fmt.Errorf("failed to parse email template: %w", err)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, map[string]string{
		"UserID":         createdInvitation.UserEmailID,
		"Organization":   getOrgName(org),
		"OrganizationID": org.ID,
		"InviteID":       createdInvitation.ID.String(),
	})
	if err != nil {
		return Invitation{}, fmt.Errorf("failed to parse email template: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", s.dialer.FromHeader())
	msg.SetHeader("To", createdInvitation.UserEmailID)
	msg.SetHeader("Subject", conf.MailTemplate.Subject)
	msg.SetBody("text/html", tpl.String())
	if err := s.dialer.DialAndSend(msg); err != nil {
		return Invitation{}, err
	}
	return createdInvitation, nil
}

func getOrgName(org organization.Organization) string {
	if org.Title != "" {
		return org.Title
	}
	return org.Name
}

func (s Service) createRelations(ctx context.Context, invitationID uuid.UUID, orgID, userEmail string) error {
	_, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        invitationID.String(),
			Namespace: schema.InvitationNamespace,
		},
		Subject: relation.Subject{
			ID:        str.GenerateUserSlug(userEmail),
			Namespace: schema.UserPrincipal,
		},
		RelationName: schema.UserRelationName,
	})
	if err != nil {
		return fmt.Errorf("failed to create relation for invitation: %w", err)
	}
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        invitationID.String(),
			Namespace: schema.InvitationNamespace,
		},
		Subject: relation.Subject{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	})
	if err != nil {
		return fmt.Errorf("failed to create relation for invitation: %w", err)
	}
	return nil
}

func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	// Remove every relation anchored on the invitation object, not just the org
	// one. createRelations writes both a user (app/invitation:<id>#user) and an
	// org (app/invitation:<id>#org) tuple; filtering by org alone leaked the user
	// tuple on every accept/expire/delete. Omitting RelationName matches both.
	err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id.String(),
			Namespace: schema.InvitationNamespace,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete relation for invitation: %w", err)
	}
	return s.repo.Delete(ctx, id)
}

// InitInvitationCleanup starts a background job that deletes expired invitations
// on a daily schedule. Nothing else removes an invite once it expires (Accept
// only rejects it, Create just writes a new one over it), so without this the
// expired invites keep their row and both SpiceDB tuples forever.
func (s Service) InitInvitationCleanup(ctx context.Context) error {
	_, err := s.cron.AddFunc(invitationCleanupSchedule, func() {
		if err := s.DeleteExpiredInvitations(ctx); err != nil {
			slog.WarnContext(ctx, "failed to clean up expired invitations", "err", err)
		}
	})
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

// DeleteExpiredInvitations removes every invitation past its expiry. It goes
// through Delete (not a raw row purge) so each invite's SpiceDB tuples AND its
// invitations row are removed together — a row-only delete would leak the
// #user / #org tuples behind.
func (s Service) DeleteExpiredInvitations(ctx context.Context) error {
	expired, err := s.repo.ListExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to list expired invitations: %w", err)
	}
	deleted := 0
	for _, inv := range expired {
		if err := s.Delete(ctx, inv.ID); err != nil {
			slog.WarnContext(ctx, "failed to delete expired invitation", "invitation_id", inv.ID.String(), "err", err)
			continue
		}
		deleted++
	}
	if deleted > 0 {
		slog.DebugContext(ctx, "deleted expired invitations", "count", deleted)
	}
	return nil
}

// Close stops the background cleanup job.
func (s Service) Close() error {
	return s.cron.Stop().Err()
}

// check if user is already part of the organization that the invitation is created for
func (s Service) isUserOrgMember(ctx context.Context, orgID, userID string) (user.User, bool, error) {
	userOb, err := s.userService.GetByID(ctx, userID)
	if errors.Is(err, user.ErrNotExist) {
		return userOb, false, nil
	}
	if err != nil {
		return userOb, false, err
	}

	orgIDs, err := s.membershipSvc.ListResourcesByPrincipal(ctx, authenticate.Principal{
		ID:   userOb.ID,
		Type: schema.UserPrincipal,
	}, schema.OrganizationNamespace, membership.ResourceFilter{})
	if err != nil {
		return userOb, false, err
	}
	for _, id := range orgIDs {
		if id == orgID {
			return userOb, true, nil
		}
	}
	return userOb, false, nil
}

// Accept invites a user to an organization
func (s Service) Accept(ctx context.Context, id uuid.UUID) error {
	invite, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if invite.ExpiresAt.Before(time.Now()) {
		return ErrInviteExpired
	}

	userOb, userOrgMember, err := s.isUserOrgMember(ctx, invite.OrgID, invite.UserEmailID)
	if err != nil {
		return err
	}

	// Use the first role ID from the invitation, fall back to viewer.
	orgRoleID := schema.RoleOrganizationViewer
	conf := s.getConfig(ctx)
	if conf.WithRoles && len(invite.RoleIDs) > 0 {
		orgRoleID = invite.RoleIDs[0]
	}

	if !userOrgMember {
		// User is not yet a member — add with the invitation's role.
		// ErrAlreadyMember is possible in a race (user added between invite creation
		// and acceptance) — treat as success since the user is already in the org.
		err = s.membershipSvc.AddOrganizationMember(ctx, invite.OrgID, userOb.ID, schema.UserPrincipal, orgRoleID)
		if err != nil && !errors.Is(err, membership.ErrAlreadyMember) {
			return err
		}
	}

	// check if the invitation has a group membership
	if len(invite.GroupIDs) > 0 {
		principal := authenticate.Principal{ID: userOb.ID, Type: schema.UserPrincipal}
		userGroups, err := s.groupSvc.List(ctx, group.Filter{Principal: &principal})
		if err != nil {
			return err
		}
		for _, groupID := range invite.GroupIDs {
			grp, err := s.groupSvc.Get(ctx, groupID)
			if err != nil {
				return err
			}

			alreadyGroupMember := false
			for _, g := range userGroups {
				if g.ID == grp.ID {
					alreadyGroupMember = true
					continue
				}
			}
			if !alreadyGroupMember {
				if err = s.membershipSvc.SetGroupMemberRole(ctx, grp.ID, userOb.ID, schema.UserPrincipal, schema.GroupMemberRole); err != nil {
					return err
				}
			}
		}
	}

	// fetch organization details for audit record
	org, err := s.orgSvc.Get(ctx, invite.OrgID)
	if err != nil {
		return err
	}

	// create audit record for invitation acceptance
	_, err = s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: pkgAuditRecord.OrganizationInvitationAcceptedEvent,
		Resource: models.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &models.Target{
			ID:   userOb.ID,
			Type: pkgAuditRecord.UserType,
			Name: userOb.Title,
			Metadata: map[string]any{
				"email": userOb.Email,
			},
		},
		OrgID:      invite.OrgID,
		OccurredAt: time.Now(),
	})
	if err != nil {
		return err
	}

	// delete the invitation
	return s.Delete(ctx, id)
}
