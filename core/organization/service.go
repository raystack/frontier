package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/audit"

	"github.com/raystack/frontier/core/preference"

	"github.com/raystack/frontier/core/policy"

	"github.com/raystack/frontier/core/authenticate"

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

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
	authnService    AuthnService
	policyService   PolicyService
	prefService     PreferencesService
}

func NewService(repository Repository, relationService RelationService,
	userService UserService, authnService AuthnService, policyService PolicyService,
	prefService PreferencesService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
		authnService:    authnService,
		policyService:   policyService,
		prefService:     prefService,
	}
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

	// attach user as owner
	if err = s.AddMember(ctx, newOrg.ID, schema.OwnerRelationName, principal); err != nil {
		return newOrg, err
	}

	// attach org to central platform
	if err = s.AttachToPlatform(ctx, newOrg.ID); err != nil {
		return newOrg, err
	}

	return newOrg, nil
}

func (s Service) AddMember(ctx context.Context, orgID, relationName string, principal authenticate.Principal) error {
	roleID := MemberRole
	if relationName == schema.OwnerRelationName {
		roleID = AdminRole
	}
	if _, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        roleID,
		ResourceID:    orgID,
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   principal.ID,
		PrincipalType: principal.Type,
	}); err != nil {
		return err
	}
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        principal.ID,
			Namespace: principal.Type,
		},
		RelationName: relationName,
	}); err != nil {
		return err
	}

	audit.GetAuditor(ctx, orgID).Log(audit.OrgMemberCreatedEvent, audit.Target{
		ID:   principal.ID,
		Type: principal.Type,
	})
	return nil
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

	subjectIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        principal.ID,
			Namespace: principal.Type,
		},
		RelationName: schema.MembershipPermission,
	})
	if err != nil {
		return nil, err
	}
	if len(subjectIDs) == 0 {
		// no organizations
		return []Organization{}, nil
	}

	filter.IDs = subjectIDs
	return s.repository.List(ctx, filter)
}

func (s Service) AddUsers(ctx context.Context, orgID string, userIDs []string) error {
	var err error
	for _, userID := range userIDs {
		if currentErr := s.AddMember(ctx, orgID, schema.MemberRelationName, authenticate.Principal{
			ID:   userID,
			Type: schema.UserPrincipal,
		}); currentErr != nil {
			err = errors.Join(err, currentErr)
		}
	}
	return err
}

// RemoveUsers removes users from an organization as members
// it doesn't remove user access to projects or other resources provided
// by policies, don't call directly, use cascade deleter
func (s Service) RemoveUsers(ctx context.Context, orgID string, userIDs []string) error {
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
		}); err != nil {
			err = errors.Join(err, currentErr)
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
	case errors.Is(err, ErrInvalidUUID):
		return Organization{}, ErrInvalidID
	case errors.Is(err, ErrInvalidID):
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

	// Add user as organization owner
	if err = s.AddMember(ctx, newOrg.ID, schema.OwnerRelationName, authenticate.Principal{
		ID:   usr.ID,
		Type: schema.UserPrincipal,
	}); err != nil {
		return newOrg, fmt.Errorf("failed to add user as owner: %w", err)
	}

	// Attach org to central platform
	if err = s.AttachToPlatform(ctx, newOrg.ID); err != nil {
		return newOrg, fmt.Errorf("failed to attach to platform: %w", err)
	}

	return newOrg, nil
}
