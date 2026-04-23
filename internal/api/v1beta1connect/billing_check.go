package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) CheckFeatureEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse], error) {
	errorLogger := NewErrorLogger()

	// Always infer billing_id from org_id
	cust, err := h.customerService.GetByOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{}), nil
		}
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		errorLogger.LogServiceError(ctx, request, "CheckFeatureEntitlement.GetByOrgID", err,
			"org_id", request.Msg.GetOrgId())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	checkStatus, err := h.entitlementService.Check(ctx, cust.ID, request.Msg.GetFeature())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckFeatureEntitlement", err,
			"billing_id", cust.ID,
			"org_id", request.Msg.GetOrgId(),
			"feature", request.Msg.GetFeature())
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
			"org_id", request.Msg.GetOrgId())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if len(customerList) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
	}

	customer := customerList[0]
	customerDetails, err := h.customerService.GetDetails(ctx, customer.ID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckCreditEntitlement.GetDetails", err,
			"customer_id", customer.ID,
			"org_id", request.Msg.GetOrgId())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	creditBalance, err := h.creditService.GetBalance(ctx, customer.ID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckCreditEntitlement.GetBalance", err,
			"customer_id", customer.ID,
			"org_id", request.Msg.GetOrgId())
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
