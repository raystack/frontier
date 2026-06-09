package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/product"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.CreateCheckoutRequest]) (*connect.Response[frontierv1beta1.CreateCheckoutResponse], error) {
	// Always infer billing_id from org_id (ignore billing_id from request for security)
	billingID, err := h.GetBillingAccountFromOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateCheckout.GetBillingAccountFromOrgID: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}

	// check if setup requested
	if request.Msg.GetSetupBody() != nil && request.Msg.GetSetupBody().GetPaymentMethod() {
		newCheckout, err := h.checkoutService.CreateSessionForPaymentMethod(ctx, checkout.Checkout{
			CustomerID: billingID,
			SuccessUrl: request.Msg.GetSuccessUrl(),
			CancelUrl:  request.Msg.GetCancelUrl(),
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateCheckout.CreateSessionForPaymentMethod: billing_id=%s: %w", billingID, err))
		}

		return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}), nil
	}

	// check if customer portal requested
	if request.Msg.GetSetupBody() != nil && request.Msg.GetSetupBody().GetCustomerPortal() {
		newCheckout, err := h.checkoutService.CreateSessionForCustomerPortal(ctx, checkout.Checkout{
			CustomerID: billingID,
			SuccessUrl: request.Msg.GetSuccessUrl(),
			CancelUrl:  request.Msg.GetCancelUrl(),
		})
		if err != nil {
			if errors.Is(err, checkout.ErrKycCompleted) {
				return nil, connect.NewError(connect.CodeFailedPrecondition, ErrPortalChangesKycCompleted)
			}
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateCheckout.CreateSessionForCustomerPortal: billing_id=%s: %w", billingID, err))
		}

		return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}), nil
	}

	// check if checkout requested (subscription or product)
	if request.Msg.GetSubscriptionBody() == nil && request.Msg.GetProductBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	planID := ""
	var skipTrial bool
	var cancelAfterTrial bool
	if request.Msg.GetSubscriptionBody() != nil {
		planID = request.Msg.GetSubscriptionBody().GetPlan()
		skipTrial = request.Msg.GetSubscriptionBody().GetSkipTrial()
		cancelAfterTrial = request.Msg.GetSubscriptionBody().GetCancelAfterTrial()
	}

	var featureID string
	var quantity int64
	if request.Msg.GetProductBody() != nil {
		featureID = request.Msg.GetProductBody().GetProduct()
		quantity = request.Msg.GetProductBody().GetQuantity()
	}
	newCheckout, err := h.checkoutService.Create(ctx, checkout.Checkout{
		CustomerID:       billingID,
		SuccessUrl:       request.Msg.GetSuccessUrl(),
		CancelUrl:        request.Msg.GetCancelUrl(),
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateCheckout.Create: billing_id=%s plan_id=%s product_id=%s quantity=%d skip_trial=%v cancel_after_trial=%v: %w", billingID, planID, featureID, quantity, skipTrial, cancelAfterTrial, err))
	}

	return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(newCheckout),
	}), nil
}

func (h *ConnectHandler) DelegatedCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.DelegatedCheckoutRequest]) (*connect.Response[frontierv1beta1.DelegatedCheckoutResponse], error) {
	// Always infer billing_id from org_id (ignore billing_id from request for security)
	billingID, err := h.GetBillingAccountFromOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DelegatedCheckout.GetBillingAccountFromOrgID: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}

	var planID string
	var skipTrial bool
	var cancelAfterTrail bool
	var providerCouponID string
	if request.Msg.GetSubscriptionBody() != nil {
		planID = request.Msg.GetSubscriptionBody().GetPlan()
		skipTrial = request.Msg.GetSubscriptionBody().GetSkipTrial()
		cancelAfterTrail = request.Msg.GetSubscriptionBody().GetCancelAfterTrial()
		providerCouponID = request.Msg.GetSubscriptionBody().GetProviderCouponId()
	}
	var productID string
	var productQuantity int64
	if request.Msg.GetProductBody() != nil {
		productID = request.Msg.GetProductBody().GetProduct()
		productQuantity = request.Msg.GetProductBody().GetQuantity()
	}
	subs, prod, err := h.checkoutService.Apply(ctx, checkout.Checkout{
		CustomerID:       billingID,
		PlanID:           planID,
		ProductID:        productID,
		Quantity:         productQuantity,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrail,
		ProviderCouponID: providerCouponID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DelegatedCheckout.Apply: billing_id=%s plan_id=%s product_id=%s product_quantity=%d skip_trial=%v cancel_after_trial=%v provider_coupon_id=%s: %w", billingID, planID, productID, productQuantity, skipTrial, cancelAfterTrail, providerCouponID, err))
	}

	var subsPb *frontierv1beta1.Subscription
	if subs != nil {
		if subsPb, err = transformSubscriptionToPB(*subs); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DelegatedCheckout: subscription_id=%s: %w", subs.ID, err))
		}
	}
	var productPb *frontierv1beta1.Product
	if prod != nil {
		if productPb, err = transformProductToPB(*prod); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DelegatedCheckout: product_id=%s: %w", prod.ID, err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.DelegatedCheckoutResponse{
		Subscription: subsPb,
		Product:      productPb,
	}), nil
}

func (h *ConnectHandler) ListCheckouts(ctx context.Context, request *connect.Request[frontierv1beta1.ListCheckoutsRequest]) (*connect.Response[frontierv1beta1.ListCheckoutsResponse], error) {
	if request.Msg.GetOrgId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	// Always infer billing_id from org_id (ignore billing_id from request for security)
	billingID, err := h.GetBillingAccountFromOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListCheckouts.GetBillingAccountFromOrgID: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}

	var checkouts []*frontierv1beta1.CheckoutSession
	checkoutList, err := h.checkoutService.List(ctx, checkout.Filter{
		CustomerID: billingID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListCheckouts.List: billing_id=%s org_id=%s: %w", billingID, request.Msg.GetOrgId(), err))
	}
	for _, v := range checkoutList {
		checkouts = append(checkouts, transformCheckoutToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.ListCheckoutsResponse{
		CheckoutSessions: checkouts,
	}), nil
}

func (h *ConnectHandler) GetCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.GetCheckoutRequest]) (*connect.Response[frontierv1beta1.GetCheckoutResponse], error) {
	if request.Msg.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	ch, err := h.checkoutService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetCheckout.GetByID: checkout_id=%s: %w", request.Msg.GetId(), err))
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
