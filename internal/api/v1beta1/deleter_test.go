package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_DeleteProject(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.CascadeDeleter)
		req     *frontierv1beta1.DeleteProjectRequest
		want    *frontierv1beta1.DeleteProjectResponse
		wantErr error
	}{
		{
			name: "should return success if deleted by id return nil error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteProject(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(nil)
			},
			req: &frontierv1beta1.DeleteProjectRequest{
				Id: "some-id",
			},
			want:    &frontierv1beta1.DeleteProjectResponse{},
			wantErr: nil,
		},
		{
			name: "should return error if deleter service encounters an error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteProject(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(errors.New("some error"))
			},
			req: &frontierv1beta1.DeleteProjectRequest{
				Id: "some-id",
			},
			want:    nil,
			wantErr: status.Errorf(codes.Internal, errors.New("some error").Error()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeleteSrv := new(mocks.CascadeDeleter)
			if tt.setup != nil {
				tt.setup(mockDeleteSrv)
			}
			mockDel := Handler{deleterService: mockDeleteSrv}
			resp, err := mockDel.DeleteProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})

	}
}

func TestHandler_DeleteOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.CascadeDeleter)
		req     *frontierv1beta1.DeleteOrganizationRequest
		want    *frontierv1beta1.DeleteOrganizationResponse
		wantErr error
	}{
		{
			name: "should return success if deleted by id return nil error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(nil)
			},
			req: &frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			},
			want:    &frontierv1beta1.DeleteOrganizationResponse{},
			wantErr: nil,
		},
		{
			name: "should return error if deleter service encounters an error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(errors.New("some_error"))
			},
			req: &frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			},
			want:    nil,
			wantErr: status.Errorf(codes.Internal, errors.New("some_error").Error()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDelOrg := new(mocks.CascadeDeleter)
			if tt.setup != nil {
				tt.setup(mockDelOrg)
			}
			mockDel := Handler{deleterService: mockDelOrg}
			resp, err := mockDel.DeleteOrganization(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}

}
