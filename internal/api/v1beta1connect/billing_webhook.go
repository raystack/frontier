package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/event"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type EventService interface {
	BillingWebhook(ctx context.Context, event event.ProviderWebhookEvent) error
}

func (h *ConnectHandler) BillingWebhookCallback(ctx context.Context, request *connect.Request[frontierv1beta1.BillingWebhookCallbackRequest]) (*connect.Response[frontierv1beta1.BillingWebhookCallbackResponse], error) {
	logger := grpczap.Extract(ctx)
	if request.Msg.GetProvider() != "stripe" {
		logger.Error("provider not supported", zap.String("provider", request.Msg.GetProvider()))
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBillingProviderNotSupported)
	}

	if err := h.eventService.BillingWebhook(ctx, event.ProviderWebhookEvent{
		Name: request.Msg.GetProvider(),
		Body: request.Msg.GetBody(),
	}); err != nil {
		logger.Error("failed to process billing webhook", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to process billing webhook: %w", err))
	}
	return connect.NewResponse(&frontierv1beta1.BillingWebhookCallbackResponse{}), nil
}
