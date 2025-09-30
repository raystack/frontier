package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/event"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnectHandler_BillingWebhookCallback(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(es *mocks.EventService)
		request *connect.Request[frontierv1beta1.BillingWebhookCallbackRequest]
		want    *connect.Response[frontierv1beta1.BillingWebhookCallbackResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return invalid argument error when provider is not stripe",
			setup: func(es *mocks.EventService) {
				// No expectations as we return early on provider validation
			},
			request: connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
				Provider: "paypal",
				Body:     []byte("webhook_body"),
			}),
			want:    nil,
			wantErr: ErrBillingProviderNotSupported,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return invalid argument error when provider is empty",
			setup: func(es *mocks.EventService) {
				// No expectations as we return early on provider validation
			},
			request: connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
				Provider: "",
				Body:     []byte("webhook_body"),
			}),
			want:    nil,
			wantErr: ErrBillingProviderNotSupported,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return internal error when event service fails",
			setup: func(es *mocks.EventService) {
				es.On("BillingWebhook", mock.Anything, event.ProviderWebhookEvent{
					Name: "stripe",
					Body: []byte("webhook_body"),
				}).Return(errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
				Provider: "stripe",
				Body:     []byte("webhook_body"),
			}),
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should successfully process stripe webhook",
			setup: func(es *mocks.EventService) {
				es.On("BillingWebhook", mock.Anything, event.ProviderWebhookEvent{
					Name: "stripe",
					Body: []byte("valid_stripe_webhook"),
				}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
				Provider: "stripe",
				Body:     []byte("valid_stripe_webhook"),
			}),
			want: connect.NewResponse(&frontierv1beta1.BillingWebhookCallbackResponse{}),
		},
		{
			name: "should successfully process stripe webhook with empty body",
			setup: func(es *mocks.EventService) {
				es.On("BillingWebhook", mock.Anything, event.ProviderWebhookEvent{
					Name: "stripe",
					Body: []byte(""),
				}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
				Provider: "stripe",
				Body:     []byte(""),
			}),
			want: connect.NewResponse(&frontierv1beta1.BillingWebhookCallbackResponse{}),
		},
		{
			name: "should successfully process stripe webhook with large body",
			setup: func(es *mocks.EventService) {
				largeBody := make([]byte, 10000) // 10KB body
				es.On("BillingWebhook", mock.Anything, event.ProviderWebhookEvent{
					Name: "stripe",
					Body: largeBody,
				}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.BillingWebhookCallbackRequest{
				Provider: "stripe",
				Body:     make([]byte, 10000),
			}),
			want: connect.NewResponse(&frontierv1beta1.BillingWebhookCallbackResponse{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventService := &mocks.EventService{}
			if tt.setup != nil {
				tt.setup(mockEventService)
			}

			handler := &ConnectHandler{
				eventService: mockEventService,
			}

			got, err := handler.BillingWebhookCallback(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.errCode, connect.CodeOf(err))
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockEventService.AssertExpectations(t)
		})
	}
}
