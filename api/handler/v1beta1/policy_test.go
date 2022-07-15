package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/model"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testPolicyMap = map[string]model.Policy{
	"test": {
		Id: "test",
		Action: model.Action{
			Id:   "read",
			Name: "Read",
			Namespace: model.Namespace{
				Id:        "resource-1",
				Name:      "Resource 1",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		Namespace: model.Namespace{
			Id:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		Role: model.Role{
			Id:       "reader",
			Name:     "Reader",
			Metadata: map[string]any{},
			Namespace: model.Namespace{
				Id:        "resource-1",
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
			MockPolicySrv: mockPolicySrv{ListPoliciesFunc: func(ctx context.Context) (actions []model.Policy, err error) {
				return []model.Policy{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		},
		{
			title: "success",
			MockPolicySrv: mockPolicySrv{ListPoliciesFunc: func(ctx context.Context) (actions []model.Policy, err error) {
				var testPoliciesList []model.Policy
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

			mockDep := Dep{PolicyService: tt.MockPolicySrv}
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
			mockPolicySrv: mockPolicySrv{CreatePolicyFunc: func(ctx context.Context, policy model.Policy) ([]model.Policy, error) {
				return []model.Policy{}, errors.New("some error")
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
			mockPolicySrv: mockPolicySrv{CreatePolicyFunc: func(ctx context.Context, policy model.Policy) ([]model.Policy, error) {
				return []model.Policy{
					{
						Id: "test",
						Action: model.Action{
							Id:   "read",
							Name: "Read",
							Namespace: model.Namespace{
								Id:        "resource-1",
								Name:      "Resource 1",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							CreatedAt: time.Time{},
							UpdatedAt: time.Time{},
						},
						Namespace: model.Namespace{
							Id:        "resource-1",
							Name:      "Resource 1",
							CreatedAt: time.Time{},
							UpdatedAt: time.Time{},
						},
						Role: model.Role{
							Id:       "reader",
							Name:     "Reader",
							Metadata: map[string]any{},
							Namespace: model.Namespace{
								Id:        "resource-1",
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

			mockDep := Dep{PolicyService: tt.mockPolicySrv}
			resp, err := mockDep.CreatePolicy(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockPolicySrv struct {
	GetPolicyFunc    func(ctx context.Context, id string) (model.Policy, error)
	CreatePolicyFunc func(ctx context.Context, policy model.Policy) ([]model.Policy, error)
	ListPoliciesFunc func(ctx context.Context) ([]model.Policy, error)
	UpdatePolicyFunc func(ctx context.Context, id string, policy model.Policy) ([]model.Policy, error)
}

func (m mockPolicySrv) GetPolicy(ctx context.Context, id string) (model.Policy, error) {
	return m.GetPolicyFunc(ctx, id)
}

func (m mockPolicySrv) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	return m.ListPoliciesFunc(ctx)
}

func (m mockPolicySrv) CreatePolicy(ctx context.Context, policy model.Policy) ([]model.Policy, error) {
	return m.CreatePolicyFunc(ctx, policy)
}

func (m mockPolicySrv) UpdatePolicy(ctx context.Context, id string, policy model.Policy) ([]model.Policy, error) {
	return m.UpdatePolicyFunc(ctx, id, policy)
}
