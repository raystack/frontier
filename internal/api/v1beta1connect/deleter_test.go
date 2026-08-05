package v1beta1connect

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func TestHandler_DeleteProject(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.CascadeDeleter)
		request *connect.Request[frontierv1beta1.DeleteProjectRequest]
		want    *connect.Response[frontierv1beta1.DeleteProjectResponse]
		wantErr error
	}{
		{
			name: "should return success if deleted by id return nil error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteProject(mock.Anything, "some-id").Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectRequest{
				Id: "some-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteProjectResponse{}),
			wantErr: nil,
		},
		{
			name: "should return error if deleter service encounters an error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteProject(mock.Anything, "some-id").Return(errors.New("some error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteProject.DeleteProject: project_id=some-id: %w", errors.New("some error"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeleteSrv := new(mocks.CascadeDeleter)
			if tt.setup != nil {
				tt.setup(mockDeleteSrv)
			}
			mockDel := &ConnectHandler{deleterService: mockDeleteSrv}
			resp, err := mockDel.DeleteProject(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

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
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id", false).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}),
			wantErr: nil,
		},
		{
			name: "should pass the token forfeit acknowledgment through to the service",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id", true).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id:                      "some-id",
				AcknowledgeTokenForfeit: true,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}),
			wantErr: nil,
		},
		{
			name: "should return error if deleter service encounters an error",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id", false).Return(errors.New("some_error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteOrganization.DeleteOrganization: organization_id=some-id: %w", errors.New("some_error"))),
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

	t.Run("should return failed precondition with one violation per blocker when the delete is blocked", func(t *testing.T) {
		blocked := &deleter.BlockedError{
			OrgID: "some-id",
			Blockers: []deleter.Blocker{
				{Type: deleter.BlockerActiveSubscription, Subject: "sub-1", Message: "subscription[sub-1] is active: cancel it, then retry the delete"},
				{Type: deleter.BlockerUnusedTokens, Subject: "cust-1", Message: "billing account[cust-1] has 100 unused tokens that deleting the organization forfeits: retry the delete with acknowledge_token_forfeit set to proceed"},
			},
		}
		mockDelOrg := new(mocks.CascadeDeleter)
		mockDelOrg.EXPECT().DeleteOrganization(mock.Anything, "some-id", false).Return(blocked)
		mockDep := &ConnectHandler{deleterService: mockDelOrg}

		resp, err := mockDep.DeleteOrganization(context.Background(), connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
			Id: "some-id",
		}))
		assert.Nil(t, resp)

		var connectErr *connect.Error
		assert.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
		assert.Contains(t, connectErr.Message(), "cancel it")
		assert.Contains(t, connectErr.Message(), "acknowledge_token_forfeit")

		assert.Len(t, connectErr.Details(), 1)
		detail, detailErr := connectErr.Details()[0].Value()
		assert.NoError(t, detailErr)
		failure, ok := detail.(*errdetails.PreconditionFailure)
		assert.True(t, ok)
		assert.Len(t, failure.GetViolations(), 2)
		assert.Equal(t, deleter.BlockerActiveSubscription, failure.GetViolations()[0].GetType())
		assert.Equal(t, "sub-1", failure.GetViolations()[0].GetSubject())
		assert.Equal(t, deleter.BlockerUnusedTokens, failure.GetViolations()[1].GetType())
		assert.Equal(t, "cust-1", failure.GetViolations()[1].GetSubject())
	})
}
