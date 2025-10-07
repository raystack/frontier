package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
