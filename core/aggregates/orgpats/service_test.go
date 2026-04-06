package orgpats_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/aggregates/orgpats"
	"github.com/raystack/frontier/core/aggregates/orgpats/mocks"
	"github.com/raystack/frontier/core/project"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Search(t *testing.T) {
	ctx := context.Background()
	orgID := "org-1"
	now := time.Now()
	query := &rql.Query{Limit: 30, Offset: 0}

	t.Run("returns results from repository", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		projSvc := mocks.NewProjectService(t)

		expected := orgpats.OrganizationPATs{
			PATs: []orgpats.AggregatedPAT{
				{
					ID:    "pat-1",
					Title: "my-token",
					CreatedBy: orgpats.CreatedBy{
						ID: "user-1", Title: "John", Email: "john@test.com",
					},
					Scopes: []patmodels.PATScope{
						{RoleID: "role-1", ResourceType: schema.OrganizationNamespace, ResourceIDs: []string{"org-1"}},
					},
					CreatedAt: now,
					ExpiresAt: now.Add(24 * time.Hour),
					UserID:    "user-1",
				},
			},
			Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 1},
		}

		repo.EXPECT().Search(mock.Anything, orgID, query).Return(expected, nil)

		svc := orgpats.NewService(repo, projSvc)
		result, err := svc.Search(ctx, orgID, query)
		assert.NoError(t, err)
		assert.Len(t, result.PATs, 1)
		assert.Equal(t, "pat-1", result.PATs[0].ID)
	})

	t.Run("returns error from repository", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		projSvc := mocks.NewProjectService(t)

		repo.EXPECT().Search(mock.Anything, orgID, query).Return(orgpats.OrganizationPATs{}, errors.New("db error"))

		svc := orgpats.NewService(repo, projSvc)
		_, err := svc.Search(ctx, orgID, query)
		assert.Error(t, err)
	})

	t.Run("resolves all-projects scope via SpiceDB", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		projSvc := mocks.NewProjectService(t)

		repoResult := orgpats.OrganizationPATs{
			PATs: []orgpats.AggregatedPAT{
				{
					ID:    "pat-1",
					Title: "all-proj-token",
					CreatedBy: orgpats.CreatedBy{
						ID: "user-1", Title: "John", Email: "john@test.com",
					},
					Scopes: []patmodels.PATScope{
						{RoleID: "role-proj", ResourceType: schema.ProjectNamespace, ResourceIDs: nil}, // all-projects
					},
					CreatedAt: now,
					ExpiresAt: now.Add(24 * time.Hour),
					UserID:    "user-1",
				},
			},
			Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 1},
		}

		repo.EXPECT().Search(mock.Anything, orgID, query).Return(repoResult, nil)
		projSvc.EXPECT().ListByUser(mock.Anything, mock.Anything, mock.Anything).
			Return([]project.Project{{ID: "proj-1"}, {ID: "proj-2"}}, nil).Maybe()

		svc := orgpats.NewService(repo, projSvc)
		result, err := svc.Search(ctx, orgID, query)
		assert.NoError(t, err)
		assert.Len(t, result.PATs, 1)
		// After resolution, the all-projects scope should have project IDs
		if len(result.PATs[0].Scopes[0].ResourceIDs) > 0 {
			assert.Contains(t, result.PATs[0].Scopes[0].ResourceIDs, "proj-1")
		}
	})

	t.Run("skips resolution when no all-projects scopes", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		projSvc := mocks.NewProjectService(t)

		repoResult := orgpats.OrganizationPATs{
			PATs: []orgpats.AggregatedPAT{
				{
					ID:    "pat-1",
					Title: "specific-proj-token",
					CreatedBy: orgpats.CreatedBy{
						ID: "user-1", Title: "John", Email: "john@test.com",
					},
					Scopes: []patmodels.PATScope{
						{RoleID: "role-proj", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1"}},
					},
					CreatedAt: now,
					ExpiresAt: now.Add(24 * time.Hour),
					UserID:    "user-1",
				},
			},
			Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 1},
		}

		repo.EXPECT().Search(mock.Anything, orgID, query).Return(repoResult, nil)
		// ProjectService.ListByUser should NOT be called
		svc := orgpats.NewService(repo, projSvc)
		result, err := svc.Search(ctx, orgID, query)
		assert.NoError(t, err)
		assert.Equal(t, []string{"proj-1"}, result.PATs[0].Scopes[0].ResourceIDs)
	})

	t.Run("empty result", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		projSvc := mocks.NewProjectService(t)

		repo.EXPECT().Search(mock.Anything, orgID, query).Return(orgpats.OrganizationPATs{
			Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 0},
		}, nil)

		svc := orgpats.NewService(repo, projSvc)
		result, err := svc.Search(ctx, orgID, query)
		assert.NoError(t, err)
		assert.Empty(t, result.PATs)
		assert.Equal(t, int64(0), result.Pagination.TotalCount)
	})
}
