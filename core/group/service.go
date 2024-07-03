package group

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/core/audit"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/policy"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/relation"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	ListRelations(ctx context.Context, rel relation.Relation) ([]relation.Relation, error)
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
	GroupMemberCount(ctx context.Context, ids []string) ([]policy.MemberCount, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	authnService    AuthnService
	policyService   PolicyService
}

func NewService(repository Repository, relationService RelationService,
	authnService AuthnService, policyService PolicyService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		authnService:    authnService,
		policyService:   policyService,
	}
}

func (s Service) Create(ctx context.Context, grp Group) (Group, error) {
	principal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Group{}, fmt.Errorf("%w: %s", authenticate.ErrInvalidID, err.Error())
	}

	newGroup, err := s.repository.Create(ctx, grp)
	if err != nil {
		return Group{}, err
	}

	// attach group to org
	if err = s.addAsOrgMember(ctx, newGroup); err != nil {
		return Group{}, err
	}
	// add relationship between group to org
	if err = s.addOrgToGroup(ctx, newGroup); err != nil {
		return Group{}, err
	}

	// attach current user to group as owner
	if err = s.addOwner(ctx, newGroup.ID, principal); err != nil {
		return Group{}, err
	}

	return newGroup, nil
}

func (s Service) Get(ctx context.Context, id string) (Group, error) {
	return s.repository.GetByID(ctx, id)
}

func (s Service) GetByIDs(ctx context.Context, ids []string) ([]Group, error) {
	return s.repository.GetByIDs(ctx, ids, Filter{})
}

func (s Service) List(ctx context.Context, flt Filter) ([]Group, error) {
	if flt.OrganizationID == "" && len(flt.GroupIDs) == 0 && !flt.SU {
		return nil, ErrInvalidID
	}

	groups, err := s.repository.List(ctx, flt)
	if err != nil {
		return nil, err
	}
	if flt.WithMemberCount && len(groups) > 0 {
		memberCounts, err := s.policyService.GroupMemberCount(ctx, utils.Map(groups, func(grp Group) string {
			return grp.ID
		}))
		if err != nil {
			return nil, fmt.Errorf("faile to fetch member count: %w", err)
		}
		for i := range groups {
			for _, count := range memberCounts {
				if groups[i].ID == count.ID {
					groups[i].MemberCount = count.Count
				}
			}
		}
	}

	return groups, nil
}

func (s Service) Update(ctx context.Context, grp Group) (Group, error) {
	if strings.TrimSpace(grp.ID) != "" {
		return s.repository.UpdateByID(ctx, grp)
	}
	return Group{}, ErrInvalidID
}

func (s Service) ListByUser(ctx context.Context, principalID, principalType string, flt Filter) ([]Group, error) {
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			Namespace: principalType,
			ID:        principalID,
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
	flt.GroupIDs = subjectIDs
	return s.List(ctx, flt)
}

// AddMember adds a subject(user) to group as member
func (s Service) AddMember(ctx context.Context, groupID string, principal authenticate.Principal) error {
	// first create a policy for the user as member of the group
	if err := s.addMemberPolicy(ctx, groupID, principal); err != nil {
		return err
	}

	// then create a relation between group and user as member
	rel := relation.Relation{
		Object: relation.Object{
			ID:        groupID,
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			ID:        principal.ID,
			Namespace: principal.Type,
		},
		RelationName: schema.MemberRelationName,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

// addOwner adds a user as an owner of group by creating a policy of owner role and an owner relation
func (s Service) addOwner(ctx context.Context, groupID string, principal authenticate.Principal) error {
	pol := policy.Policy{
		RoleID:        schema.GroupOwnerRole,
		ResourceID:    groupID,
		ResourceType:  schema.GroupNamespace,
		PrincipalID:   principal.ID,
		PrincipalType: principal.Type,
	}
	if _, err := s.policyService.Create(ctx, pol); err != nil {
		return err
	}
	// then create a relation between group and user
	rel := relation.Relation{
		Object: relation.Object{
			ID:        groupID,
			Namespace: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			ID:        principal.ID,
			Namespace: principal.Type,
		},
		RelationName: schema.OwnerRelationName,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

// add a policy to user as member of group
func (s Service) addMemberPolicy(ctx context.Context, groupID string, principal authenticate.Principal) error {
	pol := policy.Policy{
		RoleID:        schema.GroupMemberRole,
		ResourceID:    groupID,
		ResourceType:  schema.GroupNamespace,
		PrincipalID:   principal.ID,
		PrincipalType: principal.Type,
	}
	if _, err := s.policyService.Create(ctx, pol); err != nil {
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

	groupIDs := make([]string, 0, len(relations))
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
		if currentErr := s.AddMember(ctx, groupID, authenticate.Principal{
			ID:   userID,
			Type: schema.UserPrincipal,
		}); currentErr != nil {
			err = errors.Join(err, currentErr)
		}
	}
	return err
}

// RemoveUsers removes users from a group as members
func (s Service) RemoveUsers(ctx context.Context, groupID string, userIDs []string) error {
	group, err := s.repository.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	return s.removeUsers(ctx, group, userIDs)
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

func (s Service) Delete(ctx context.Context, id string) error {
	group, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// remove all users from group
	policies, err := s.policyService.List(ctx, policy.Filter{
		GroupID: id,
	})
	if err != nil {
		return err
	}
	userIDs := make([]string, 0)
	for _, pol := range policies {
		if pol.PrincipalType == schema.UserPrincipal || pol.PrincipalType == schema.ServiceUserPrincipal {
			userIDs = append(userIDs, pol.PrincipalID)
		}
	}
	if err := s.removeUsers(ctx, group, userIDs); err != nil {
		return fmt.Errorf("failed to remove users: %w", err)
	}

	// delete pending relations
	if err = s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.GroupPrincipal,
	}}); err != nil {
		return err
	}

	err = s.repository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return audit.NewLogger(ctx, group.OrganizationID).Log(audit.GroupDeletedEvent, audit.GroupTarget(id))
}

func (s Service) removeUsers(ctx context.Context, group Group, userIDs []string) error {
	var err error

	for _, userID := range userIDs {
		// remove all access via policies
		userPolicies, currentErr := s.policyService.List(ctx, policy.Filter{
			GroupID:     group.ID,
			PrincipalID: userID,
		})
		if currentErr != nil && !errors.Is(currentErr, policy.ErrNotExist) {
			err = errors.Join(err, currentErr)
			continue
		}
		for _, pol := range userPolicies {
			if policyErr := s.policyService.Delete(ctx, pol.ID); policyErr != nil {
				err = errors.Join(err, policyErr)
			}
		}

		// remove all relations
		if currentErr := s.relationService.Delete(ctx, relation.Relation{
			Object: relation.Object{
				ID:        group.ID,
				Namespace: schema.GroupNamespace,
			},
			Subject: relation.Subject{
				ID: userID,
			},
		}); currentErr != nil {
			err = errors.Join(err, currentErr)
		}

		if currentErr == nil {
			audit.GetAuditor(ctx, group.OrganizationID).LogWithAttrs(audit.GroupMemberRemovedEvent, audit.GroupTarget(group.ID), map[string]string{
				"userID": userID,
			})
		}
	}

	return err
}
