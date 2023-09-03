package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
	authnService    AuthnService
	defaultState    State
}

func NewService(repository Repository, relationService RelationService,
	userService UserService, authnService AuthnService, orgDisableOnCreation bool) *Service {
	defaultState := Enabled
	if orgDisableOnCreation {
		defaultState = Disabled
	}
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
		authnService:    authnService,
		defaultState:    defaultState,
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

func (s Service) Create(ctx context.Context, org Organization) (Organization, error) {
	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("%w: %s", user.ErrNotExist, err.Error())
	}

	newOrg, err := s.repository.Create(ctx, Organization{
		Name:     org.Name,
		Title:    org.Title,
		Avatar:   org.Avatar,
		Metadata: org.Metadata,
		State:    s.defaultState,
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
		return s.ListByUser(ctx, f.UserID)
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

func (s Service) ListByUser(ctx context.Context, userID string) ([]Organization, error) {
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        userID,
			Namespace: schema.UserPrincipal,
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
	return s.repository.GetByIDs(ctx, subjectIDs)
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
func (s Service) RemoveUsers(ctx context.Context, orgID string, userIDs []string) error {
	var err error
	for _, userID := range userIDs {
		if currentErr := s.relationService.Delete(ctx, relation.Relation{
			Object: relation.Object{
				ID:        orgID,
				Namespace: schema.OrganizationNamespace,
			},
			Subject: relation.Subject{
				ID:        userID,
				Namespace: schema.UserPrincipal,
			},
			RelationName: schema.MemberRelationName,
		}); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	return err
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
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
