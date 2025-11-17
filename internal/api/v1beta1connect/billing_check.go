package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
)

func (h *ConnectHandler) CheckFeatureEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse], error) {
	errorLogger := NewErrorLogger()

	checkStatus, err := h.entitlementService.Check(ctx, request.Msg.GetBillingId(), request.Msg.GetFeature())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckFeatureEntitlement", err,
			zap.String("billing_id", request.Msg.GetBillingId()),
			zap.String("feature", request.Msg.GetFeature()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
		Status: checkStatus,
	}), nil
}

func (h *ConnectHandler) CheckCreditEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckCreditEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckCreditEntitlementResponse], error) {
	errorLogger := NewErrorLogger()

	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckCreditEntitlement.List", err,
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if len(customerList) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
	}

	customer := customerList[0]
	customerDetails, err := h.customerService.GetDetails(ctx, customer.ID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckCreditEntitlement.GetDetails", err,
			zap.String("customer_id", customer.ID),
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	creditBalance, err := h.creditService.GetBalance(ctx, customer.ID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckCreditEntitlement.GetBalance", err,
			zap.String("customer_id", customer.ID),
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if creditBalance-request.Msg.GetAmount() >= customerDetails.CreditMin {
		return connect.NewResponse(&frontierv1beta1.CheckCreditEntitlementResponse{
			Status: true,
		}), nil
	}

	return connect.NewResponse(&frontierv1beta1.CheckCreditEntitlementResponse{
		Status: false,
	}), nil
}
