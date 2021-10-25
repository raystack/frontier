package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/internal/org"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

var testOrgMap = map[string]org.Organization{
	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
		Id:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Name: "Org 1",
		Slug: "org-1",
		Metadata: map[string]interface{}{
			"email": "org1@org1.com",
			"count": 10,
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
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []org.Organization, err error) {
				return []org.Organization{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []org.Organization, err error) {
				var testOrgList []org.Organization
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
							"count": structpb.NewNumberValue(10),
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
			mockOrgSrv: mockOrgSrv{CreateOrganizationFunc: func(ctx context.Context, o org.Organization) (org.Organization, error) {
				return org.Organization{}, errors.New("some error")
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
			title: "success",
			mockOrgSrv: mockOrgSrv{CreateOrganizationFunc: func(ctx context.Context, o org.Organization) (org.Organization, error) {
				return org.Organization{
					Id:       "new-abc",
					Name:     "some org",
					Slug:     "abc",
					Metadata: nil,
				}, nil
			}},
			req: &shieldv1.CreateOrganizationRequest{Body: &shieldv1.OrganizationRequestBody{
				Name:     "some org",
				Slug:     "abc",
				Metadata: &structpb.Struct{},
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
	GetOrganizationFunc    func(ctx context.Context, id string) (org.Organization, error)
	CreateOrganizationFunc func(ctx context.Context, org org.Organization) (org.Organization, error)
	ListOrganizationsFunc  func(ctx context.Context) ([]org.Organization, error)
	UpdateOrganizationFunc func(ctx context.Context, toUpdate org.Organization) (org.Organization, error)
}

func (m mockOrgSrv) GetOrganization(ctx context.Context, id string) (org.Organization, error) {
	return m.GetOrganizationFunc(ctx, id)
}

func (m mockOrgSrv) CreateOrganization(ctx context.Context, org org.Organization) (org.Organization, error) {
	return m.CreateOrganizationFunc(ctx, org)
}

func (m mockOrgSrv) ListOrganizations(ctx context.Context) ([]org.Organization, error) {
	return m.ListOrganizationsFunc(ctx)
}

func (m mockOrgSrv) UpdateOrganization(ctx context.Context, toUpdate org.Organization) (org.Organization, error) {
	return m.UpdateOrganizationFunc(ctx, toUpdate)
}
