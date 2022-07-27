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
			mockOrgSrv: mockOrgSrv{ListFunc: func(ctx context.Context) (organizations []organization.Organization, err error) {
				return []organization.Organization{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockOrgSrv: mockOrgSrv{ListFunc: func(ctx context.Context) (organizations []organization.Organization, err error) {
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
			mockOrgSrv: mockOrgSrv{CreateFunc: func(ctx context.Context, o organization.Organization) (organization.Organization, error) {
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
			mockOrgSrv: mockOrgSrv{CreateFunc: func(ctx context.Context, o organization.Organization) (organization.Organization, error) {
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
	GetByIDFunc           func(ctx context.Context, id string) (organization.Organization, error)
	GetBySlugFunc         func(ctx context.Context, slug string) (organization.Organization, error)
	CreateFunc            func(ctx context.Context, org organization.Organization) (organization.Organization, error)
	ListFunc              func(ctx context.Context) ([]organization.Organization, error)
	UpdateByIDFunc        func(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	UpdateBySlugFunc      func(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	AddAdminByIDFunc      func(ctx context.Context, id string, userIds []string) ([]user.User, error)
	AddAdminBySlugFunc    func(ctx context.Context, slug string, userIds []string) ([]user.User, error)
	RemoveAdminByIDFunc   func(ctx context.Context, id string, userId string) ([]user.User, error)
	RemoveAdminBySlugFunc func(ctx context.Context, slug string, userId string) ([]user.User, error)
	ListAdminsFunc        func(ctx context.Context, id string) ([]user.User, error)
}

func (m mockOrgSrv) GetByID(ctx context.Context, id string) (organization.Organization, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m mockOrgSrv) GetBySlug(ctx context.Context, id string) (organization.Organization, error) {
	return m.GetBySlugFunc(ctx, id)
}

func (m mockOrgSrv) Create(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	return m.CreateFunc(ctx, org)
}

func (m mockOrgSrv) List(ctx context.Context) ([]organization.Organization, error) {
	return m.ListFunc(ctx)
}

func (m mockOrgSrv) UpdateByID(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error) {
	return m.UpdateByIDFunc(ctx, toUpdate)
}

func (m mockOrgSrv) UpdateBySlug(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error) {
	return m.UpdateBySlugFunc(ctx, toUpdate)
}

func (m mockOrgSrv) AddAdminByID(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	return m.AddAdminByIDFunc(ctx, id, userIds)
}

func (m mockOrgSrv) AddAdminBySlug(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	return m.AddAdminBySlugFunc(ctx, id, userIds)
}

func (m mockOrgSrv) RemoveAdminByID(ctx context.Context, id string, userId string) ([]user.User, error) {
	return m.RemoveAdminByIDFunc(ctx, id, userId)
}

func (m mockOrgSrv) RemoveAdminBySlug(ctx context.Context, id string, userId string) ([]user.User, error) {
	return m.RemoveAdminBySlugFunc(ctx, id, userId)
}

func (m mockOrgSrv) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return m.ListAdminsFunc(ctx, id)
}
