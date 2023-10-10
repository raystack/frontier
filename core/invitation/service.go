package invitation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/pkg/logger"
	"go.uber.org/zap"

	"github.com/mcuadros/go-defaults"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/preference"
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
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

type OrganizationService interface {
	Get(ctx context.Context, id string) (organization.Organization, error)
	AddMember(ctx context.Context, orgID, relationName string, principal authenticate.Principal) error
	ListByUser(ctx context.Context, userID string, f organization.Filter) ([]organization.Organization, error)
}

type GroupService interface {
	Get(ctx context.Context, id string) (group.Group, error)
	AddMember(ctx context.Context, groupID string, principal authenticate.Principal) error
	ListByUser(ctx context.Context, userID string, flt group.Filter) ([]group.Group, error)
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) (policy.Policy, error)
}

type PreferencesService interface {
	LoadPlatformPreferences(ctx context.Context) (map[string]string, error)
}

type Service struct {
	dialer          mailer.Dialer
	repo            Repository
	orgSvc          OrganizationService
	groupSvc        GroupService
	userService     UserService
	relationService RelationService
	policyService   PolicyService
	prefService     PreferencesService
}

func NewService(dialer mailer.Dialer, repo Repository,
	orgSvc OrganizationService, grpSvc GroupService,
	userService UserService, relService RelationService,
	policyService PolicyService, prefService PreferencesService) *Service {
	return &Service{
		dialer:          dialer,
		repo:            repo,
		orgSvc:          orgSvc,
		groupSvc:        grpSvc,
		userService:     userService,
		relationService: relService,
		policyService:   policyService,
		prefService:     prefService,
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
		logger.Ctx(ctx).Error("failed to load platform preferences for invitation", zap.Error(err))
		// don't fail
	}
	c.WithRoles = prefs[preference.PlatformInviteWithRoles] == "true"
	c.MailTemplate.Subject = prefs[preference.PlatformInviteMailSubject]
	c.MailTemplate.Body = prefs[preference.PlatformInviteMailBody]
	return c
}

func (s Service) Create(ctx context.Context, invitation Invitation) (Invitation, error) {
	if invitation.ID == uuid.Nil {
		invitation.ID = uuid.New()
	}
	conf := s.getConfig(ctx)
	if !conf.WithRoles {
		// clear roles if not allowed at instance level
		invitation.RoleIDs = nil
	}

	org, err := s.orgSvc.Get(ctx, invitation.OrgID)
	if err != nil {
		return Invitation{}, fmt.Errorf("invalid organization: %w", err)
	}
	// populate invitation with its uuid just in case it was passed as name
	invitation.OrgID = org.ID

	// check if user is already a member of the organization
	// if yes, we don't invite the user to the same organization again
	_, userOrgMember, err := s.isUserOrgMember(ctx, invitation.OrgID, invitation.UserID)
	if err != nil {
		return invitation, err
	}
	if userOrgMember {
		return invitation, fmt.Errorf("user %s is already a member of organization %s", invitation.UserID, invitation.OrgID)
	}

	invitation, err = s.repo.Set(ctx, invitation)
	if err != nil {
		return Invitation{}, err
	}

	// create relations for authz
	if err = s.createRelations(ctx, invitation.ID, org.ID, invitation.UserID); err != nil {
		return Invitation{}, err
	}

	// notify user
	t, err := template.New("body").Parse(conf.MailTemplate.Body)
	if err != nil {
		return Invitation{}, fmt.Errorf("failed to parse email template: %w", err)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, map[string]string{
		"UserID":         invitation.UserID,
		"Organization":   org.Name,
		"OrganizationID": org.ID,
		"InviteID":       invitation.ID.String(),
	})
	if err != nil {
		return Invitation{}, fmt.Errorf("failed to parse email template: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", s.dialer.FromHeader())
	msg.SetHeader("To", invitation.UserID)
	msg.SetHeader("Subject", conf.MailTemplate.Subject)
	msg.SetBody("text/html", tpl.String())
	if err := s.dialer.DialAndSend(msg); err != nil {
		return invitation, err
	}
	return invitation, nil
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
	err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id.String(),
			Namespace: schema.InvitationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete relation for invitation: %w", err)
	}
	return s.repo.Delete(ctx, id)
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

	orgs, err := s.orgSvc.ListByUser(ctx, userOb.ID, organization.Filter{})
	if err != nil {
		return userOb, false, err
	}
	for _, org := range orgs {
		if org.ID == orgID {
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

	// check if user is already a member of the organization
	// if yes, check if any other part of the invitation applies like group membership
	userOb, userOrgMember, err := s.isUserOrgMember(ctx, invite.OrgID, invite.UserID)
	if err != nil {
		return err
	}
	if !userOrgMember {
		// if not, add user to the organization
		if err = s.orgSvc.AddMember(ctx, invite.OrgID, schema.MemberRelationName, authenticate.Principal{
			ID:   userOb.ID,
			Type: schema.UserPrincipal,
		}); err != nil {
			return err
		}
	}

	// check if the invitation has a group membership
	if len(invite.GroupIDs) > 0 {
		userGroups, err := s.groupSvc.ListByUser(ctx, userOb.ID, group.Filter{})
		if err != nil {
			return err
		}
		for _, groupID := range invite.GroupIDs {
			// check if group id is valid
			grp, err := s.groupSvc.Get(ctx, groupID)
			if err != nil {
				return err
			}

			// check if user is already a member of the group
			// if yes, skip
			alreadyGroupMember := false
			for _, g := range userGroups {
				if g.ID == grp.ID {
					alreadyGroupMember = true
					continue
				}
			}
			if !alreadyGroupMember {
				if err = s.groupSvc.AddMember(ctx, grp.ID, authenticate.Principal{
					ID:   userOb.ID,
					Type: schema.UserPrincipal,
				}); err != nil {
					return err
				}
			}
		}
	}

	// check if invitation has a list of roles which we want to assign to the user at org level
	var roleErr error
	if len(invite.RoleIDs) > 0 {
		conf := s.getConfig(ctx)
		if conf.WithRoles {
			for _, inviteRoleID := range invite.RoleIDs {
				if _, err := s.policyService.Create(ctx, policy.Policy{
					RoleID:        inviteRoleID,
					ResourceID:    invite.OrgID,
					ResourceType:  schema.OrganizationNamespace,
					PrincipalID:   userOb.ID,
					PrincipalType: schema.UserPrincipal,
				}); err != nil {
					roleErr = errors.Join(roleErr, err)
				}
			}
			if roleErr != nil {
				return roleErr
			}
		}
	}

	// delete the invitation
	return s.Delete(ctx, id)
}
