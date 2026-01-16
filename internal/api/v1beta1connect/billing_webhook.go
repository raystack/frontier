package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/event"
	"github.com/raystack/frontier/pkg/server/consts"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
)

func (h *ConnectHandler) BillingWebhookCallback(ctx context.Context, request *connect.Request[frontierv1beta1.BillingWebhookCallbackRequest]) (*connect.Response[frontierv1beta1.BillingWebhookCallbackResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.GetProvider() != "stripe" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBillingProviderNotSupported)
	}

	// Extract Stripe webhook signature from headers and add to context
	// This is required for webhook signature verification in the event service
	if webhookSignature := request.Header().Get(consts.StripeWebhookSignature); webhookSignature != "" {
		ctx = customer.SetStripeWebhookSignatureInContext(ctx, webhookSignature)
	}

	if err := h.eventService.BillingWebhook(ctx, event.ProviderWebhookEvent{
		Name: request.Msg.GetProvider(),
		Body: request.Msg.GetBody(),
	}); err != nil {
		errorLogger.LogServiceError(ctx, request, "BillingWebhookCallback.BillingWebhook", err,
			zap.String("provider", request.Msg.GetProvider()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.BillingWebhookCallbackResponse{}), nil
}
