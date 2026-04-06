package v1beta1connect

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	svc "github.com/raystack/frontier/core/aggregates/orgpats"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_SearchOrganizationPATs(t *testing.T) {
	testOrgID := "9f256f86-31a3-11ec-8d3d-0242ac130003"
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name    string
		setup   func(ps *mocks.OrgPATsService)
		request *connect.Request[frontierv1beta1.SearchOrganizationPATsRequest]
		wantErr error
		wantLen int
	}{
		{
			name: "should return internal error on service failure",
			setup: func(ps *mocks.OrgPATsService) {
				ps.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
					Return(svc.OrganizationPATs{}, fmt.Errorf("db error"))
			},
			request: connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
				OrgId: testOrgID,
			}),
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return invalid argument on bad input error",
			setup: func(ps *mocks.OrgPATsService) {
				ps.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
					Return(svc.OrganizationPATs{}, fmt.Errorf("bad: %w", postgres.ErrBadInput))
			},
			request: connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
				OrgId: testOrgID,
			}),
			wantErr: connect.NewError(connect.CodeInvalidArgument, postgres.ErrBadInput),
		},
		{
			name: "should return empty results",
			setup: func(ps *mocks.OrgPATsService) {
				ps.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
					Return(svc.OrganizationPATs{
						PATs:       []svc.AggregatedPAT{},
						Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 0},
					}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
				OrgId: testOrgID,
			}),
			wantLen: 0,
		},
		{
			name: "should return PATs with scopes and created_by",
			setup: func(ps *mocks.OrgPATsService) {
				ps.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
					Return(svc.OrganizationPATs{
						PATs: []svc.AggregatedPAT{
							{
								ID:    "pat-1",
								Title: "my-token",
								CreatedBy: svc.CreatedBy{
									ID:    "user-1",
									Title: "John Doe",
									Email: "john@test.com",
								},
								Scopes: []patmodels.PATScope{
									{RoleID: "role-1", ResourceType: schema.OrganizationNamespace, ResourceIDs: []string{testOrgID}},
									{RoleID: "role-2", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1"}},
								},
								CreatedAt: now,
								ExpiresAt: now.Add(24 * time.Hour),
							},
						},
						Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 1},
					}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
				OrgId: testOrgID,
			}),
			wantLen: 1,
		},
		{
			name: "should handle PAT with last_used_at",
			setup: func(ps *mocks.OrgPATsService) {
				lastUsed := now.Add(-1 * time.Hour)
				ps.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
					Return(svc.OrganizationPATs{
						PATs: []svc.AggregatedPAT{
							{
								ID:    "pat-2",
								Title: "used-token",
								CreatedBy: svc.CreatedBy{
									ID: "user-1", Title: "John", Email: "john@test.com",
								},
								Scopes:     []patmodels.PATScope{},
								CreatedAt:  now,
								ExpiresAt:  now.Add(24 * time.Hour),
								LastUsedAt: &lastUsed,
							},
						},
						Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 1},
					}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
				OrgId: testOrgID,
			}),
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgPATsSrv := new(mocks.OrgPATsService)

			if tt.setup != nil {
				tt.setup(mockOrgPATsSrv)
			}

			handler := &ConnectHandler{
				orgPATsService: mockOrgPATsSrv,
			}

			resp, err := handler.SearchOrganizationPATs(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, connect.CodeOf(tt.wantErr), connect.CodeOf(err))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Msg.GetOrganizationPats(), tt.wantLen)

				if tt.wantLen > 0 {
					pat := resp.Msg.GetOrganizationPats()[0]
					assert.NotEmpty(t, pat.GetId())
					assert.NotEmpty(t, pat.GetTitle())
					assert.NotNil(t, pat.GetCreatedBy())
					assert.NotEmpty(t, pat.GetCreatedBy().GetId())
					assert.NotNil(t, pat.GetCreatedAt())
					assert.NotNil(t, pat.GetExpiresAt())
				}
			}

			mockOrgPATsSrv.AssertExpectations(t)
		})
	}

	t.Run("should transform scopes correctly", func(t *testing.T) {
		mockOrgPATsSrv := new(mocks.OrgPATsService)

		mockOrgPATsSrv.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
			Return(svc.OrganizationPATs{
				PATs: []svc.AggregatedPAT{
					{
						ID:    "pat-1",
						Title: "scoped-token",
						CreatedBy: svc.CreatedBy{
							ID: "user-1", Title: "John", Email: "john@test.com",
						},
						Scopes: []patmodels.PATScope{
							{RoleID: "role-org", ResourceType: schema.OrganizationNamespace, ResourceIDs: []string{testOrgID}},
							{RoleID: "role-proj", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1", "proj-2"}},
						},
						CreatedAt: now,
						ExpiresAt: now.Add(24 * time.Hour),
					},
				},
				Pagination: utils.Page{Offset: 0, Limit: 30, TotalCount: 1},
			}, nil)

		handler := &ConnectHandler{orgPATsService: mockOrgPATsSrv}
		resp, err := handler.SearchOrganizationPATs(context.Background(), connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
			OrgId: testOrgID,
		}))

		assert.NoError(t, err)
		pat := resp.Msg.GetOrganizationPats()[0]
		assert.Len(t, pat.GetScopes(), 2)

		scopeTypes := make(map[string]bool)
		for _, sc := range pat.GetScopes() {
			scopeTypes[sc.GetResourceType()] = true
			assert.NotEmpty(t, sc.GetRoleId())
		}
		assert.True(t, scopeTypes[schema.OrganizationNamespace])
		assert.True(t, scopeTypes[schema.ProjectNamespace])
	})

	t.Run("should return pagination", func(t *testing.T) {
		mockOrgPATsSrv := new(mocks.OrgPATsService)

		mockOrgPATsSrv.EXPECT().Search(mock.Anything, testOrgID, mock.Anything).
			Return(svc.OrganizationPATs{
				PATs:       []svc.AggregatedPAT{},
				Pagination: utils.Page{Offset: 10, Limit: 30, TotalCount: 100},
			}, nil)

		handler := &ConnectHandler{orgPATsService: mockOrgPATsSrv}
		resp, err := handler.SearchOrganizationPATs(context.Background(), connect.NewRequest(&frontierv1beta1.SearchOrganizationPATsRequest{
			OrgId: testOrgID,
		}))

		assert.NoError(t, err)
		assert.Equal(t, uint32(10), resp.Msg.GetPagination().GetOffset())
		assert.Equal(t, uint32(30), resp.Msg.GetPagination().GetLimit())
		assert.Equal(t, uint32(100), resp.Msg.GetPagination().GetTotalCount())
	})
}
