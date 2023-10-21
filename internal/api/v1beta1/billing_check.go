package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type EntitlementService interface {
	Check(ctx context.Context, customerID, featureID string) (bool, error)
}

func (h Handler) CheckFeatureEntitlement(ctx context.Context, request *frontierv1beta1.CheckFeatureEntitlementRequest) (*frontierv1beta1.CheckFeatureEntitlementResponse, error) {
	logger := grpczap.Extract(ctx)

	checkStatus, err := h.entitlementService.Check(ctx, request.GetCustomerId(), request.GetFeature())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.CheckFeatureEntitlementResponse{
		Status: checkStatus,
	}, nil
}
