package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectHandler_ListConsentDocuments(t *testing.T) {
	tests := []struct {
		name  string
		setup func(consentSrv *mocks.ConsentService)
		want  []*frontierv1beta1.ConsentDocument
	}{
		{
			name: "should return every configured document with all four fields",
			setup: func(consentSrv *mocks.ConsentService) {
				consentSrv.EXPECT().Documents().Return([]consent.Document{
					{
						ID:      "eula",
						Title:   "End User License Agreement",
						Version: "2026-02-14",
						URL:     "https://example.org/legal/eula/2026-02-14",
					},
					{
						ID:      "privacy_policy",
						Title:   "Privacy Policy",
						Version: "2026-04-01",
						URL:     "https://example.org/legal/privacy/2026-04-01",
					},
					{
						ID:      "terms_of_service",
						Title:   "Terms & Conditions",
						Version: "2026-04-01",
						URL:     "https://example.org/legal/terms/2026-04-01",
					},
				})
			},
			want: []*frontierv1beta1.ConsentDocument{
				{
					Id:      "eula",
					Title:   "End User License Agreement",
					Version: "2026-02-14",
					Url:     "https://example.org/legal/eula/2026-02-14",
				},
				{
					Id:      "privacy_policy",
					Title:   "Privacy Policy",
					Version: "2026-04-01",
					Url:     "https://example.org/legal/privacy/2026-04-01",
				},
				{
					Id:      "terms_of_service",
					Title:   "Terms & Conditions",
					Version: "2026-04-01",
					Url:     "https://example.org/legal/terms/2026-04-01",
				},
			},
		},
		{
			// no documents means no checkbox, so one client build works
			// against a deployment that asks for consent and one that does not
			name: "should return an empty list and no error when consent is disabled",
			setup: func(consentSrv *mocks.ConsentService) {
				consentSrv.EXPECT().Documents().Return(nil)
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConsentSrv := new(mocks.ConsentService)
			if tt.setup != nil {
				tt.setup(mockConsentSrv)
			}

			handler := &ConnectHandler{consentService: mockConsentSrv}

			resp, err := handler.ListConsentDocuments(context.Background(),
				connect.NewRequest(&frontierv1beta1.ListConsentDocumentsRequest{}))

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Msg.GetDocuments(), len(tt.want))
			for i, want := range tt.want {
				got := resp.Msg.GetDocuments()[i]
				assert.Equal(t, want.GetId(), got.GetId())
				assert.Equal(t, want.GetTitle(), got.GetTitle())
				assert.Equal(t, want.GetVersion(), got.GetVersion())
				assert.Equal(t, want.GetUrl(), got.GetUrl())
			}
			mockConsentSrv.AssertExpectations(t)
		})
	}
}

// TestConnectHandler_ListConsentDocuments_OrderIsTheServiceOrder pins that the
// handler passes the service's ordering through untouched. The service orders
// by id, which is what makes the response stable across restarts.
func TestConnectHandler_ListConsentDocuments_OrderIsTheServiceOrder(t *testing.T) {
	mockConsentSrv := new(mocks.ConsentService)
	mockConsentSrv.EXPECT().Documents().Return([]consent.Document{
		{ID: "eula", Title: "EULA", Version: "1", URL: "https://example.org/eula"},
		{ID: "privacy_policy", Title: "Privacy", Version: "1", URL: "https://example.org/privacy"},
		{ID: "terms_of_service", Title: "Terms", Version: "1", URL: "https://example.org/terms"},
	})

	handler := &ConnectHandler{consentService: mockConsentSrv}
	resp, err := handler.ListConsentDocuments(context.Background(),
		connect.NewRequest(&frontierv1beta1.ListConsentDocumentsRequest{}))

	require.NoError(t, err)
	ids := make([]string, 0, len(resp.Msg.GetDocuments()))
	for _, document := range resp.Msg.GetDocuments() {
		ids = append(ids, document.GetId())
	}
	assert.Equal(t, []string{"eula", "privacy_policy", "terms_of_service"}, ids)
}
