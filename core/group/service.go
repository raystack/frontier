package group

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	ListRelations(ctx context.Context, rel relation.Relation) ([]relation.Relation, error)
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
	authnService    AuthnService
}

func NewService(repository Repository, relationService RelationService,
	userService UserService, authnService AuthnService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
		authnService:    authnService,
	}
}

func (s Service) Create(ctx context.Context, grp Group) (Group, error) {
	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Group{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	}

	newGroup, err := s.repository.Create(ctx, grp)
	if err != nil {
		return Group{}, err
	}

	// attach group to org
	if err = s.addAsOrgMember(ctx, newGroup); err != nil {
		return Group{}, err
	}

	// attach current user to group as owner
	if err = s.AddMember(ctx, newGroup.ID, schema.OwnerRelationName, principal); err != nil {
		return Group{}, err
	}

	// add relationship between group to org
	if err = s.addOrgToGroup(ctx, newGroup); err != nil {
		return Group{}, err
	}

	return newGroup, nil
}

func (s Service) Get(ctx context.Context, id string) (Group, error) {
	return s.repository.GetByID(ctx, id)
}

func (s Service) List(ctx context.Context, flt Filter) ([]Group, error) {
	return s.repository.List(ctx, flt)
}

func (s Service) Update(ctx context.Context, grp Group) (Group, error) {
	if strings.TrimSpace(grp.ID) != "" {
		return s.repository.UpdateByID(ctx, grp)
	}
	return Group{}, ErrInvalidID
}

func (s Service) ListByUser(ctx context.Context, userId string, flt Filter) ([]Group, error) {
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.UserPrincipal,
			ID:        userId,
		},
		RelationName: schema.MembershipPermission,
	})
	if err != nil {
		return nil, err
	}
	if len(subjectIDs) == 0 {
		// no groups
		return nil, nil
	}
	return s.repository.GetByIDs(ctx, subjectIDs, flt)
}

// AddMember adds a subject(user) to group as member
func (s Service) AddMember(ctx context.Context, groupID, relationName string, principal authenticate.Principal) error {
	rel := relation.Relation{
		Object: relation.Object{
			ID:        groupID,
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			ID:        principal.ID,
			Namespace: principal.Type,
		},
		RelationName: relationName,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

// addOrgToGroup creates an inverse relation that connects group to org
func (s Service) addOrgToGroup(ctx context.Context, team Group) error {
	rel := relation.Relation{
		Object: relation.Object{
			ID:        team.ID,
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			ID:        team.OrganizationID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	}

	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

// addAsOrgMember connects group as a member to org
func (s Service) addAsOrgMember(ctx context.Context, team Group) error {
	rel := relation.Relation{
		Object: relation.Object{
			ID:        team.OrganizationID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:              team.ID,
			Namespace:       schema.GroupNamespace,
			SubRelationName: schema.MemberRelationName,
		},
		RelationName: schema.MemberRelationName,
	}

	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

// ListByOrganization will be useful for nested groups but we don't do that at the moment
// so it will not be directly used
func (s Service) ListByOrganization(ctx context.Context, id string) ([]Group, error) {
	relations, err := s.relationService.ListRelations(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			ID:              id,
			Namespace:       schema.OrganizationNamespace,
			SubRelationName: schema.OrganizationRelationName,
		},
	})
	if err != nil {
		return nil, err
	}

	var groupIDs []string
	for _, rel := range relations {
		groupIDs = append(groupIDs, rel.Object.ID)
	}
	if len(groupIDs) == 0 {
		// no groups
		return []Group{}, nil
	}
	return s.repository.GetByIDs(ctx, groupIDs, Filter{})
}

func (s Service) AddUsers(ctx context.Context, groupID string, userIDs []string) error {
	var err error
	for _, userID := range userIDs {
		currentErr := s.AddMember(ctx, groupID, schema.MemberRelationName, authenticate.Principal{
			ID:   userID,
			Type: schema.UserPrincipal,
		})
		if currentErr != nil {
			err = errors.Join(err, currentErr)
		}
	}
	return err
}

// RemoveUsers removes users from a group as members
func (s Service) RemoveUsers(ctx context.Context, groupID string, userIDs []string) error {
	var err error
	for _, userID := range userIDs {
		if currentErr := s.relationService.Delete(ctx, relation.Relation{
			Object: relation.Object{
				ID:        groupID,
				Namespace: schema.GroupNamespace,
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

func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.GroupPrincipal,
	}}); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}
