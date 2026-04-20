package organization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/auditrecord"

	"github.com/raystack/frontier/core/preference"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/core/authenticate"

	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/str"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/metrics"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Organization, error)
	GetByIDs(ctx context.Context, ids []string) ([]Organization, error)
	GetByName(ctx context.Context, name string) (Organization, error)
	Create(ctx context.Context, org Organization) (Organization, error)
	List(ctx context.Context, flt Filter) ([]Organization, error)
	UpdateByID(ctx context.Context, org Organization) (Organization, error)
	UpdateByName(ctx context.Context, org Organization) (Organization, error)
	SetState(ctx context.Context, id string, state State) error
	Delete(ctx context.Context, id string) error
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(ctx context.Context, user user.User) (user.User, error)
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
	OrgMemberCount(ctx context.Context, id string) (policy.MemberCount, error)
}

type PreferencesService interface {
	LoadPlatformPreferences(ctx context.Context) (map[string]string, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord auditrecord.AuditRecord) (auditrecord.AuditRecord, error)
}

type RoleService interface {
	Get(ctx context.Context, idOrName string) (role.Role, error)
}

type MembershipService interface {
	AddOrganizationMember(ctx context.Context, orgID, principalID, principalType, roleID string) error
}

type Service struct {
	repository            Repository
	relationService       RelationService
	userService           UserService
	authnService          AuthnService
	policyService         PolicyService
	prefService           PreferencesService
	auditRecordRepository AuditRecordRepository
	roleService           RoleService
	membershipService     MembershipService
}

func NewService(repository Repository, relationService RelationService,
	userService UserService, authnService AuthnService, policyService PolicyService,
	prefService PreferencesService, auditRecordRepository AuditRecordRepository,
	roleService RoleService) *Service {
	return &Service{
		repository:            repository,
		relationService:       relationService,
		userService:           userService,
		authnService:          authnService,
		policyService:         policyService,
		prefService:           prefService,
		auditRecordRepository: auditRecordRepository,
		roleService:           roleService,
	}
}

// SetMembershipService sets the membership dependency after construction to break
// the circular init order between organization and membership services.
func (s *Service) SetMembershipService(ms MembershipService) {
	s.membershipService = ms
}

// extractPrincipalInfo extracts display name and metadata from principal
func extractPrincipalInfo(principal authenticate.Principal) (string, map[string]any) {
	metadata := make(map[string]any)
	var name string

	if principal.PAT != nil {
		name = principal.PAT.Title
		if principal.User != nil {
			metadata["user"] = map[string]any{
				"id":    principal.User.ID,
				"name":  principal.User.Name,
				"title": principal.User.Title,
				"email": principal.User.Email,
			}
		}
	} else if principal.User != nil {
		name = principal.User.Title
		metadata["email"] = principal.User.Email
	} else if principal.ServiceUser != nil {
		name = principal.ServiceUser.Title
	}

	return name, metadata
}

// Get returns an enabled organization by id or name. Will return `org is disabled` error if the organization is disabled
func (s Service) Get(ctx context.Context, idOrName string) (Organization, error) {
	if utils.IsValidUUID(idOrName) {
		orgResp, err := s.repository.GetByID(ctx, idOrName)
		if err != nil {
			return Organization{}, err
		}
		if orgResp.State == Disabled {
			return Organization{}, ErrDisabled
		}
		return orgResp, nil
	}

	orgResp, err := s.repository.GetByName(ctx, idOrName)
	if err != nil {
		return Organization{}, err
	}
	if orgResp.State == Disabled {
		return Organization{}, ErrDisabled
	}
	return orgResp, nil
}

// GetRaw returns an organization(both enabled and disabled) by id or name
func (s Service) GetRaw(ctx context.Context, idOrName string) (Organization, error) {
	if utils.IsValidUUID(idOrName) {
		return s.repository.GetByID(ctx, idOrName)
	}
	return s.repository.GetByName(ctx, idOrName)
}

