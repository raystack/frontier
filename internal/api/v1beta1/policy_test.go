package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/pkg/metadata"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testPolicyMap = map[string]policy.Policy{
	"test": {
		ID: "test",
		Action: action.Action{
			ID:   "read",
			Name: "Read",
			Namespace: namespace.Namespace{
				ID:        "resource-1",
				Name:      "Resource 1",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		Namespace: namespace.Namespace{
			ID:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		Role: role.Role{
			ID:       "reader",
			Name:     "Reader",
			Metadata: metadata.Metadata{},
			Namespace: namespace.Namespace{
				ID:        "resource-1",
				Name:      "Resource 1",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
		},
	},
}

func TestListPolicies(t *testing.T) {
	t.Parallel()
	table := []struct {
		title string
		setup func(ps *mocks.PolicyService)
		req   *shieldv1beta1.ListPoliciesRequest
		want  *shieldv1beta1.ListPoliciesResponse
		err   error
	}{
		{
			title: "error in Policy Service",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().List(mock.Anything).Return([]policy.Policy{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "success",
			setup: func(ps *mocks.PolicyService) {
				var testPoliciesList []policy.Policy
				for _, p := range testPolicyMap {
					testPoliciesList = append(testPoliciesList, p)
				}
				ps.EXPECT().List(mock.Anything).Return(testPoliciesList, nil)
			},
			want: &shieldv1beta1.ListPoliciesResponse{Policies: []*shieldv1beta1.Policy{
				{
					Id: "test",
					Action: &shieldv1beta1.Action{
						Id:   "read",
						Name: "Read",
						Namespace: &shieldv1beta1.Namespace{
							Id:        "resource-1",
							Name:      "Resource 1",
							CreatedAt: timestamppb.New(time.Time{}),
							UpdatedAt: timestamppb.New(time.Time{}),
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					Namespace: &shieldv1beta1.Namespace{
						Id:        "resource-1",
						Name:      "Resource 1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					Role: &shieldv1beta1.Role{
						Id:       "reader",
						Name:     "Reader",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						Namespace: &shieldv1beta1.Namespace{
							Id:        "resource-1",
							Name:      "Resource 1",
							CreatedAt: timestamppb.New(time.Time{}),
							UpdatedAt: timestamppb.New(time.Time{}),
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

	table := []struct {
		title string
		setup func(ps *mocks.PolicyService)
		req   *shieldv1beta1.CreatePolicyRequest
		want  *shieldv1beta1.CreatePolicyResponse
		err   error
	}{
		{
			title: "error in creating policy",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.Anything, mock.Anything).Return([]policy.Policy{}, errors.New("some error"))
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
			title: "success",
			setup: func(ps *mocks.PolicyService) {
				ps.EXPECT().Create(mock.Anything, mock.Anything).Return([]policy.Policy{
					{
						ID: "test",
						Action: action.Action{
							ID:   "read",
							Name: "Read",
							Namespace: namespace.Namespace{
								ID:        "resource-1",
								Name:      "Resource 1",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							CreatedAt: time.Time{},
							UpdatedAt: time.Time{},
						},
						Namespace: namespace.Namespace{
							ID:        "resource-1",
							Name:      "Resource 1",
							CreatedAt: time.Time{},
							UpdatedAt: time.Time{},
						},
						Role: role.Role{
							ID:       "reader",
							Name:     "Reader",
							Metadata: metadata.Metadata{},
							Namespace: namespace.Namespace{
								ID:        "resource-1",
								Name:      "Resource 1",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
						},
					},
				}, nil)
			},
			req: &shieldv1beta1.CreatePolicyRequest{Body: &shieldv1beta1.PolicyRequestBody{
				ActionId:    "read",
				NamespaceId: "resource-1",
				RoleId:      "reader",
			}},
			want: &shieldv1beta1.CreatePolicyResponse{Policies: []*shieldv1beta1.Policy{
				{
					Id: "test",
					Action: &shieldv1beta1.Action{
						Id:   "read",
						Name: "Read",
						Namespace: &shieldv1beta1.Namespace{
							Id:        "resource-1",
							Name:      "Resource 1",
							CreatedAt: timestamppb.New(time.Time{}),
							UpdatedAt: timestamppb.New(time.Time{}),
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					Namespace: &shieldv1beta1.Namespace{
						Id:        "resource-1",
						Name:      "Resource 1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					Role: &shieldv1beta1.Role{
						Id:       "reader",
						Name:     "Reader",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						Namespace: &shieldv1beta1.Namespace{
							Id:        "resource-1",
							Name:      "Resource 1",
							CreatedAt: timestamppb.New(time.Time{}),
							UpdatedAt: timestamppb.New(time.Time{}),
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

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
