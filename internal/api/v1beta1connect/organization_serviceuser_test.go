package v1beta1connect

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/store/postgres"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchOrganizationServiceUsers(t *testing.T) {
	const orgID = "1a02f4ae-9d76-446d-bc9c-a42c46e4f852"

	stringFilter := func(name, op, value string) *frontierv1beta1.RQLFilter {
		return &frontierv1beta1.RQLFilter{
			Name:     name,
			Operator: op,
			Value:    &frontierv1beta1.RQLFilter_StringValue{StringValue: value},
		}
	}

	tests := []struct {
		name string
		// nil means the service must not be reached at all
		setup    func(*mocks.OrgServiceUserService)
		orgSetup func(*mocks.OrganizationService)
		id       string
		query    *frontierv1beta1.RQLRequest
		wantCode connect.Code
		wantMsg  string
		wantLen  int
	}{
		{
			name: "returns the rows the service gives back",
			setup: func(s *mocks.OrgServiceUserService) {
				s.EXPECT().Search(mock.Anything, orgID, mock.Anything).Return(orgserviceuser.OrganizationServiceUsers{
					ServiceUsers: []orgserviceuser.AggregatedServiceUser{
						{ID: "su-1", Title: "first", OrgID: orgID},
						{ID: "su-2", Title: "second", OrgID: orgID, Projects: []orgserviceuser.Project{{ID: "p1", Name: "proj", Title: "Proj"}}},
					},
					Pagination: orgserviceuser.Page{Limit: 50, Offset: 0},
				}, nil)
			},
			query:   &frontierv1beta1.RQLRequest{Limit: 50},
			wantLen: 2,
		},
		{
			// a datetime the proto carries as a string but rql cannot parse. this must
			// be caught before the query is built, not surface as an internal error
			name:     "unparsable created_at is rejected before reaching the service",
			query:    &frontierv1beta1.RQLRequest{Limit: 50, Filters: []*frontierv1beta1.RQLFilter{stringFilter("created_at", "gt", "yesterday-ish")}},
			wantCode: connect.CodeInvalidArgument,
			wantMsg:  "failed to validate rql query",
		},
		{
			name:     "an operator the field does not allow is rejected",
			query:    &frontierv1beta1.RQLRequest{Limit: 50, Filters: []*frontierv1beta1.RQLFilter{stringFilter("created_at", "in", "2026-01-01T00:00:00Z,2026-03-01T00:00:00Z")}},
			wantCode: connect.CodeInvalidArgument,
			wantMsg:  "failed to validate rql query",
		},
		{
			name:     "a field that is not on the aggregate is rejected",
			query:    &frontierv1beta1.RQLRequest{Limit: 50, Filters: []*frontierv1beta1.RQLFilter{stringFilter("nonsense", "eq", "x")}},
			wantCode: connect.CodeInvalidArgument,
			wantMsg:  "failed to read rql query",
		},
		{
			name:     "an unsortable field is rejected",
			query:    &frontierv1beta1.RQLRequest{Limit: 50, Sort: []*frontierv1beta1.RQLSort{{Name: "nonsense", Order: "asc"}}},
			wantCode: connect.CodeInvalidArgument,
			wantMsg:  "failed to validate rql query",
		},
		{
			// the store rejects input the proto layer cannot judge, such as a column
			// the query does not select. that has to stay a bad request
			name: "bad input from the store maps to invalid argument and keeps the reason",
			setup: func(s *mocks.OrgServiceUserService) {
				s.EXPECT().Search(mock.Anything, orgID, mock.Anything).Return(
					orgserviceuser.OrganizationServiceUsers{},
					fmt.Errorf("%w: group_by is not supported", postgres.ErrBadInput))
			},
			query:    &frontierv1beta1.RQLRequest{Limit: 50},
			wantCode: connect.CodeInvalidArgument,
			wantMsg:  "group_by is not supported",
		},
		{
			// the store filters on a uuid column, so a name has to be resolved first
			// rather than reaching postgres as a cast error
			name: "org addressed by name is resolved to its id",
			id:   "compare-org",
			orgSetup: func(o *mocks.OrganizationService) {
				o.EXPECT().GetRaw(mock.Anything, "compare-org").Return(organization.Organization{ID: orgID}, nil)
			},
			setup: func(s *mocks.OrgServiceUserService) {
				s.EXPECT().Search(mock.Anything, orgID, mock.Anything).Return(orgserviceuser.OrganizationServiceUsers{}, nil)
			},
			query:   &frontierv1beta1.RQLRequest{Limit: 50},
			wantLen: 0,
		},
		{
			name: "org that does not exist is not found",
			orgSetup: func(o *mocks.OrganizationService) {
				o.EXPECT().GetRaw(mock.Anything, mock.Anything).Return(organization.Organization{}, organization.ErrNotExist)
			},
			query:    &frontierv1beta1.RQLRequest{Limit: 50},
			wantCode: connect.CodeNotFound,
		},
		{
			name: "any other store failure is internal",
			setup: func(s *mocks.OrgServiceUserService) {
				s.EXPECT().Search(mock.Anything, orgID, mock.Anything).Return(
					orgserviceuser.OrganizationServiceUsers{}, fmt.Errorf("connection refused"))
			},
			query:    &frontierv1beta1.RQLRequest{Limit: 50},
			wantCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewOrgServiceUserService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			orgSvc := mocks.NewOrganizationService(t)
			if tt.orgSetup != nil {
				tt.orgSetup(orgSvc)
			} else if tt.setup != nil {
				orgSvc.EXPECT().GetRaw(mock.Anything, mock.Anything).Return(organization.Organization{ID: orgID}, nil)
			}
			handler := &ConnectHandler{orgServiceUserService: svc, orgService: orgSvc}

			id := tt.id
			if id == "" {
				id = orgID
			}
			resp, err := handler.SearchOrganizationServiceUsers(context.Background(),
				connect.NewRequest(&frontierv1beta1.SearchOrganizationServiceUsersRequest{Id: id, Query: tt.query}))

			if tt.wantMsg == "" && tt.wantCode == 0 {
				assert.NoError(t, err)
				assert.Len(t, resp.Msg.GetOrganizationServiceUsers(), tt.wantLen)
				return
			}

			assert.Error(t, err)
			assert.Equal(t, tt.wantCode, connect.CodeOf(err))
			if tt.wantMsg != "" {
				assert.Contains(t, err.Error(), tt.wantMsg)
			}
		})
	}
}
