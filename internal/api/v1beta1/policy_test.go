package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/shield/pkg/utils"

	"github.com/raystack/shield/internal/bootstrap/schema"

	"github.com/raystack/shield/core/policy"
	"github.com/raystack/shield/internal/api/v1beta1/mocks"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	testPolicyID           = utils.NewString()
	testPolicyResourceType = "compute"
	testPolicyMap          = map[string]policy.Policy{
		testPolicyID: {
			ID:            testPolicyID,
			PrincipalType: schema.UserPrincipal,
			PrincipalID:   testUserID,
			ResourceID:    testResourceID,
			ResourceType:  testPolicyResourceType,
			RoleID:        "reader",
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
					Id:        testPolicyID,
					RoleId:    "reader",
					Resource:  schema.JoinNamespaceAndResourceID(testPolicyResourceType, testResourceID),
					Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, testUserID),
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
					RoleID:        "Admin",
					ResourceID:    "id",
					ResourceType:  "ns",
					PrincipalID:   "id",
					PrincipalType: "ns",
				}).Return(policy.Policy{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				RoleId:    "Admin",
				Resource:  "ns:id",
				Principal: "ns:id",
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad request error if foreign reference not exist",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					RoleID:        "Admin",
					ResourceID:    "id",
					ResourceType:  "ns",
					PrincipalID:   "id",
					PrincipalType: "ns",
				}).Return(policy.Policy{}, policy.ErrInvalidDetail)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				RoleId:    "Admin",
				Resource:  "ns:id",
				Principal: "ns:id",
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if policy service return nil error",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ResourceType:  testPolicyResourceType,
					RoleID:        "reader",
					ResourceID:    "id",
					PrincipalID:   "id",
					PrincipalType: testPolicyResourceType,
				}).Return(policy.Policy{
					ID:           "test",
					ResourceType: testPolicyResourceType,
					ResourceID:   "id",
					RoleID:       "reader",
				}, nil)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				RoleId:    "reader",
				Resource:  schema.JoinNamespaceAndResourceID(testPolicyResourceType, "id"),
				Principal: schema.JoinNamespaceAndResourceID(testPolicyResourceType, "id"),
			}},
			want: &shieldv1beta1.CreatePolicyResponse{Policy: &shieldv1beta1.Policy{
				Id:        "test",
				RoleId:    "reader",
				Resource:  schema.JoinNamespaceAndResourceID(testPolicyResourceType, "id"),
				Principal: ":",
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
					Id:        testPolicyID,
					RoleId:    testPolicyMap[testPolicyID].RoleID,
					Resource:  schema.JoinNamespaceAndResourceID(testPolicyMap[testPolicyID].ResourceType, testResourceID),
					Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, testUserID),
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
