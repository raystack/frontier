package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CheckoutService interface {
	Create(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	GetByID(ctx context.Context, id string) (checkout.Checkout, error)
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
	Apply(ctx context.Context, ch checkout.Checkout) (*subscription.Subscription, *product.Product, error)
	CreateSessionForPaymentMethod(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	CreateSessionForCustomerPortal(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
}

func (h *ConnectHandler) CreateCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.CreateCheckoutRequest]) (*connect.Response[frontierv1beta1.CreateCheckoutResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichCreateCheckoutRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	// check if setup requested
	if enrichedReq.GetSetupBody() != nil && enrichedReq.GetSetupBody().GetPaymentMethod() {
		newCheckout, err := h.checkoutService.CreateSessionForPaymentMethod(ctx, checkout.Checkout{
			CustomerID: enrichedReq.GetBillingId(),
			SuccessUrl: enrichedReq.GetSuccessUrl(),
			CancelUrl:  enrichedReq.GetCancelUrl(),
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}), nil
	}

	// check if customer portal requested
	if enrichedReq.GetSetupBody() != nil && enrichedReq.GetSetupBody().GetCustomerPortal() {
		newCheckout, err := h.checkoutService.CreateSessionForCustomerPortal(ctx, checkout.Checkout{
			CustomerID: enrichedReq.GetBillingId(),
			SuccessUrl: enrichedReq.GetSuccessUrl(),
			CancelUrl:  enrichedReq.GetCancelUrl(),
		})
		if err != nil {
			if errors.Is(err, checkout.ErrKycCompleted) {
				return nil, connect.NewError(connect.CodeFailedPrecondition, ErrPortalChangesKycCompleted)
			}
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}), nil
	}

	// check if checkout requested (subscription or product)
	if enrichedReq.GetSubscriptionBody() == nil && enrichedReq.GetProductBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	planID := ""
	var skipTrial bool
	var cancelAfterTrial bool
	if enrichedReq.GetSubscriptionBody() != nil {
		planID = enrichedReq.GetSubscriptionBody().GetPlan()
		skipTrial = enrichedReq.GetSubscriptionBody().GetSkipTrial()
		cancelAfterTrial = enrichedReq.GetSubscriptionBody().GetCancelAfterTrial()
	}

	var featureID string
	var quantity int64
	if enrichedReq.GetProductBody() != nil {
		featureID = enrichedReq.GetProductBody().GetProduct()
		quantity = enrichedReq.GetProductBody().GetQuantity()
	}
	newCheckout, err := h.checkoutService.Create(ctx, checkout.Checkout{
		CustomerID:       enrichedReq.GetBillingId(),
		SuccessUrl:       enrichedReq.GetSuccessUrl(),
		CancelUrl:        enrichedReq.GetCancelUrl(),
		PlanID:           planID,
		ProductID:        featureID,
		Quantity:         quantity,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrial,
	})
	if err != nil {
		if errors.Is(err, product.ErrPerSeatLimitReached) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrPerSeatLimitReached)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(newCheckout),
	}), nil
}

