package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	modelv1 "github.com/odpf/shield/model/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

var testOrgMap = map[string]modelv1.Organization{
	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
		Id:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Name: "Org 1",
		Slug: "org-1",
		Metadata: map[string]string{
			"email": "org1@org1.com",
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
		want       *shieldv1.ListOrganizationsResponse
		err        error
	}{
		{
			title: "error in Org Service",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []modelv1.Organization, err error) {
				return []modelv1.Organization{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []modelv1.Organization, err error) {
				var testOrgList []modelv1.Organization
				for _, o := range testOrgMap {
					testOrgList = append(testOrgList, o)
				}
				return testOrgList, nil
			}},
			want: &shieldv1.ListOrganizationsResponse{Organizations: []*shieldv1.Organization{
				{
					Id:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
					Name: "Org 1",
					Slug: "org-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
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

			mockDep := Dep{OrgService: tt.mockOrgSrv}
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
		req        *shieldv1.CreateOrganizationRequest
		want       *shieldv1.CreateOrganizationResponse
		err        error
	}{
		{
			title: "error in fetching org list",
			mockOrgSrv: mockOrgSrv{CreateOrganizationFunc: func(ctx context.Context, o modelv1.Organization) (modelv1.Organization, error) {
				return modelv1.Organization{}, errors.New("some error")
			}},
			req: &shieldv1.CreateOrganizationRequest{Body: &shieldv1.OrganizationRequestBody{
				Name:     "some org",
				Slug:     "abc",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "int values in metadata map",
			req: &shieldv1.CreateOrganizationRequest{Body: &shieldv1.OrganizationRequestBody{
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
			mockOrgSrv: mockOrgSrv{CreateOrganizationFunc: func(ctx context.Context, o modelv1.Organization) (modelv1.Organization, error) {
				return modelv1.Organization{
					Id:       "new-abc",
					Name:     "some org",
					Slug:     "abc",
					Metadata: nil,
				}, nil
			}},
			req: &shieldv1.CreateOrganizationRequest{Body: &shieldv1.OrganizationRequestBody{
				Name: "some org",
				Slug: "abc",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("a"),
					},
				},
			}},
			want: &shieldv1.CreateOrganizationResponse{Organization: &shieldv1.Organization{
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

			mockDep := Dep{OrgService: tt.mockOrgSrv}
			resp, err := mockDep.CreateOrganization(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockOrgSrv struct {
	GetOrganizationFunc    func(ctx context.Context, id string) (modelv1.Organization, error)
	CreateOrganizationFunc func(ctx context.Context, org modelv1.Organization) (modelv1.Organization, error)
	ListOrganizationsFunc  func(ctx context.Context) ([]modelv1.Organization, error)
	UpdateOrganizationFunc func(ctx context.Context, toUpdate modelv1.Organization) (modelv1.Organization, error)
}

func (m mockOrgSrv) Get(ctx context.Context, id string) (modelv1.Organization, error) {
	return m.GetOrganizationFunc(ctx, id)
}

func (m mockOrgSrv) Create(ctx context.Context, org modelv1.Organization) (modelv1.Organization, error) {
	return m.CreateOrganizationFunc(ctx, org)
}

func (m mockOrgSrv) List(ctx context.Context) ([]modelv1.Organization, error) {
	return m.ListOrganizationsFunc(ctx)
}

func (m mockOrgSrv) Update(ctx context.Context, toUpdate modelv1.Organization) (modelv1.Organization, error) {
	return m.UpdateOrganizationFunc(ctx, toUpdate)
}
