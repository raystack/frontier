package v1beta1connect

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/store/postgres"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchOrganizations(t *testing.T) {
	tests := []struct {
		name      string
		searchErr error
		wantCode  connect.Code
		wantMsg   string
	}{
		{
			name:      "bad input from the store maps to invalid argument and keeps the reason",
			searchErr: fmt.Errorf("%w: value is not a valid uuid", postgres.ErrBadInput),
			wantCode:  connect.CodeInvalidArgument,
			wantMsg:   "value is not a valid uuid",
		},
		{
			name:      "any other store failure is internal",
			searchErr: fmt.Errorf("connection refused"),
			wantCode:  connect.CodeInternal,
		},
		{
			name: "a query the store accepts answers normally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewOrgBillingService(t)
			svc.EXPECT().Search(mock.Anything, mock.Anything).Return(orgbilling.OrgBilling{}, tt.searchErr)
			handler := &ConnectHandler{orgBillingService: svc}

			resp, err := handler.SearchOrganizations(context.Background(),
				connect.NewRequest(&frontierv1beta1.SearchOrganizationsRequest{
					Query: &frontierv1beta1.RQLRequest{Limit: 10},
				}))

			if tt.wantCode == 0 {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
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
