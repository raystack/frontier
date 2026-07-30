package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/relation"
)

type RelationService interface {
	Delete(ctx context.Context, rel relation.Relation) error
}

type PolicyService interface {
	ProjectMemberCount(ctx context.Context, ids []string) ([]policy.MemberCount, error)
}

type AuthnService interface {
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
}

type MembershipService interface {
	ListProjectsByPrincipal(ctx context.Context, principal authenticate.Principal, orgID string, nonInherited bool) ([]string, error)
	OnProjectCreated(ctx context.Context, projectID, orgID, creatorID, creatorType string) error
}

type Service struct {
	repository        Repository
	relationService   RelationService
	policyService     PolicyService
	authnService      AuthnService
	membershipService MembershipService
}

func NewService(repository Repository, relationService RelationService,
	policyService PolicyService, authnService AuthnService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		policyService:   policyService,
		authnService:    authnService,
	}
}

// SetMembershipService sets the membership dependency after construction to
// break the circular init order between project and membership services.
func (s *Service) SetMembershipService(ms MembershipService) {
	s.membershipService = ms
}

func (s Service) Get(ctx context.Context, idOrName string) (Project, error) {
	if utils.IsValidUUID(idOrName) {
		return s.repository.GetByID(ctx, idOrName)
	}
	return s.repository.GetByName(ctx, idOrName)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	if s.membershipService == nil {
		return Project{}, fmt.Errorf("project: membership service is not set")
	}

	currentPrincipal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Project{}, err
	}

	newProject, err := s.repository.Create(ctx, prj)
	if err != nil {
		return Project{}, err
	}

	// PAT → resolve to underlying user so ownership is on the user, not the token
	subjectID, subjectType := currentPrincipal.ResolveSubject()
	if err = s.membershipService.OnProjectCreated(ctx, newProject.ID, prj.Organization.ID, subjectID, subjectType); err != nil {
		if cleanupErr := s.repository.Delete(ctx, newProject.ID); cleanupErr != nil {
			return Project{}, errors.Join(err, fmt.Errorf("rollback project create: %w", cleanupErr))
		}
		return Project{}, err
	}
	return newProject, nil
}

func (s Service) List(ctx context.Context, f Filter) ([]Project, error) {
	if f.Principal != nil {
		if !utils.IsValidUUID(f.Principal.ID) {
			return nil, ErrInvalidUUID
		}
		if f.Principal.Type == "" {
			return nil, ErrInvalidPrincipalType
		}
		if s.membershipService == nil {
			return nil, fmt.Errorf("project: membership service is not set")
		}
		ids, err := s.membershipService.ListProjectsByPrincipal(ctx, *f.Principal, f.OrgID, f.NonInherited)
		if err != nil {
			return nil, err
		}
		if len(f.ProjectIDs) > 0 {
			ids = utils.Intersection(ids, f.ProjectIDs)
		}
		if len(ids) == 0 {
			return []Project{}, nil
		}
		f.ProjectIDs = ids
	}

	projects, err := s.repository.List(ctx, f)
	if err != nil {
		return nil, err
	}

	if f.WithMemberCount && len(projects) > 0 {
		// get member count for each project
		projectIDs := utils.Map(projects, func(p Project) string {
			return p.ID
		})
		memberCounts, err := s.policyService.ProjectMemberCount(ctx, projectIDs)
		if err != nil {
			return nil, err
		}
		for i := range projects {
			for _, count := range memberCounts {
				if projects[i].ID == count.ID {
					projects[i].MemberCount = count.Count
				}
			}
		}
	}

	return projects, nil
}

func (s Service) Update(ctx context.Context, prj Project) (Project, error) {
	if utils.IsValidUUID(prj.ID) {
		return s.repository.UpdateByID(ctx, prj)
	}
	return s.repository.UpdateByName(ctx, prj)
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

// Disable is a reversible soft-stop: it flips the project's state only and
// deliberately leaves every SpiceDB relation in place, so Enable restores
// access exactly as it was. Disable is NOT a revocation — tearing down the
// tuples is Delete's job (see core/deleter). Authz checks that read SpiceDB
// directly still pass while a project is disabled, by design.
func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

// DeleteModel doesn't delete the nested resource, only itself
func (s Service) DeleteModel(ctx context.Context, id string) error {
	// delete all relations where resource is an object
	// all relations where project is an subject should already been deleted by now
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.ProjectNamespace,
	}}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		return err
	}
	return s.repository.Delete(ctx, id)
}