// GetDefaultOrgStateOnCreate gets from preferences if we need to disable org on create
func (s Service) GetDefaultOrgStateOnCreate(ctx context.Context) (State, error) {
	prefs, err := s.prefService.LoadPlatformPreferences(ctx)
	if err != nil {
		return Enabled, fmt.Errorf("failed to read platform preferences for org creation: %w", err)
	}
	if prefs[preference.PlatformDisableOrgsOnCreate] == "true" {
		return Disabled, nil
	}
	return Enabled, nil
}

func (s Service) Create(ctx context.Context, org Organization) (Organization, error) {
	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("%w: %s", user.ErrNotExist, err.Error())
	}
	if principal.Type != schema.UserPrincipal {
		return Organization{}, ErrUserPrincipalOnly
	}

	defaultState, err := s.GetDefaultOrgStateOnCreate(ctx)
	if err != nil {
		return Organization{}, err
	}

	newOrg, err := s.repository.Create(ctx, Organization{
		Name:     org.Name,
		Title:    org.Title,
		Avatar:   org.Avatar,
		Metadata: org.Metadata,
		State:    defaultState,
	})
	if err != nil {
		return Organization{}, err
	}

	// Use AddOrganizationMember (not SetOrganizationMemberRole) because the user
	// is not yet a member — the org was just created. Set requires existing membership.
	if err = s.membershipService.AddOrganizationMember(ctx, newOrg.ID, principal.ID, principal.Type, AdminRole); err != nil {
		return newOrg, err
	}

	// attach org to central platform
	if err = s.AttachToPlatform(ctx, newOrg.ID); err != nil {
		return newOrg, err
	}

	return newOrg, nil
}

func (s Service) AttachToPlatform(ctx context.Context, orgID string) error {
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		RelationName: schema.PlatformRelationName,
	}); err != nil {
		return err
	}
	return nil
}

func (s Service) List(ctx context.Context, f Filter) ([]Organization, error) {
	if f.UserID != "" {
		return s.ListByUser(ctx, authenticate.Principal{
			ID:   f.UserID,
			Type: schema.UserPrincipal,
		}, f)
	}

	// state gets filtered in db
	return s.repository.List(ctx, f)
}

func (s Service) Update(ctx context.Context, org Organization) (Organization, error) {
	if org.ID != "" {
		return s.repository.UpdateByID(ctx, org)
	}
	return s.repository.UpdateByName(ctx, org)
}

func (s Service) ListByUser(ctx context.Context, principal authenticate.Principal, filter Filter) ([]Organization, error) {
	if metrics.ServiceOprLatency != nil {
		promCollect := metrics.ServiceOprLatency("organization", "ListByUser")
		defer promCollect()
	}

	subjectID, subjectType := principal.ResolveSubject()
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        subjectID,
			Namespace: subjectType,
		},
		RelationName: schema.MembershipPermission,
	})
	if err != nil {
		return nil, err
	}

	if principal.PAT != nil {
		subjectIDs = utils.Intersection(subjectIDs, []string{principal.PAT.OrgID})
	}

	if len(subjectIDs) == 0 {
		// no organizations
		return []Organization{}, nil
	}

	filter.IDs = subjectIDs
	return s.repository.List(ctx, filter)
}

