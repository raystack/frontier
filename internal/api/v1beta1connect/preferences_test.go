package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_DescribePreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *connect.Request[frontierv1beta1.DescribePreferencesRequest]
		want    *connect.Response[frontierv1beta1.DescribePreferencesResponse]
		wantErr error
	}{
		{
			name: "should describe preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Describe(mock.AnythingOfType("context.backgroundCtx")).Return([]preference.Trait{{
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
			req: connect.NewRequest(&frontierv1beta1.DescribePreferencesRequest{}),
			want: connect.NewResponse(&frontierv1beta1.DescribePreferencesResponse{
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
			}),
			wantErr: nil,
		},
		{
			name: "should return empty traits list when service returns empty slice",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Describe(mock.AnythingOfType("context.backgroundCtx")).Return([]preference.Trait{})
			},
			req: connect.NewRequest(&frontierv1beta1.DescribePreferencesRequest{}),
			want: connect.NewResponse(&frontierv1beta1.DescribePreferencesResponse{
				Traits: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should handle traits with different input types",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Describe(mock.AnythingOfType("context.backgroundCtx")).Return([]preference.Trait{
					{
						ResourceType: "organization",
						Name:         "text_input",
						Input:        preference.TraitInputText,
					},
					{
						ResourceType: "organization",
						Name:         "select_input",
						Input:        preference.TraitInputSelect,
					},
					{
						ResourceType: "organization",
						Name:         "checkbox_input",
						Input:        preference.TraitInputCheckbox,
					},
				})
			},
			req: connect.NewRequest(&frontierv1beta1.DescribePreferencesRequest{}),
			want: connect.NewResponse(&frontierv1beta1.DescribePreferencesResponse{
				Traits: []*frontierv1beta1.PreferenceTrait{
					{
						ResourceType: "organization",
						Name:         "text_input",
						Input:        &frontierv1beta1.PreferenceTrait_Text{},
					},
					{
						ResourceType: "organization",
						Name:         "select_input",
						Input:        &frontierv1beta1.PreferenceTrait_Select{},
					},
					{
						ResourceType: "organization",
						Name:         "checkbox_input",
						Input:        &frontierv1beta1.PreferenceTrait_Checkbox{},
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			h := &ConnectHandler{preferenceService: mockPreferenceServ}
			got, err := h.DescribePreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestConnectHandler_CreateOrganizationPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *connect.Request[frontierv1beta1.CreateOrganizationPreferencesRequest]
		want    *connect.Response[frontierv1beta1.CreateOrganizationPreferencesResponse]
		wantErr error
	}{
		{
			name: "should create organization preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
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
			req: connect.NewRequest(&frontierv1beta1.CreateOrganizationPreferencesRequest{
				Id: "some_resource_id",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "some_name",
						Value: "some_value",
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateOrganizationPreferencesResponse{
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
			}),
			wantErr: nil,
		},
		{
			name: "should return invalid argument error if trait not found",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "invalid_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.OrganizationNamespace,
				}).Return(preference.Preference{}, preference.ErrTraitNotFound)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateOrganizationPreferencesRequest{
				Id: "some_resource_id",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "invalid_name",
						Value: "some_value",
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, preference.ErrTraitNotFound),
		},
		{
			name: "should return internal error for other service errors",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "some_name",
					Value:        "some_value",
					ResourceID:   "some_resource_id",
					ResourceType: schema.OrganizationNamespace,
				}).Return(preference.Preference{}, errors.New("database error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateOrganizationPreferencesRequest{
				Id: "some_resource_id",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "some_name",
						Value: "some_value",
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should create multiple preferences successfully",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "pref1",
					Value:        "value1",
					ResourceID:   "org_id",
					ResourceType: schema.OrganizationNamespace,
				}).Return(preference.Preference{
					ID:           "id1",
					Name:         "pref1",
					Value:        "value1",
					ResourceID:   "org_id",
					ResourceType: schema.OrganizationNamespace,
				}, nil)
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "pref2",
					Value:        "value2",
					ResourceID:   "org_id",
					ResourceType: schema.OrganizationNamespace,
				}).Return(preference.Preference{
					ID:           "id2",
					Name:         "pref2",
					Value:        "value2",
					ResourceID:   "org_id",
					ResourceType: schema.OrganizationNamespace,
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateOrganizationPreferencesRequest{
				Id: "org_id",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{Name: "pref1", Value: "value1"},
					{Name: "pref2", Value: "value2"},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateOrganizationPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "id1",
						Name:         "pref1",
						Value:        "value1",
						ResourceId:   "org_id",
						ResourceType: schema.OrganizationNamespace,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
					{
						Id:           "id2",
						Name:         "pref2",
						Value:        "value2",
						ResourceId:   "org_id",
						ResourceType: schema.OrganizationNamespace,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			h := &ConnectHandler{preferenceService: mockPreferenceServ}
			got, err := h.CreateOrganizationPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestConnectHandler_ListOrganizationPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *connect.Request[frontierv1beta1.ListOrganizationPreferencesRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationPreferencesResponse]
		wantErr error
	}{
		{
			name: "should list organization preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
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
			req: connect.NewRequest(&frontierv1beta1.ListOrganizationPreferencesRequest{
				Id: "some_id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "some_id",
						Name:         "some_name",
						Value:        "some_value",
						ResourceId:   "some_resource_id",
						ResourceType: "some_resource_type",
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return empty list when no preferences found",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					OrgID: "empty_org",
				}).Return([]preference.Preference{}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListOrganizationPreferencesRequest{
				Id: "empty_org",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationPreferencesResponse{
				Preferences: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should return internal error when service fails",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					OrgID: "error_org",
				}).Return(nil, errors.New("database error"))
			},
			req: connect.NewRequest(&frontierv1beta1.ListOrganizationPreferencesRequest{
				Id: "error_org",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should list multiple preferences successfully",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					OrgID: "multi_org",
				}).Return([]preference.Preference{
					{
						ID:           "pref1",
						Name:         "notification_email",
						Value:        "enabled",
						ResourceID:   "multi_org",
						ResourceType: schema.OrganizationNamespace,
					},
					{
						ID:           "pref2",
						Name:         "theme",
						Value:        "dark",
						ResourceID:   "multi_org",
						ResourceType: schema.OrganizationNamespace,
					},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListOrganizationPreferencesRequest{
				Id: "multi_org",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "pref1",
						Name:         "notification_email",
						Value:        "enabled",
						ResourceId:   "multi_org",
						ResourceType: schema.OrganizationNamespace,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
					{
						Id:           "pref2",
						Name:         "theme",
						Value:        "dark",
						ResourceId:   "multi_org",
						ResourceType: schema.OrganizationNamespace,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			h := &ConnectHandler{preferenceService: mockPreferenceServ}
			got, err := h.ListOrganizationPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestConnectHandler_CreateUserPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *connect.Request[frontierv1beta1.CreateUserPreferencesRequest]
		want    *connect.Response[frontierv1beta1.CreateUserPreferencesResponse]
		wantErr error
	}{
		{
			name: "should create user preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "theme",
					Value:        "dark",
					ResourceID:   "user_id_123",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{
					ID:           "pref_id_123",
					Name:         "theme",
					Value:        "dark",
					ResourceID:   "user_id_123",
					ResourceType: schema.UserPrincipal,
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserPreferencesRequest{
				Id: "user_id_123",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "theme",
						Value: "dark",
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateUserPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "pref_id_123",
						Name:         "theme",
						Value:        "dark",
						ResourceId:   "user_id_123",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return internal error when service fails",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "language",
					Value:        "en",
					ResourceID:   "user_id_456",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{}, errors.New("database connection failed"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserPreferencesRequest{
				Id: "user_id_456",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{
						Name:  "language",
						Value: "en",
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should create multiple user preferences successfully",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "theme",
					Value:        "light",
					ResourceID:   "user_id_789",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{
					ID:           "pref_1",
					Name:         "theme",
					Value:        "light",
					ResourceID:   "user_id_789",
					ResourceType: schema.UserPrincipal,
				}, nil)
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), preference.Preference{
					Name:         "timezone",
					Value:        "UTC",
					ResourceID:   "user_id_789",
					ResourceType: schema.UserPrincipal,
				}).Return(preference.Preference{
					ID:           "pref_2",
					Name:         "timezone",
					Value:        "UTC",
					ResourceID:   "user_id_789",
					ResourceType: schema.UserPrincipal,
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserPreferencesRequest{
				Id: "user_id_789",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{
					{Name: "theme", Value: "light"},
					{Name: "timezone", Value: "UTC"},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateUserPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "pref_1",
						Name:         "theme",
						Value:        "light",
						ResourceId:   "user_id_789",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
					{
						Id:           "pref_2",
						Name:         "timezone",
						Value:        "UTC",
						ResourceId:   "user_id_789",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should handle empty preferences list",
			setup: func(m *mocks.PreferenceService) {
				// No expectations since no preferences to create
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserPreferencesRequest{
				Id:     "user_empty",
				Bodies: []*frontierv1beta1.PreferenceRequestBody{},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateUserPreferencesResponse{
				Preferences: nil,
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			h := &ConnectHandler{preferenceService: mockPreferenceServ}
			got, err := h.CreateUserPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestConnectHandler_ListUserPreferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.PreferenceService)
		req     *connect.Request[frontierv1beta1.ListUserPreferencesRequest]
		want    *connect.Response[frontierv1beta1.ListUserPreferencesResponse]
		wantErr error
	}{
		{
			name: "should list user preferences on success",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					UserID: "user_123",
				}).Return([]preference.Preference{
					{
						ID:           "pref_1",
						Name:         "theme",
						Value:        "dark",
						ResourceID:   "user_123",
						ResourceType: schema.UserPrincipal,
					},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListUserPreferencesRequest{
				Id: "user_123",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListUserPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "pref_1",
						Name:         "theme",
						Value:        "dark",
						ResourceId:   "user_123",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return empty list when no preferences found",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					UserID: "user_empty",
				}).Return([]preference.Preference{}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListUserPreferencesRequest{
				Id: "user_empty",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListUserPreferencesResponse{
				Preferences: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should return internal error when service fails",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					UserID: "user_error",
				}).Return(nil, errors.New("database connection failed"))
			},
			req: connect.NewRequest(&frontierv1beta1.ListUserPreferencesRequest{
				Id: "user_error",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should list multiple user preferences successfully",
			setup: func(m *mocks.PreferenceService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), preference.Filter{
					UserID: "user_multi",
				}).Return([]preference.Preference{
					{
						ID:           "pref_theme",
						Name:         "theme",
						Value:        "light",
						ResourceID:   "user_multi",
						ResourceType: schema.UserPrincipal,
					},
					{
						ID:           "pref_lang",
						Name:         "language",
						Value:        "en",
						ResourceID:   "user_multi",
						ResourceType: schema.UserPrincipal,
					},
					{
						ID:           "pref_tz",
						Name:         "timezone",
						Value:        "UTC",
						ResourceID:   "user_multi",
						ResourceType: schema.UserPrincipal,
					},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListUserPreferencesRequest{
				Id: "user_multi",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListUserPreferencesResponse{
				Preferences: []*frontierv1beta1.Preference{
					{
						Id:           "pref_theme",
						Name:         "theme",
						Value:        "light",
						ResourceId:   "user_multi",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
					{
						Id:           "pref_lang",
						Name:         "language",
						Value:        "en",
						ResourceId:   "user_multi",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
					{
						Id:           "pref_tz",
						Name:         "timezone",
						Value:        "UTC",
						ResourceId:   "user_multi",
						ResourceType: schema.UserPrincipal,
						UpdatedAt:    timestamppb.New(time.Time{}),
						CreatedAt:    timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPreferenceServ := new(mocks.PreferenceService)
			if tt.setup != nil {
				tt.setup(mockPreferenceServ)
			}
			h := &ConnectHandler{preferenceService: mockPreferenceServ}
			got, err := h.ListUserPreferences(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