func (h *ConnectHandler) DelegatedCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.DelegatedCheckoutRequest]) (*connect.Response[frontierv1beta1.DelegatedCheckoutResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichDelegatedCheckoutRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	var planID string
	var skipTrial bool
	var cancelAfterTrail bool
	var providerCouponID string
	if enrichedReq.GetSubscriptionBody() != nil {
		planID = enrichedReq.GetSubscriptionBody().GetPlan()
		skipTrial = enrichedReq.GetSubscriptionBody().GetSkipTrial()
		cancelAfterTrail = enrichedReq.GetSubscriptionBody().GetCancelAfterTrial()
		providerCouponID = enrichedReq.GetSubscriptionBody().GetProviderCouponId()
	}
	var productID string
	var productQuantity int64
	if enrichedReq.GetProductBody() != nil {
		productID = enrichedReq.GetProductBody().GetProduct()
		productQuantity = enrichedReq.GetProductBody().GetQuantity()
	}
	subs, prod, err := h.checkoutService.Apply(ctx, checkout.Checkout{
		CustomerID:       enrichedReq.GetBillingId(),
		PlanID:           planID,
		ProductID:        productID,
		Quantity:         productQuantity,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrail,
		ProviderCouponID: providerCouponID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var subsPb *frontierv1beta1.Subscription
	if subs != nil {
		if subsPb, err = transformSubscriptionToPB(*subs); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	var productPb *frontierv1beta1.Product
	if prod != nil {
		if productPb, err = transformProductToPB(*prod); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DelegatedCheckoutResponse{
		Subscription: subsPb,
		Product:      productPb,
	}), nil
}

func (h *ConnectHandler) ListCheckouts(ctx context.Context, request *connect.Request[frontierv1beta1.ListCheckoutsRequest]) (*connect.Response[frontierv1beta1.ListCheckoutsResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichListCheckoutsRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	var checkouts []*frontierv1beta1.CheckoutSession
	checkoutList, err := h.checkoutService.List(ctx, checkout.Filter{
		CustomerID: enrichedReq.GetBillingId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range checkoutList {
		checkouts = append(checkouts, transformCheckoutToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.ListCheckoutsResponse{
		CheckoutSessions: checkouts,
	}), nil
}

func (h *ConnectHandler) GetCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.GetCheckoutRequest]) (*connect.Response[frontierv1beta1.GetCheckoutResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichGetCheckoutRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	if enrichedReq.GetOrgId() == "" || enrichedReq.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	ch, err := h.checkoutService.GetByID(ctx, enrichedReq.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(ch),
	}), nil
}

func transformCheckoutToPB(ch checkout.Checkout) *frontierv1beta1.CheckoutSession {
	return &frontierv1beta1.CheckoutSession{
		Id:          ch.ID,
		CheckoutUrl: ch.CheckoutUrl,
		SuccessUrl:  ch.SuccessUrl,
		CancelUrl:   ch.CancelUrl,
		State:       ch.State,
		Plan:        ch.PlanID,
		Product:     ch.ProductID,
		CreatedAt:   timestamppb.New(ch.CreatedAt),
		UpdatedAt:   timestamppb.New(ch.UpdatedAt),
		ExpireAt:    timestamppb.New(ch.ExpireAt),
	}
}

// enrichCreateCheckoutRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichCreateCheckoutRequest(ctx context.Context, req *frontierv1beta1.CreateCheckoutRequest) (*frontierv1beta1.CreateCheckoutRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.CreateCheckoutRequest{
		BillingId:        req.GetBillingId(),
		OrgId:            req.GetOrgId(),
		SuccessUrl:       req.GetSuccessUrl(),
		CancelUrl:        req.GetCancelUrl(),
		SubscriptionBody: req.GetSubscriptionBody(),
		ProductBody:      req.GetProductBody(),
		SetupBody:        req.GetSetupBody(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichDelegatedCheckoutRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichDelegatedCheckoutRequest(ctx context.Context, req *frontierv1beta1.DelegatedCheckoutRequest) (*frontierv1beta1.DelegatedCheckoutRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.DelegatedCheckoutRequest{
		BillingId:        req.GetBillingId(),
		OrgId:            req.GetOrgId(),
		SubscriptionBody: req.GetSubscriptionBody(),
		ProductBody:      req.GetProductBody(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichListCheckoutsRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichListCheckoutsRequest(ctx context.Context, req *frontierv1beta1.ListCheckoutsRequest) (*frontierv1beta1.ListCheckoutsRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.ListCheckoutsRequest{
		BillingId: req.GetBillingId(),
		OrgId:     req.GetOrgId(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichGetCheckoutRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichGetCheckoutRequest(ctx context.Context, req *frontierv1beta1.GetCheckoutRequest) (*frontierv1beta1.GetCheckoutRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.GetCheckoutRequest{
		BillingId: req.GetBillingId(),
		OrgId:     req.GetOrgId(),
		Id:        req.GetId(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}
