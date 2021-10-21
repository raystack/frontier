package v1

import (
	"context"
	"errors"
	"testing"

	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
	"github.com/odpf/shield/org"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//var testOrgMap = map[string]org.Organization{
//	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
//		Id:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
//		Name: "Org 1",
//		Slug: "org-1",
//		Metadata: map[string]interface{}{
//			"email": "org1@org1.com",
//			"count": 10,
//		},
//		CreatedAt: time.Time{},
//		UpdatedAt: time.Time{},
//	}, "a755e76c-31a3-11ec-8d3d-0242ac130003": {
//		Id:   "a755e76c-31a3-11ec-8d3d-0242ac130003",
//		Name: "Org 2",
//		Slug: "org-2",
//		Metadata: map[string]interface{}{
//			"email":       "org2@org2.com",
//			"admin_count": 2,
//		},
//		CreatedAt: time.Time{},
//		UpdatedAt: time.Time{},
//	}, "b07131f8-31a3-11ec-8d3d-0242ac130003": {
//		Id:   "b07131f8-31a3-11ec-8d3d-0242ac130003",
//		Name: "Org 3",
//		Slug: "org-3",
//		Metadata: map[string]interface{}{
//			"email": "org3@org3.com",
//		},
//		CreatedAt: time.Time{},
//		UpdatedAt: time.Time{},
//	}, "bb97e9e6-31a3-11ec-8d3d-0242ac130003": {
//		Id:   "bb97e9e6-31a3-11ec-8d3d-0242ac130003",
//		Name: "Org 4",
//		Slug: "org-4",
//		Metadata: map[string]interface{}{
//			"Project": "Project 1",
//		},
//		CreatedAt: time.Time{},
//		UpdatedAt: time.Time{},
//	},
//}

func TestListOrganizations(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		mockOrgSrv mockOrgSrv
		want       *shieldv1.ListOrganizationsResponse
		err        error
	}{
		{
			title: "error in fetching org list",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []org.Organization, err error) {
				return []org.Organization{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Dep{OrgService: tt.mockOrgSrv}
			resp, err := mockDep.ListOrganizations(context.Background(), nil)
			assert.Equal(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestCreateOrganization(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		mockOrgSrv mockOrgSrv
		want       *shieldv1.ListOrganizationsResponse
		err        error
	}{
		{
			title: "error in fetching org list",
			mockOrgSrv: mockOrgSrv{ListOrganizationsFunc: func(ctx context.Context) (organizations []org.Organization, err error) {
				return []org.Organization{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Dep{OrgService: tt.mockOrgSrv}
			resp, err := mockDep.ListOrganizations(context.Background(), nil)
			assert.Equal(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
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
