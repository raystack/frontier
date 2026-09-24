package v1beta1connect

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/store/postgres"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Every case fails before the handler writes to the stream, so it is never used.
func TestExportOrganizationUsers(t *testing.T) {
	const orgID = "1a02f4ae-9d76-446d-bc9c-a42c46e4f852"

	tests := []struct {
		name      string
		exportErr error
		wantCode  connect.Code
		wantMsg   string
	}{
		{
			name:      "bad input from the store maps to invalid argument and keeps the reason",
			exportErr: fmt.Errorf("%w: value is not a valid uuid", postgres.ErrBadInput),
			wantCode:  connect.CodeInvalidArgument,
			wantMsg:   "value is not a valid uuid",
		},
		{
			name:      "an org with nothing to export maps to invalid argument",
			exportErr: fmt.Errorf("%w: no users found for organization %s", orgusers.ErrNoContent, orgID),
			wantCode:  connect.CodeInvalidArgument,
			wantMsg:   "no data to export",
		},
		{
			name:      "any other store failure is internal",
			exportErr: fmt.Errorf("connection refused"),
			wantCode:  connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewOrgUsersService(t)
			svc.EXPECT().Export(mock.Anything, orgID).Return(nil, "", tt.exportErr)
			handler := &ConnectHandler{orgUsersService: svc}

			err := handler.ExportOrganizationUsers(context.Background(),
				connect.NewRequest(&frontierv1beta1.ExportOrganizationUsersRequest{Id: orgID}), nil)

			assert.Error(t, err)
			assert.Equal(t, tt.wantCode, connect.CodeOf(err))
			if tt.wantMsg != "" {
				assert.Contains(t, err.Error(), tt.wantMsg)
			}
		})
	}
}
