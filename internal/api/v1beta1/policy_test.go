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
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	"github.com/stretchr/testify/assert"
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
			Metadata: map[string]any{},
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
		title         string
		MockPolicySrv mockPolicySrv
		req           *shieldv1beta1.ListPoliciesRequest
		want          *shieldv1beta1.ListPoliciesResponse
		err           error
	}{
		{
			title: "error in Policy Service",
			MockPolicySrv: mockPolicySrv{ListFunc: func(ctx context.Context) (actions []policy.Policy, err error) {
				return []policy.Policy{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		},
		{
			title: "success",
			MockPolicySrv: mockPolicySrv{ListFunc: func(ctx context.Context) (actions []policy.Policy, err error) {
				var testPoliciesList []policy.Policy
				for _, p := range testPolicyMap {
					testPoliciesList = append(testPoliciesList, p)
				}
				return testPoliciesList, nil
			}},
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

			mockDep := Handler{policyService: tt.MockPolicySrv}
			resp, err := mockDep.ListPolicies(context.Background(), tt.req)

			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestCreatePolicy(t *testing.T) {
	t.Parallel()

	table := []struct {
		title         string
		mockPolicySrv mockPolicySrv
		req           *shieldv1beta1.CreatePolicyRequest
		want          *shieldv1beta1.CreatePolicyResponse
		err           error
	}{
		{
			title: "error in creating policy",
			mockPolicySrv: mockPolicySrv{CreateFunc: func(ctx context.Context, pol policy.Policy) ([]policy.Policy, error) {
				return []policy.Policy{}, errors.New("some error")
			}},
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
			mockPolicySrv: mockPolicySrv{CreateFunc: func(ctx context.Context, pol policy.Policy) ([]policy.Policy, error) {
				return []policy.Policy{
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
							Metadata: map[string]any{},
							Namespace: namespace.Namespace{
								ID:        "resource-1",
								Name:      "Resource 1",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
						},
					},
				}, nil
			}},
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

			mockDep := Handler{policyService: tt.mockPolicySrv}
			resp, err := mockDep.CreatePolicy(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockPolicySrv struct {
	GetFunc    func(ctx context.Context, id string) (policy.Policy, error)
	CreateFunc func(ctx context.Context, pol policy.Policy) ([]policy.Policy, error)
	ListFunc   func(ctx context.Context) ([]policy.Policy, error)
	UpdateFunc func(ctx context.Context, pol policy.Policy) ([]policy.Policy, error)
}

func (m mockPolicySrv) Get(ctx context.Context, id string) (policy.Policy, error) {
	return m.GetFunc(ctx, id)
}

func (m mockPolicySrv) List(ctx context.Context) ([]policy.Policy, error) {
	return m.ListFunc(ctx)
}

func (m mockPolicySrv) Create(ctx context.Context, pol policy.Policy) ([]policy.Policy, error) {
	return m.CreateFunc(ctx, pol)
}

func (m mockPolicySrv) Update(ctx context.Context, pol policy.Policy) ([]policy.Policy, error) {
	return m.UpdateFunc(ctx, pol)
}
