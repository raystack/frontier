package v1beta1connect

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/core/organization"
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
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id").Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}),
			wantErr: nil,
		},
		{
			name: "should return not found when the org is already gone",
			setup: func(as *mocks.CascadeDeleter) {
				as.EXPECT().DeleteOrganization(mock.Anything, "some-id").Return(organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, organization.ErrNotExist),
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
				{Type: deleter.BlockerUnpaidInvoice, Subject: "inv-1", Message: "invoice[inv-1] is unpaid: pay it via its hosted payment page, then retry the delete"},
				{Type: deleter.BlockerNegativeTokenBalance, Subject: "cust-1", Message: "billing account[cust-1] owes 50 tokens: contact support to settle the balance, then retry the delete"},
			},
		}
		mockDelOrg := new(mocks.CascadeDeleter)
		mockDelOrg.EXPECT().DeleteOrganization(mock.Anything, "some-id").Return(blocked)
		mockDep := &ConnectHandler{deleterService: mockDelOrg}

		resp, err := mockDep.DeleteOrganization(context.Background(), connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{
			Id: "some-id",
		}))
		assert.Nil(t, resp)

		var connectErr *connect.Error
		assert.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
		assert.Contains(t, connectErr.Message(), "pay it")
		assert.Contains(t, connectErr.Message(), "contact support")

		assert.Len(t, connectErr.Details(), 1)
		detail, detailErr := connectErr.Details()[0].Value()
		assert.NoError(t, detailErr)
		failure, ok := detail.(*errdetails.PreconditionFailure)
		assert.True(t, ok)
		assert.Len(t, failure.GetViolations(), 2)
		assert.Equal(t, deleter.BlockerUnpaidInvoice, failure.GetViolations()[0].GetType())
		assert.Equal(t, "inv-1", failure.GetViolations()[0].GetSubject())
		assert.Equal(t, deleter.BlockerNegativeTokenBalance, failure.GetViolations()[1].GetType())
		assert.Equal(t, "cust-1", failure.GetViolations()[1].GetSubject())
	})
}

func TestHandler_CheckOrganizationDelete(t *testing.T) {
	t.Run("should report the blockers with can_delete false", func(t *testing.T) {
		mockDel := new(mocks.CascadeDeleter)
		mockDel.EXPECT().CheckOrganizationDelete(mock.Anything, "some-id").Return([]deleter.Blocker{
			{Type: deleter.BlockerActiveSubscription, Subject: "sub-1", SubjectType: deleter.SubjectSubscription, Message: "subscription[sub-1] is active on a paid plan: downgrade to the standard plan, then retry the delete"},
			{Type: deleter.BlockerUnpaidInvoice, Subject: "inv-1", SubjectType: deleter.SubjectInvoice, Message: "invoice[inv-1] is unpaid: pay it via its hosted payment page, then retry the delete"},
		}, nil)
		mockDep := &ConnectHandler{deleterService: mockDel}

		resp, err := mockDep.CheckOrganizationDelete(context.Background(), connect.NewRequest(&frontierv1beta1.CheckOrganizationDeleteRequest{
			Id: "some-id",
		}))
		assert.NoError(t, err)
		assert.False(t, resp.Msg.GetCanDelete())
		assert.Len(t, resp.Msg.GetBlockers(), 2)
		assert.Equal(t, deleter.BlockerActiveSubscription, resp.Msg.GetBlockers()[0].GetType())
		assert.Equal(t, "sub-1", resp.Msg.GetBlockers()[0].GetSubject())
		assert.Equal(t, deleter.SubjectSubscription, resp.Msg.GetBlockers()[0].GetSubjectType())
		assert.Equal(t, deleter.BlockerUnpaidInvoice, resp.Msg.GetBlockers()[1].GetType())
		assert.Equal(t, deleter.SubjectInvoice, resp.Msg.GetBlockers()[1].GetSubjectType())
	})

	t.Run("should report can_delete true when nothing blocks", func(t *testing.T) {
		mockDel := new(mocks.CascadeDeleter)
		mockDel.EXPECT().CheckOrganizationDelete(mock.Anything, "some-id").Return(nil, nil)
		mockDep := &ConnectHandler{deleterService: mockDel}

		resp, err := mockDep.CheckOrganizationDelete(context.Background(), connect.NewRequest(&frontierv1beta1.CheckOrganizationDeleteRequest{
			Id: "some-id",
		}))
		assert.NoError(t, err)
		assert.True(t, resp.Msg.GetCanDelete())
		assert.Empty(t, resp.Msg.GetBlockers())
	})

	t.Run("should return not found for a missing org", func(t *testing.T) {
		mockDel := new(mocks.CascadeDeleter)
		mockDel.EXPECT().CheckOrganizationDelete(mock.Anything, "some-id").Return(nil, organization.ErrNotExist)
		mockDep := &ConnectHandler{deleterService: mockDel}

		resp, err := mockDep.CheckOrganizationDelete(context.Background(), connect.NewRequest(&frontierv1beta1.CheckOrganizationDeleteRequest{
			Id: "some-id",
		}))
		assert.Nil(t, resp)
		assert.Equal(t, connect.NewError(connect.CodeNotFound, organization.ErrNotExist), err)
	})
}
