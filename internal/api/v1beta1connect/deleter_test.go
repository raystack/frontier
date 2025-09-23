package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_DeleteOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.CascadeDeleter)
		request *connect.Request[frontierv1beta1.DeleteOrganizationRequest]
		want    *connect.Response[frontierv1beta1.DeleteOrganizationResponse]
		wantErr error
	}{
		{
			name: "should return success if deleted by id return nil error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id").Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}),
			wantErr: nil,
		},
		{
			name: "should return error if deleter service encounters an error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id").Return(errors.New("some_error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDelOrg := new(mocks.CascadeDeleter)
			if tt.setup != nil {
				tt.setup(mockDelOrg)
			}
			mockDep := &ConnectHandler{deleterService: mockDelOrg}
			resp, err := mockDep.DeleteOrganization(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}