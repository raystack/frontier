package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/core/user"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/shield/core/organization"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var testOrgMap = map[string]organization.Organization{
	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
		ID:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Name: "Org 1",
		Slug: "org-1",
		Metadata: map[string]any{
			"email":  "org1@org1.com",
			"age":    21,
			"intern": true,
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestListOrganizations(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		mockOrgSrv mockOrgSrv
		want       *shieldv1beta1.ListOrganizationsResponse
		err        error
	}{
		{
			title: "error in Org Service",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []organization.Organization, err error) {
				return []organization.Organization{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []organization.Organization, err error) {
				var testOrgList []organization.Organization
				for _, o := range testOrgMap {
					testOrgList = append(testOrgList, o)
				}
				return testOrgList, nil
			}},
			want: &shieldv1beta1.ListOrganizationsResponse{Organizations: []*shieldv1beta1.Organization{
				{
					Id:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
					Name: "Org 1",
					Slug: "org-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email":  structpb.NewStringValue("org1@org1.com"),
							"age":    structpb.NewNumberValue(21),
							"intern": structpb.NewBoolValue(true),
						},
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

			mockDep := Handler{orgService: tt.mockOrgSrv}
			resp, err := mockDep.ListOrganizations(context.Background(), nil)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestCreateOrganization(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		mockOrgSrv mockOrgSrv
		req        *shieldv1beta1.CreateOrganizationRequest
		want       *shieldv1beta1.CreateOrganizationResponse
		err        error
	}{
		{
			title: "error in fetching org list",
			mockOrgSrv: mockOrgSrv{CreateOrganizationFunc: func(ctx context.Context, o organization.Organization) (organization.Organization, error) {
				return organization.Organization{}, errors.New("some error")
			}},
			req: &shieldv1beta1.CreateOrganizationRequest{Body: &shieldv1beta1.OrganizationRequestBody{
				Name:     "some org",
				Slug:     "abc",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "int values in metadata map",
			req: &shieldv1beta1.CreateOrganizationRequest{Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "some org",
				Slug: "abc",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"count": structpb.NewNumberValue(10),
					},
				},
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "success",
			mockOrgSrv: mockOrgSrv{CreateOrganizationFunc: func(ctx context.Context, o organization.Organization) (organization.Organization, error) {
				return organization.Organization{
					ID:       "new-abc",
					Name:     "some org",
					Slug:     "abc",
					Metadata: nil,
				}, nil
			}},
			req: &shieldv1beta1.CreateOrganizationRequest{Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "some org",
				Slug: "abc",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("a"),
					},
				},
			}},
			want: &shieldv1beta1.CreateOrganizationResponse{Organization: &shieldv1beta1.Organization{
				Id:        "new-abc",
				Name:      "some org",
				Slug:      "abc",
				Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Handler{orgService: tt.mockOrgSrv}
			resp, err := mockDep.CreateOrganization(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockOrgSrv struct {
	GetOrganizationFunc    func(ctx context.Context, id string) (organization.Organization, error)
	CreateOrganizationFunc func(ctx context.Context, org organization.Organization) (organization.Organization, error)
	ListOrganizationsFunc  func(ctx context.Context) ([]organization.Organization, error)
	UpdateOrganizationFunc func(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	AddAdminFunc           func(ctx context.Context, id string, userIds []string) ([]user.User, error)
	ListAdminsFunc         func(ctx context.Context, id string) ([]user.User, error)
	RemoveAdminFunc        func(ctx context.Context, id string, userId string) ([]user.User, error)
}

func (m mockOrgSrv) Get(ctx context.Context, id string) (organization.Organization, error) {
	return m.GetOrganizationFunc(ctx, id)
}

func (m mockOrgSrv) Create(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	return m.CreateOrganizationFunc(ctx, org)
}

func (m mockOrgSrv) List(ctx context.Context) ([]organization.Organization, error) {
	return m.ListOrganizationsFunc(ctx)
}

func (m mockOrgSrv) Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error) {
	return m.UpdateOrganizationFunc(ctx, toUpdate)
}

func (m mockOrgSrv) AddAdmin(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	return m.AddAdminFunc(ctx, id, userIds)
}

func (m mockOrgSrv) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return m.ListAdminsFunc(ctx, id)
}

func (m mockOrgSrv) RemoveAdmin(ctx context.Context, id string, userId string) ([]user.User, error) {
	return m.RemoveAdminFunc(ctx, id, userId)
}
