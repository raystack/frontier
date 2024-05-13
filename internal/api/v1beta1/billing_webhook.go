package v1beta1

import (
	"context"
	"go.uber.org/zap"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService interface {
}

func (h Handler) BillingWebhookCallback(ctx context.Context, req *frontierv1beta1.BillingWebhookCallbackRequest) (*frontierv1beta1.BillingWebhookCallbackResponse, error) {
	logger := grpczap.Extract(ctx)
	if req.GetProvider() != "stripe" {
		logger.Error("provider not supported", zap.String("provider", req.GetProvider()))
		return nil, status.Errorf(codes.InvalidArgument, "provider not supported")
	}

	// accept signature from header and pass it downstream
	// TODO
	return &frontierv1beta1.BillingWebhookCallbackResponse{}, nil
}
