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
	// check if setup requested
	if request.Msg.GetSetupBody() != nil && request.Msg.GetSetupBody().GetPaymentMethod() {
		newCheckout, err := h.checkoutService.CreateSessionForPaymentMethod(ctx, checkout.Checkout{
			CustomerID: request.Msg.GetBillingId(),
			SuccessUrl: request.Msg.GetSuccessUrl(),
			CancelUrl:  request.Msg.GetCancelUrl(),
		})
		if err != nil {
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(newCheckout),
	}), nil
}

func (h *ConnectHandler) DelegatedCheckout(ctx context.Context, request *connect.Request[frontierv1beta1.DelegatedCheckoutRequest]) (*connect.Response[frontierv1beta1.DelegatedCheckoutResponse], error) {
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
