// Package v1beta1 provides v1beta1  
package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/pkg/metadata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
		Metadata: metadata.Metadata{
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
		title string
		setup func(os *mocks.OrganizationService)
		want  *shieldv1beta1.ListOrganizationsResponse
		err   error
	}{
		{
			title: "error in Org Service",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]organization.Organization{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		}, {
			title: "success",
			setup: func(os *mocks.OrganizationService) {
				var testOrgList []organization.Organization
				for _, o := range testOrgMap {
					testOrgList = append(testOrgList, o)
				}
				os.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return(testOrgList, nil)
			},
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

			mockOrgSrv := new(mocks.OrganizationService)
			mockDep := Handler{orgService: mockOrgSrv}
			if tt.setup != nil {
				tt.setup(mockOrgSrv)
			}
			resp, err := mockDep.ListOrganizations(context.Background(), nil)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestCreateOrganization(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		setup func(os *mocks.OrganizationService)
		req   *shieldv1beta1.CreateOrganizationRequest
		want  *shieldv1beta1.CreateOrganizationResponse
		err   error
	}{
		{
			title: "error in fetching org list",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), organization.Organization{
					Name:     "some org",
					Slug:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, errors.New("some error"))
			},
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
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), organization.Organization{
					Name: "some org",
					Slug: "abc",
					Metadata: metadata.Metadata{
						"email": "a",
					},
				}).Return(organization.Organization{
					ID:   "new-abc",
					Name: "some org",
					Slug: "abc",
					Metadata: metadata.Metadata{
						"email": "a",
					},
				}, nil)
			},
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
				Id:   "new-abc",
				Name: "some org",
				Slug: "abc",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("a"),
					}},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()
			mockOrgSrv := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockOrgSrv)
			}
			mockDep := Handler{orgService: mockOrgSrv}
			resp, err := mockDep.CreateOrganization(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}