// RemoveUsers removes users from an organization as members
// it doesn't remove user access to projects or other resources provided
// by policies, don't call directly, use cascade deleter
func (s Service) RemoveUsers(ctx context.Context, orgID string, userIDs []string) error {
	// Fetch organization once for all users
	org, orgErr := s.repository.GetByID(ctx, orgID)
	if orgErr != nil {
		return orgErr
	}

	var err error
	for _, userID := range userIDs {
		// remove all access via policies
		userPolicies, currErr := s.policyService.List(ctx, policy.Filter{
			OrgID:       orgID,
			PrincipalID: userID,
		})
		if currErr != nil && !errors.Is(currErr, policy.ErrNotExist) {
			err = errors.Join(err, currErr)
			continue
		}
		for _, pol := range userPolicies {
			if policyErr := s.policyService.Delete(ctx, pol.ID); policyErr != nil {
				err = errors.Join(err, policyErr)
			}
		}

		// remove user from org
		if currentErr := s.relationService.Delete(ctx, relation.Relation{
			Object: relation.Object{
				ID:        orgID,
				Namespace: schema.OrganizationNamespace,
			},
			Subject: relation.Subject{
				ID:        userID,
				Namespace: schema.UserPrincipal,
			},
		}); currentErr != nil {
			err = errors.Join(err, currentErr)
		}

		// Get user details for audit record
		usr, userErr := s.userService.GetByID(ctx, userID)
		var userName string
		var userMetadata map[string]any
		if userErr == nil {
			userName, userMetadata = extractPrincipalInfo(authenticate.Principal{
				ID:   usr.ID,
				Type: schema.UserPrincipal,
				User: &usr,
			})
		}

		// Create audit record for member removal
		_, auditErr := s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
			Event: pkgAuditRecord.OrganizationMemberRemovedEvent,
			Resource: auditrecord.Resource{
				ID:   orgID,
				Type: pkgAuditRecord.OrganizationType,
				Name: org.Title,
			},
			Target: &auditrecord.Target{
				ID:       userID,
				Type:     pkgAuditRecord.UserType,
				Name:     userName,
				Metadata: userMetadata,
			},
			OrgID:      orgID,
			OccurredAt: time.Now(),
		})
		if auditErr != nil {
			err = errors.Join(err, auditErr)
		}

		audit.GetAuditor(ctx, orgID).Log(audit.OrgMemberDeletedEvent, audit.UserTarget(userID))
	}
	return err
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	err := s.repository.SetState(ctx, id, Disabled)
	if err == nil {
		audit.GetAuditor(ctx, id).Log(audit.OrgDisabledEvent, audit.OrgTarget(id))
	}
	return err
}

// DeleteModel doesn't delete the nested resource, only itself
func (s Service) DeleteModel(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.OrganizationNamespace,
	}}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func (s Service) MemberCount(ctx context.Context, orgID string) (int64, error) {
	mc, err := s.policyService.OrgMemberCount(ctx, orgID)
	return int64(mc.Count), err
}

func (s Service) AdminCreate(ctx context.Context, org Organization, ownerEmail string) (Organization, error) {
	// Validate email
	if !user.IsValidEmail(ownerEmail) {
		return Organization{}, user.ErrInvalidEmail
	}

	// Check if organization already exists by name (including disabled orgs)
	_, err := s.GetRaw(ctx, org.Name)
	switch {
	case err == nil:
		return Organization{}, ErrConflict
	case errors.Is(err, ErrNotExist):
		// This is the expected case - proceed with creation
	case errors.Is(err, ErrInvalidUUID), errors.Is(err, ErrInvalidID):
		return Organization{}, ErrInvalidID
	default:
		return Organization{}, err
	}

	// Check if user exists
	var usr user.User
	usr, err = s.userService.GetByEmail(ctx, ownerEmail)
	if err != nil {
		if errors.Is(err, user.ErrNotExist) {
			// User doesn't exist, create it
			usr, err = s.userService.Create(ctx, user.User{
				Email: ownerEmail,
				Name:  str.GenerateUserSlug(ownerEmail),
				State: user.Enabled,
			})
			if err != nil {
				return Organization{}, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return Organization{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	// Get default state for org creation
	defaultState, err := s.GetDefaultOrgStateOnCreate(ctx)
	if err != nil {
		return Organization{}, err
	}

	// Create organization
	newOrg, err := s.repository.Create(ctx, Organization{
		Name:     org.Name,
		Title:    org.Title,
		Avatar:   org.Avatar,
		Metadata: org.Metadata,
		State:    defaultState,
	})
	if err != nil {
		return Organization{}, err
	}

	// Use AddOrganizationMember (not SetOrganizationMemberRole) because the user
	// is not yet a member — the org was just created. Set requires existing membership.
	if err = s.membershipService.AddOrganizationMember(ctx, newOrg.ID, usr.ID, schema.UserPrincipal, AdminRole); err != nil {
		return newOrg, fmt.Errorf("failed to add user as owner: %w", err)
	}

	// Attach org to central platform
	if err = s.AttachToPlatform(ctx, newOrg.ID); err != nil {
		return newOrg, fmt.Errorf("failed to attach to platform: %w", err)
	}

	return newOrg, nil
}
