package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/event"
	"go.uber.org/zap"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService interface {
	BillingWebhook(ctx context.Context, event event.ProviderWebhookEvent) error
}

func (h Handler) BillingWebhookCallback(ctx context.Context, req *frontierv1beta1.BillingWebhookCallbackRequest) (*frontierv1beta1.BillingWebhookCallbackResponse, error) {
	logger := grpczap.Extract(ctx)
	if req.GetProvider() != "stripe" {
		logger.Error("provider not supported", zap.String("provider", req.GetProvider()))
		return nil, status.Errorf(codes.InvalidArgument, "provider not supported")
	}

	if err := h.eventService.BillingWebhook(ctx, event.ProviderWebhookEvent{
		Name: req.GetProvider(),
		Body: req.GetBody(),
	}); err != nil {
		logger.Error("failed to process billing webhook", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to process billing webhook: %v", err)
	}
	return &frontierv1beta1.BillingWebhookCallbackResponse{}, nil
}
