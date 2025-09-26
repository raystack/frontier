package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type EntitlementService interface {
	Check(ctx context.Context, customerID, featureID string) (bool, error)
	CheckPlanEligibility(ctx context.Context, customerID string) error
}

func (h *ConnectHandler) CheckFeatureEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse], error) {
	checkStatus, err := h.entitlementService.Check(ctx, request.Msg.GetBillingId(), request.Msg.GetFeature())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
		Status: checkStatus,
	}), nil
}
