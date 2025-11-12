package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/product"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.CreateCheckoutRequest]) (*connect.Response[frontierv1beta1.CreateCheckoutResponse], error) {
	errorLogger := NewErrorLogger()

	// check if setup requested
	if request.Msg.GetSetupBody() != nil && request.Msg.GetSetupBody().GetPaymentMethod() {
		newCheckout, err := h.checkoutService.CreateSessionForPaymentMethod(ctx, checkout.Checkout{
			CustomerID: request.Msg.GetBillingId(),
			SuccessUrl: request.Msg.GetSuccessUrl(),
			CancelUrl:  request.Msg.GetCancelUrl(),
		})
		if err != nil {
			errorLogger.LogServiceError(ctx, request, "CreateCheckout.CreateSessionForPaymentMethod", err,
				zap.String("billing_id", request.Msg.GetBillingId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}), nil
	}

	// check if customer portal requested
	if request.Msg.GetSetupBody() != nil && request.Msg.GetSetupBody().GetCustomerPortal() {
		newCheckout, err := h.checkoutService.CreateSessionForCustomerPortal(ctx, checkout.Checkout{
			CustomerID: request.Msg.GetBillingId(),
			SuccessUrl: request.Msg.GetSuccessUrl(),
			CancelUrl:  request.Msg.GetCancelUrl(),
		})
		if err != nil {
			if errors.Is(err, checkout.ErrKycCompleted) {
				return nil, connect.NewError(connect.CodeFailedPrecondition, ErrPortalChangesKycCompleted)
			}
			errorLogger.LogServiceError(ctx, request, "CreateCheckout.CreateSessionForCustomerPortal", err,
				zap.String("billing_id", request.Msg.GetBillingId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
		CustomerID:       request.Msg.GetBillingId(),
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
		errorLogger.LogServiceError(ctx, request, "CreateCheckout.Create", err,
			zap.String("billing_id", request.Msg.GetBillingId()),
			zap.String("plan_id", planID),
			zap.String("product_id", featureID),
			zap.Int64("quantity", quantity),
			zap.Bool("skip_trial", skipTrial),
			zap.Bool("cancel_after_trial", cancelAfterTrial))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(newCheckout),
	}), nil
}

func (h *ConnectHandler) DelegatedCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.DelegatedCheckoutRequest]) (*connect.Response[frontierv1beta1.DelegatedCheckoutResponse], error) {
	errorLogger := NewErrorLogger()

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
		CustomerID:       request.Msg.GetBillingId(),
		PlanID:           planID,
		ProductID:        productID,
		Quantity:         productQuantity,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrail,
		ProviderCouponID: providerCouponID,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DelegatedCheckout.Apply", err,
			zap.String("billing_id", request.Msg.GetBillingId()),
			zap.String("plan_id", planID),
			zap.String("product_id", productID),
			zap.Int64("product_quantity", productQuantity),
			zap.Bool("skip_trial", skipTrial),
			zap.Bool("cancel_after_trial", cancelAfterTrail),
			zap.String("provider_coupon_id", providerCouponID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var subsPb *frontierv1beta1.Subscription
	if subs != nil {
		if subsPb, err = transformSubscriptionToPB(*subs); err != nil {
			errorLogger.LogTransformError(ctx, request, "DelegatedCheckout", subs.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	var productPb *frontierv1beta1.Product
	if prod != nil {
		if productPb, err = transformProductToPB(*prod); err != nil {
			errorLogger.LogTransformError(ctx, request, "DelegatedCheckout", prod.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DelegatedCheckoutResponse{
		Subscription: subsPb,
		Product:      productPb,
	}), nil
}

func (h *ConnectHandler) ListCheckouts(ctx context.Context, request *connect.Request[frontierv1beta1.ListCheckoutsRequest]) (*connect.Response[frontierv1beta1.ListCheckoutsResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.GetOrgId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	var checkouts []*frontierv1beta1.CheckoutSession
	checkoutList, err := h.checkoutService.List(ctx, checkout.Filter{
		CustomerID: request.Msg.GetBillingId(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListCheckouts.List", err,
			zap.String("billing_id", request.Msg.GetBillingId()),
			zap.String("org_id", request.Msg.GetOrgId()))
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
	errorLogger := NewErrorLogger()

	if request.Msg.GetOrgId() == "" || request.Msg.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	ch, err := h.checkoutService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetCheckout.GetByID", err,
			zap.String("checkout_id", request.Msg.GetId()),
			zap.String("org_id", request.Msg.GetOrgId()))
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
