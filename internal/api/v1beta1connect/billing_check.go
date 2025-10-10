package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type EntitlementService interface {
	Check(ctx context.Context, customerID, featureID string) (bool, error)
	CheckPlanEligibility(ctx context.Context, customerID string) error
}

func (h *ConnectHandler) CheckFeatureEntitlement(ctx context.Context, request *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichCheckFeatureEntitlementRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	checkStatus, err := h.entitlementService.Check(ctx, enrichedReq.GetBillingId(), enrichedReq.GetFeature())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if len(customerList) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
	}

	customer := customerList[0]
	customerDetails, err := h.customerService.GetDetails(ctx, customer.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	creditBalance, err := h.creditService.GetBalance(ctx, customer.ID)
	if err != nil {
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

// enrichCheckFeatureEntitlementRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichCheckFeatureEntitlementRequest(ctx context.Context, req *frontierv1beta1.CheckFeatureEntitlementRequest) (*frontierv1beta1.CheckFeatureEntitlementRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.CheckFeatureEntitlementRequest{
		ProjectId: req.GetProjectId(),
		OrgId:     req.GetOrgId(),
		BillingId: req.GetBillingId(),
		Feature:   req.GetFeature(),
	}

	// Step 1: Convert project ID to org ID if needed
	if enrichedReq.GetProjectId() != "" && enrichedReq.GetOrgId() != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrStatusOrgProjectMismatch)
	}

	if enrichedReq.GetProjectId() != "" {
		proj, err := h.GetProject(ctx, connect.NewRequest(&frontierv1beta1.GetProjectRequest{
			Id: enrichedReq.GetProjectId(),
		}))
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		if proj != nil && proj.Msg != nil && proj.Msg.GetProject() != nil {
			enrichedReq.OrgId = proj.Msg.GetProject().GetOrgId()
		}
	}

	// Step 2: Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		// Find default customer id for the org
		customers, err := h.customerService.List(ctx, customer.Filter{
			OrgID: enrichedReq.GetOrgId(),
			State: customer.ActiveState,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		if len(customers) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrCustomerNotFound)
		}
		enrichedReq.BillingId = customers[0].ID
	}

	return enrichedReq, nil
}
