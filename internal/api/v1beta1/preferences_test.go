package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_DescribePreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *frontierv1beta1.DescribePreferencesRequest
		want    *frontierv1beta1.DescribePreferencesResponse
		wantErr error
	}{
		{
			name: "should describe preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Describe(mock.AnythingOfType("*context.emptyCtx")).Return([]preference.Trait{{
					ResourceType:    "resource",
					Name:            "some_name",
					Title:           "some_title",
					Description:     "some_description",
					LongDescription: "some_long_description",
					Heading:         "some_heading",
					SubHeading:      "some_sub_heading",
					Breadcrumb:      "some_breadcrumb",
					InputHints:      "some_inputHints",
				}})
			},
			req: &frontierv1beta1.DescribePreferencesRequest{},
			want: &frontierv1beta1.DescribePreferencesResponse{
				Traits: []*frontierv1beta1.PreferenceTrait{
					{
						ResourceType:    "resource",
						Name:            "some_name",
						Title:           "some_title",
						Description:     "some_description",
						LongDescription: "some_long_description",
						Heading:         "some_heading",
						SubHeading:      "some_sub_heading",
						Breadcrumb:      "some_breadcrumb",
						InputHints:      "some_inputHints",
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			mockPref := Handler{preferenceService: mockPreferenceServ}
			res, err := mockPref.DescribePreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, res)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_CreateOrganizationPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *frontierv1beta1.CreateOrganizationPreferencesRequest
		want    *frontierv1beta1.CreateOrganizationPreferencesResponse
		wantErr error
	}{
		{
			name: "should create organization preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), preference.Preference{
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.OrganizationNamespace,
				}).Return(preference.Preference{
					ID:           "some_id",
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.OrganizationNamespace,
				}, nil)
			},
			req: &frontierv1beta1.CreateOrganizationPreferencesRequest{
				Id: "some_resource_id",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "some_name",
						Value: "some_value",
					},
				},
			},
			want: &frontierv1beta1.CreateOrganizationPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "some_id",
						Name:         "some_name",
						Value:        "some_value",
						ResourceId:   "some_resource_id",
						ResourceType: schema.OrganizationNamespace,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			mockPref := Handler{preferenceService: mockPreferenceServ}
			res, err := mockPref.CreateOrganizationPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, res)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_ListOrganizationPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *frontierv1beta1.ListOrganizationPreferencesRequest
		want    *frontierv1beta1.ListOrganizationPreferencesResponse
		wantErr error
	}{
		{
			name: "should list Organization Preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), preference.Filter{
					OrgID: "some_id",
				}).Return([]preference.Preference{
					{
						ID:           "some_id",
						Name:         "some_name",
						Value:        "some_value",
						ResourceID:   "some_resource_id",
						ResourceType: "some_resource_type",
					},
				}, nil)
			},
			req: &frontierv1beta1.ListOrganizationPreferencesRequest{
				Id: "some_id",
			},
			want: &frontierv1beta1.ListOrganizationPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "some_id",
						Name:         "some_name",
						Value:        "some_value",
						ResourceId:   "some_resource_id",
						ResourceType: "some_resource_type",
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{})},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			mockPref := Handler{preferenceService: mockPreferenceServ}
			res, err := mockPref.ListOrganizationPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, res)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_CreateUserPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService, a *mocks.AuthnService)
		req     *frontierv1beta1.CreateCurrentUserPreferencesRequest
		want    *frontierv1beta1.CreateCurrentUserPreferencesResponse
		wantErr error
	}{
		{
			name: "should create user preference on success",
			setup: func(m *mocks.PreferenceService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "some_resource_id",
				}, nil)
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), preference.Preference{
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{
					ID:           "some_id",
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.UserPrincipal,
				}, nil)
			},
			req: &frontierv1beta1.CreateCurrentUserPreferencesRequest{
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "some_name",
						Value: "some_value",
					},
				},
			},
			want: &frontierv1beta1.CreateCurrentUserPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "some_id",
						Name:         "some_name",
						Value:        "some_value",
						ResourceId:   "some_resource_id",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should return preference service return some error",
			setup: func(m *mocks.PreferenceService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "some_resource_id",
				}, nil)
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), preference.Preference{
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{}, errors.New("some_error"))
			},
			req: &frontierv1beta1.CreateCurrentUserPreferencesRequest{
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "some_name",
						Value: "some_value",
					},
				},
			},
			want:    nil,
			wantErr: status.Errorf(codes.Internal, errors.New("some_error").Error()),
		},
		{
			name: "should return error if authenServ return some error",
			setup: func(m *mocks.PreferenceService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{}, errors.New("some_error_auth"))
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), preference.Preference{
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{}, grpcInternalServerError)
			},
			req: &frontierv1beta1.CreateCurrentUserPreferencesRequest{
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "some_name",
						Value: "some_value",
					},
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			mockAuthServ := new(mocks.AuthnService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ, mockAuthServ)
			}
			mockPref := Handler{
				preferenceService: mockPreferenceServ,
				authnService:      mockAuthServ,
			}
			res, err := mockPref.CreateCurrentUserPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, res)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
