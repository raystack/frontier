package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/pkg/uuid"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	testPolicyID  = uuid.NewString()
	testPolicyMap = map[string]policy.Policy{
		testPolicyID: {
			ID:          testPolicyID,
			UserID:      testUserID,
			ResourceID:  testResourceID,
			NamespaceID: "policy-1",
			RoleID:      "reader",
		},
	}
)

func TestListPolicies(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.PolicyService)
		req   *shieldv1beta1.ListPoliciesRequest
		want  *shieldv1beta1.ListPoliciesResponse
		err   error
	}{
		{
			title: "should return internal error if policy service return some error",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().List(mock.Anything, policy.Filter{}).Return([]policy.Policy{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "should return success if policy service return nil error",
			setup: func(ps *mocks.PolicyService) {
				var testPoliciesList []policy.Policy
				for _, p := range testPolicyMap {
					testPoliciesList = append(testPoliciesList, p)
				}
				ps.EXPECT().List(mock.Anything, policy.Filter{}).Return(testPoliciesList, nil)
			},
			want: &shieldv1beta1.ListPoliciesResponse{Policies: []*shieldv1beta1.Policy{
				{
					Id:          testPolicyID,
					NamespaceId: "policy-1",
					RoleId:      "reader",
					ResourceId:  testResourceID,
					UserId:      testUserID,
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockPolicySrv := new(mocks.PolicyService)
			if tt.setup != nil {
				tt.setup(mockPolicySrv)
			}
			mockDep := Handler{policyService: mockPolicySrv}
			resp, err := mockDep.ListPolicies(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestCreatePolicy(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.PolicyService)
		req   *shieldv1beta1.CreatePolicyRequest
		want  *shieldv1beta1.CreatePolicyResponse
		err   error
	}{
		{
			title: "should return internal error if policy service return some error",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					NamespaceID: "team",
					RoleID:      "Admin",
				}).Return(policy.Policy{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				NamespaceId: "team",
				RoleId:      "Admin",
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad request error if foreign reference not exist",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					NamespaceID: "team",
					RoleID:      "Admin",
				}).Return(policy.Policy{}, policy.ErrInvalidDetail)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				NamespaceId: "team",
				RoleId:      "Admin",
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if policy service return nil error",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					NamespaceID: "policy-1",
					RoleID:      "reader",
				}).Return(policy.Policy{
					ID:          "test",
					NamespaceID: "policy-1",
					RoleID:      "reader",
				}, nil)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				NamespaceId: "policy-1",
				RoleId:      "reader",
			}},
			want: &shieldv1beta1.CreatePolicyResponse{Policy: &shieldv1beta1.Policy{
				Id:          "test",
				NamespaceId: "policy-1",
				RoleId:      "reader",
			},
			},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockPolicySrv := new(mocks.PolicyService)
			if tt.setup != nil {
				tt.setup(mockPolicySrv)
			}
			mockDep := Handler{policyService: mockPolicySrv}
			resp, err := mockDep.CreatePolicy(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_GetPolicy(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.PolicyService)
		request *shieldv1beta1.GetPolicyRequest
		want    *shieldv1beta1.GetPolicyResponse
		wantErr error
	}{
		{
			name: "should return internal error if policy service return some error",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testPolicyID).Return(policy.Policy{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetPolicyRequest{
				Id: testPolicyID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(policy.Policy{}, policy.ErrInvalidID)
			},
			request: &shieldv1beta1.GetPolicyRequest{},
			want:    nil,
			wantErr: grpcPolicyNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(policy.Policy{}, policy.ErrInvalidUUID)
			},
			request: &shieldv1beta1.GetPolicyRequest{
				Id: "some-id",
			},
			want:    nil,
			wantErr: grpcPolicyNotFoundErr,
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testPolicyID).Return(policy.Policy{}, policy.ErrNotExist)
			},
			request: &shieldv1beta1.GetPolicyRequest{
				Id: testPolicyID,
			},
			want:    nil,
			wantErr: grpcPolicyNotFoundErr,
		},
		{
			name: "should return success if policy service return nil error",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testPolicyID).Return(testPolicyMap[testPolicyID], nil)
			},
			request: &shieldv1beta1.GetPolicyRequest{
				Id: testPolicyID,
			},
			want: &shieldv1beta1.GetPolicyResponse{
				Policy: &shieldv1beta1.Policy{
					Id:          testPolicyID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					UserId:      testUserID,
					ResourceId:  testResourceID,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySrv := new(mocks.PolicyService)
			if tt.setup != nil {
				tt.setup(mockPolicySrv)
			}
			mockDep := Handler{policyService: mockPolicySrv}
			resp, err := mockDep.GetPolicy(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
