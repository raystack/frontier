package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/shield/core/policy"
	"github.com/raystack/shield/internal/api/v1beta1/mocks"
	"github.com/raystack/shield/pkg/uuid"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
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
			ActionID:    "read",
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
				ps.EXPECT().List(mock.Anything).Return([]policy.Policy{}, errors.New("some error"))
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
				ps.EXPECT().List(mock.Anything).Return(testPoliciesList, nil)
			},
			want: &shieldv1beta1.ListPoliciesResponse{Policies: []*shieldv1beta1.Policy{
				{
					Id:          testPolicyID,
					ActionId:    "read",
					NamespaceId: "policy-1",
					RoleId:      "reader",
					// @TODO(krtkvrm): issues/171
					//Action: &shieldv1beta1.Action{
					//	Id:   "read",
					//	Name: "Read",
					//	// @TODO(krtkvrm): issues/171
					//	//Namespace: &shieldv1beta1.Namespace{
					//	//	Id:        "policy-1",
					//	//	Name:      "Policy 1",
					//	//	CreatedAt: timestamppb.New(time.Time{}),
					//	//	UpdatedAt: timestamppb.New(time.Time{}),
					//	//},
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					//Namespace: &shieldv1beta1.Namespace{
					//	Id:        "policy-1",
					//	Name:      "Policy 1",
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					//Role: &shieldv1beta1.Role{
					//	Id:       "reader",
					//	Name:     "Reader",
					//	Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					//	// TODO(krtkvrm): remove namespace from role proto and replace it with namespaceid
					//	//Namespace: &shieldv1beta1.Namespace{
					//	//	Id:        "policy-1",
					//	//	Name:      "Policy 1",
					//	//	CreatedAt: timestamppb.New(time.Time{}),
					//	//	UpdatedAt: timestamppb.New(time.Time{}),
					//	//},
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
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
					ActionID:    "add-member",
				}).Return([]policy.Policy{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				NamespaceId: "team",
				RoleId:      "Admin",
				ActionId:    "add-member",
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
					ActionID:    "add-member",
				}).Return([]policy.Policy{}, policy.ErrInvalidDetail)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				NamespaceId: "team",
				RoleId:      "Admin",
				ActionId:    "add-member",
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
					ActionID:    "read",
				}).Return([]policy.Policy{
					{
						ID:          "test",
						ActionID:    "read",
						NamespaceID: "policy-1",
						RoleID:      "reader",
					},
				}, nil)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				ActionId:    "read",
				NamespaceId: "policy-1",
				RoleId:      "reader",
			}},
			want: &shieldv1beta1.CreatePolicyResponse{Policies: []*shieldv1beta1.Policy{
				{
					Id:          "test",
					NamespaceId: "policy-1",
					ActionId:    "read",
					RoleId:      "reader",
					// @TODO(krtkvrm): issues/171
					//Action: &shieldv1beta1.Action{
					//	Id:   "read",
					//	Name: "Read",
					//	// @TODO(krtkvrm): issues/171
					//	//Namespace: &shieldv1beta1.Namespace{
					//	//	Id:        "policy-1",
					//	//	Name:      "Policy 1",
					//	//	CreatedAt: timestamppb.New(time.Time{}),
					//	//	UpdatedAt: timestamppb.New(time.Time{}),
					//	//},
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					//Namespace: &shieldv1beta1.Namespace{
					//	Id:        "policy-1",
					//	Name:      "Policy 1",
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					//Role: &shieldv1beta1.Role{
					//	Id:       "reader",
					//	Name:     "Reader",
					//	Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					//	// TODO(krtkvrm): remove namespace from role proto and replace it with namespaceid
					//	//Namespace: &shieldv1beta1.Namespace{
					//	//	Id:        "policy-1",
					//	//	Name:      "Policy 1",
					//	//	CreatedAt: timestamppb.New(time.Time{}),
					//	//	UpdatedAt: timestamppb.New(time.Time{}),
					//	//},
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
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
					ActionId:    "read",
					NamespaceId: "policy-1",
					RoleId:      "reader",
					// // @TODO(krtkvrm): issues/171
					//Role: &shieldv1beta1.Role{
					//	Id:   testPolicyMap[testPolicyID].Role.ID,
					//	Name: testPolicyMap[testPolicyID].Role.Name,
					//	Metadata: &structpb.Struct{
					//		Fields: make(map[string]*structpb.Value),
					//	},
					//	// TODO(krtkvrm): remove namespace from role proto and replace it with namespaceid
					//	//Namespace: &shieldv1beta1.Namespace{
					//	//	Id:        testPolicyMap[testPolicyID].Namespace.ID,
					//	//	Name:      testPolicyMap[testPolicyID].Namespace.Name,
					//	//	CreatedAt: timestamppb.New(time.Time{}),
					//	//	UpdatedAt: timestamppb.New(time.Time{}),
					//	//},
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					//Action: &shieldv1beta1.Action{
					//	Id:   testPolicyMap[testPolicyID].ActionID,
					//	Name: testPolicyMap[testPolicyID].Action.Name,
					//	// @TODO(krtkvrm): remove namespace from role proto and replace it with namespaceid
					//	//Namespace: &shieldv1beta1.Namespace{
					//	//	Id:        testPolicyMap[testPolicyID].Namespace.ID,
					//	//	Name:      testPolicyMap[testPolicyID].Namespace.Name,
					//	//	CreatedAt: timestamppb.New(time.Time{}),
					//	//	UpdatedAt: timestamppb.New(time.Time{}),
					//	//},
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					//Namespace: &shieldv1beta1.Namespace{
					//	Id:        testPolicyMap[testPolicyID].Namespace.ID,
					//	Name:      testPolicyMap[testPolicyID].Namespace.Name,
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
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

func TestHandler_UpdatePolicy(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.PolicyService)
		request *shieldv1beta1.UpdatePolicyRequest
		want    *shieldv1beta1.UpdatePolicyResponse
		wantErr error
	}{
		{
			name: "should return internal error if policy service return some error",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ID:          testPolicyMap[testPolicyID].ID,
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return([]policy.Policy{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Id: testPolicyID,
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return([]policy.Policy{}, policy.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want:    nil,
			wantErr: grpcPolicyNotFoundErr,
		},
		{
			name: "should return not found error if id is not exist",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ID:          testPolicyMap[testPolicyID].ID,
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return([]policy.Policy{}, policy.ErrNotExist)
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Id: testPolicyID,
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want:    nil,
			wantErr: grpcPolicyNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ID:          "some-id",
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return([]policy.Policy{}, policy.ErrInvalidUUID)
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Id: "some-id",
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want:    nil,
			wantErr: grpcPolicyNotFoundErr,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ID:          testPolicyMap[testPolicyID].ID,
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return([]policy.Policy{}, policy.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Id: testPolicyID,
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return already exist error if policy service return err conflict",
			setup: func(rs *mocks.PolicyService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ID:          testPolicyMap[testPolicyID].ID,
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return([]policy.Policy{}, policy.ErrConflict)
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Id: testPolicyID,
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return success if policy service return nil",
			setup: func(rs *mocks.PolicyService) {
				var testPoliciesList []policy.Policy
				for _, p := range testPolicyMap {
					testPoliciesList = append(testPoliciesList, p)
				}
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), policy.Policy{
					ID:          testPolicyMap[testPolicyID].ID,
					RoleID:      testPolicyMap[testPolicyID].RoleID,
					NamespaceID: testPolicyMap[testPolicyID].NamespaceID,
					ActionID:    testPolicyMap[testPolicyID].ActionID,
				}).Return(testPoliciesList, nil)
			},
			request: &shieldv1beta1.UpdatePolicyRequest{
				Id: testPolicyID,
				Body: &shieldv1beta1.PolicyRequestBody{
					RoleId:      testPolicyMap[testPolicyID].RoleID,
					NamespaceId: testPolicyMap[testPolicyID].NamespaceID,
					ActionId:    testPolicyMap[testPolicyID].ActionID,
				},
			},
			want: &shieldv1beta1.UpdatePolicyResponse{
				Policies: []*shieldv1beta1.Policy{
					{
						Id:          testPolicyID,
						ActionId:    "read",
						NamespaceId: "policy-1",
						RoleId:      "reader",
						// @TODO(krtkvrm): issues/171
						//Role: &shieldv1beta1.Role{
						//	Id:   testPolicyMap[testPolicyID].Role.ID,
						//	Name: testPolicyMap[testPolicyID].Role.Name,
						//	Metadata: &structpb.Struct{
						//		Fields: make(map[string]*structpb.Value),
						//	},
						//	// @TODO(krtkvrm): remove namespace from role proto and replace it with namespaceid
						//	//Namespace: &shieldv1beta1.Namespace{
						//	//	Id:        testPolicyMap[testPolicyID].Namespace.ID,
						//	//	Name:      testPolicyMap[testPolicyID].Namespace.Name,
						//	//	CreatedAt: timestamppb.New(time.Time{}),
						//	//	UpdatedAt: timestamppb.New(time.Time{}),
						//	//},
						//	CreatedAt: timestamppb.New(time.Time{}),
						//	UpdatedAt: timestamppb.New(time.Time{}),
						//},
						//Action: &shieldv1beta1.Action{
						//	Id:   testPolicyMap[testPolicyID].Action.ID,
						//	Name: testPolicyMap[testPolicyID].Action.Name,
						//	// @TODO(krtkvrm): remove namespace from role proto and replace it with namespaceid
						//	//Namespace: &shieldv1beta1.Namespace{
						//	//	Id:        testPolicyMap[testPolicyID].Namespace.ID,
						//	//	Name:      testPolicyMap[testPolicyID].Namespace.Name,
						//	//	CreatedAt: timestamppb.New(time.Time{}),
						//	//	UpdatedAt: timestamppb.New(time.Time{}),
						//	//},
						//	CreatedAt: timestamppb.New(time.Time{}),
						//	UpdatedAt: timestamppb.New(time.Time{}),
						//},
						//Namespace: &shieldv1beta1.Namespace{
						//	Id:        testPolicyMap[testPolicyID].Namespace.ID,
						//	Name:      testPolicyMap[testPolicyID].Namespace.Name,
						//	CreatedAt: timestamppb.New(time.Time{}),
						//	UpdatedAt: timestamppb.New(time.Time{}),
						//},
					},
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
			resp, err := mockDep.UpdatePolicy(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
