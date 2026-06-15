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
	ListGroupsByPrincipal(ctx context.Context, principal authenticate.Principal, orgID string) ([]string, error)
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

	// PAT → resolve to underlying user so ownership is on the user, not the token
	subjectID, subjectType := principal.ResolveSubject()
	if err = s.membershipService.OnGroupCreated(ctx, newGroup.ID, newGroup.OrganizationID, subjectID, subjectType); err != nil {
		if cleanupErr := s.repository.Delete(ctx, newGroup.ID); cleanupErr != nil {
			return Group{}, errors.Join(err, fmt.Errorf("rollback group create: %w", cleanupErr))
		}
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
	if flt.Principal != nil {
		if s.membershipService == nil {
			return nil, fmt.Errorf("group: membership service not wired")
		}
		ids, err := s.membershipService.ListGroupsByPrincipal(ctx, *flt.Principal, flt.OrganizationID)
		if err != nil {
			return nil, err
		}
		if len(flt.GroupIDs) > 0 {
			ids = utils.Intersection(ids, flt.GroupIDs)
		}
		if len(ids) == 0 {
			return []Group{}, nil
		}
		flt.GroupIDs = ids
	}

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

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

// Disable is a reversible soft-stop: it flips the group's state only and
// deliberately leaves every SpiceDB relation in place, so Enable restores
// membership and access exactly as it was. Disable is NOT a revocation —
// tearing down the tuples is Delete's job (see core/deleter). Group reads do
// not filter on state today, so authz checks that read SpiceDB directly still
// pass while a group is disabled.
func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

// DeleteModel removes the group entity from storage and emits the audit
// event. All SpiceDB policy and relation cleanup is the caller's
// responsibility — see core/deleter.DeleteGroup for the orchestration that
// pairs this with membership.OnGroupDeleted.
func (s Service) DeleteModel(ctx context.Context, id string) error {
	group, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}

	return audit.NewLogger(ctx, group.OrganizationID).Log(audit.GroupDeletedEvent, audit.GroupTarget(id))
}
