package invitation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/policy"
	"gopkg.in/mail.v2"

	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/frontier/pkg/str"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type Repository interface {
	Set(ctx context.Context, invite Invitation) error
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
	ListByUser(ctx context.Context, userID string) ([]organization.Organization, error)
}

type GroupService interface {
	Get(ctx context.Context, id string) (group.Group, error)
	AddMember(ctx context.Context, groupID, relationName string, principal authenticate.Principal) error
	ListByUser(ctx context.Context, userID string, flt group.Filter) ([]group.Group, error)
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) (policy.Policy, error)
}

type Service struct {
	dialer          mailer.Dialer
	repo            Repository
	orgSvc          OrganizationService
	groupSvc        GroupService
	userService     UserService
	relationService RelationService
	policyService   PolicyService
	config          Config
}

func NewService(dialer mailer.Dialer, repo Repository,
	orgSvc OrganizationService, grpSvc GroupService,
	userService UserService, relService RelationService,
	policyService PolicyService, config Config) *Service {
	return &Service{
		dialer:          dialer,
		repo:            repo,
		orgSvc:          orgSvc,
		groupSvc:        grpSvc,
		userService:     userService,
		relationService: relService,
		policyService:   policyService,
		config:          config,
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

func (s Service) Create(ctx context.Context, invitation Invitation) (Invitation, error) {
	if invitation.ID == uuid.Nil {
		invitation.ID = uuid.New()
	}
	if !s.config.WithRoles {
		// clear roles if not allowed at instance level
		invitation.RoleIDs = nil
	}

	if err := s.repo.Set(ctx, invitation); err != nil {
		return Invitation{}, err
	}

	org, err := s.orgSvc.Get(ctx, invitation.OrgID)
	if err != nil {
		return Invitation{}, fmt.Errorf("invalid organization: %w", err)
	}
	// populate invitation with its uuid just in case it was passed as name
	invitation.OrgID = org.ID

	// create relations for authz
	if err = s.createRelations(ctx, invitation.ID, org.ID, invitation.UserID); err != nil {
		return Invitation{}, err
	}

	// notify user
	t, err := template.New("body").Parse(s.config.MailTemplate.Body)
	if err != nil {
		return Invitation{}, fmt.Errorf("failed to parse email template: %w", err)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, map[string]string{
		"UserID":       invitation.UserID,
		"Organization": org.Name,
	})
	if err != nil {
		return Invitation{}, fmt.Errorf("failed to parse email template: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", s.dialer.FromHeader())
	msg.SetHeader("To", invitation.UserID)
	msg.SetHeader("Subject", s.config.MailTemplate.Subject)
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

// Accept invites a user to an organization
func (s Service) Accept(ctx context.Context, id uuid.UUID) error {
	invite, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	user, err := s.userService.GetByID(ctx, invite.UserID)
	if err != nil {
		return err
	}

	// check if user is already a member of the organization
	// if yes, check if any other part of the invitation applies like group membership
	orgs, err := s.orgSvc.ListByUser(ctx, user.ID)
	if err != nil {
		return err
	}
	userOrgMember := false
	for _, org := range orgs {
		if org.ID == invite.OrgID {
			userOrgMember = true
			break
		}
	}

	// else, add user to the organization
	if !userOrgMember {
		if err = s.orgSvc.AddMember(ctx, invite.OrgID, schema.MemberRelationName, authenticate.Principal{
			ID:   user.ID,
			Type: schema.UserPrincipal,
		}); err != nil {
			return err
		}
	}

	// check if the invitation has a group membership
	if len(invite.GroupIDs) > 0 {
		userGroups, err := s.groupSvc.ListByUser(ctx, user.ID, group.Filter{})
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
				if err = s.groupSvc.AddMember(ctx, grp.ID, schema.MemberRelationName, authenticate.Principal{
					ID:   user.ID,
					Type: schema.UserPrincipal,
				}); err != nil {
					return err
				}
			}
		}
	}

	// check if invitation has a list of roles which we want to assign to the user at org level
	var roleErr error
	if len(invite.RoleIDs) > 0 && s.config.WithRoles {
		for _, inviteRoleID := range invite.RoleIDs {
			if _, err := s.policyService.Create(ctx, policy.Policy{
				RoleID:        inviteRoleID,
				ResourceID:    invite.OrgID,
				ResourceType:  schema.OrganizationNamespace,
				PrincipalID:   user.ID,
				PrincipalType: schema.UserPrincipal,
			}); err != nil {
				roleErr = errors.Join(roleErr, err)
			}
		}
		if roleErr != nil {
			return roleErr
		}
	}

	// delete the invitation
	return s.Delete(ctx, id)
}
