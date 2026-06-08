package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) CheckFeatureEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse], error) {
	// Always infer billing_id from org_id
	cust, err := h.customerService.GetByOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{}), nil
		}
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CheckFeatureEntitlement.GetByOrgID: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}

	checkStatus, err := h.entitlementService.Check(ctx, cust.ID, request.Msg.GetFeature())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CheckFeatureEntitlement: billing_id=%s org_id=%s feature=%s: %w", cust.ID, request.Msg.GetOrgId(), request.Msg.GetFeature(), err))
	}

	return connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
		Status: checkStatus,
	}), nil
}

func (h *ConnectHandler) CheckCreditEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckCreditEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckCreditEntitlementResponse], error) {
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CheckCreditEntitlement.List: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}

	if len(customerList) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
	}

	customer := customerList[0]
	customerDetails, err := h.customerService.GetDetails(ctx, customer.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CheckCreditEntitlement.GetDetails: customer_id=%s org_id=%s: %w", customer.ID, request.Msg.GetOrgId(), err))
	}

	creditBalance, err := h.creditService.GetBalance(ctx, customer.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CheckCreditEntitlement.GetBalance: customer_id=%s org_id=%s: %w", customer.ID, request.Msg.GetOrgId(), err))
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
