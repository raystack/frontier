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
	ListRelations(ctx context.Context, rel relation.Relation) ([]relation.Relation, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type PolicyService interface {
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
	GroupMemberCount(ctx context.Context, ids []string) ([]policy.MemberCount, error)
}

type MembershipService interface {
	OnGroupCreated(ctx context.Context, groupID, orgID, creatorID, creatorType string) error
}

type Service struct {
	repository        Repository
	relationService   RelationService
	authnService      AuthnService
	policyService     PolicyService
	membershipService MembershipService
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

// SetMembershipService sets the membership dependency after construction to break
// the circular init order between group and membership services.
func (s *Service) SetMembershipService(ms MembershipService) {
	s.membershipService = ms
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

	if err = s.membershipService.OnGroupCreated(ctx, newGroup.ID, newGroup.OrganizationID, principal.ID, principal.Type); err != nil {
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

func (s Service) ListByUser(ctx context.Context, principal authenticate.Principal, flt Filter) ([]Group, error) {
	subjectID, subjectType := principal.ResolveSubject()
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object:       relation.Object{Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{Namespace: subjectType, ID: subjectID},
		RelationName: schema.MembershipPermission,
	})
	if err != nil {
		return nil, err
	}
	subjectIDs, err = s.intersectPATScope(ctx, principal, schema.GroupNamespace, subjectIDs)
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

// intersectPATScope narrows resource IDs to only those the PAT is scoped to.
func (s Service) intersectPATScope(ctx context.Context, principal authenticate.Principal,
	namespace string, resourceIDs []string) ([]string, error) {
	if principal.PAT == nil || len(resourceIDs) == 0 {
		return resourceIDs, nil
	}
	patIDs, err := s.relationService.LookupResources(ctx, relation.Relation{
		Object:       relation.Object{Namespace: namespace},
		Subject:      relation.Subject{ID: principal.PAT.ID, Namespace: schema.PATPrincipal},
		RelationName: schema.GetPermission,
	})
	if err != nil {
		return nil, err
	}
	return utils.Intersection(resourceIDs, patIDs), nil
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

// DeleteModel removes the group entity and its Object-side SpiceDB relations.
// Membership cleanup (policies + member/owner relations) is the caller's
// responsibility — see core/deleter.DeleteGroup for the orchestration that
// pairs this with membership.RemoveAllGroupMembers.
func (s Service) DeleteModel(ctx context.Context, id string) error {
	group, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// delete Object-side relations on the group (org<->group hierarchy and
	// any other tuples where this group is the object)
	if err = s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.GroupNamespace,
	}}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
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
