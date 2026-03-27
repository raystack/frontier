package orgpats

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/project"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationPATs, error)
}

type ProjectService interface {
	ListByUser(ctx context.Context, principal authenticate.Principal, flt project.Filter) ([]project.Project, error)
}

type Service struct {
	repository     Repository
	projectService ProjectService
}

func NewService(repository Repository, projectService ProjectService) *Service {
	return &Service{
		repository:     repository,
		projectService: projectService,
	}
}

type CreatedBy struct {
	ID    string
	Title string
	Email string
}

type AggregatedPAT struct {
	ID         string
	Title      string
	CreatedBy  CreatedBy
	Scopes     []patmodels.PATScope
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt *time.Time
	UserID     string
}

// PATSearchFields is used for RQL validation — flat struct with rql tags.
type PATSearchFields struct {
	ID             string     `rql:"name=id,type=string"`
	Title          string     `rql:"name=title,type=string"`
	CreatedByTitle string     `rql:"name=created_by_title,type=string"`
	CreatedByEmail string     `rql:"name=created_by_email,type=string"`
	CreatedAt      time.Time  `rql:"name=created_at,type=datetime"`
	ExpiresAt      time.Time  `rql:"name=expires_at,type=datetime"`
	LastUsedAt     *time.Time `rql:"name=last_used_at,type=datetime"`
}

type OrganizationPATs struct {
	PATs       []AggregatedPAT
	Pagination utils.Page
}

func (s *Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationPATs, error) {
	result, err := s.repository.Search(ctx, orgID, query)
	if err != nil {
		return OrganizationPATs{}, err
	}

	if err := s.resolveAllProjectsScope(ctx, orgID, result.PATs); err != nil {
		return OrganizationPATs{}, err
	}

	return result, nil
}

// resolveAllProjectsScope populates ResourceIDs for all-projects scopes by calling SpiceDB.
// Groups PATs by user_id to minimize SpiceDB calls.
func (s *Service) resolveAllProjectsScope(ctx context.Context, orgID string, pats []AggregatedPAT) error {
	// Collect users that have all-projects scopes
	type allProjectsRef struct {
		patIdx   int
		scopeIdx int
	}
	userRefs := make(map[string][]allProjectsRef)

	for i, pat := range pats {
		for j, sc := range pat.Scopes {
			if sc.ResourceType == schema.ProjectNamespace && len(sc.ResourceIDs) == 0 {
				userRefs[pat.UserID] = append(userRefs[pat.UserID], allProjectsRef{i, j})
			}
		}
	}

	if len(userRefs) == 0 {
		return nil
	}

	// One SpiceDB call per unique user
	for userID, refs := range userRefs {
		principal := authenticate.Principal{
			ID:   userID,
			Type: schema.UserPrincipal,
		}
		projects, err := s.projectService.ListByUser(ctx, principal, project.Filter{OrgID: orgID})
		if err != nil {
			return err
		}

		projectIDs := make([]string, 0, len(projects))
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}

		// Copy per PAT to avoid slice aliasing — multiple PATs from the same user
		// would otherwise share the same underlying array.
		for _, ref := range refs {
			idsCopy := make([]string, len(projectIDs))
			copy(idsCopy, projectIDs)
			pats[ref.patIdx].Scopes[ref.scopeIdx].ResourceIDs = idsCopy
		}
	}

	return nil
}
