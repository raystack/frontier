package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type CheckoutService interface {
	Create(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	GetByID(ctx context.Context, id string) (checkout.Checkout, error)
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
	Apply(ctx context.Context, ch checkout.Checkout) (*subscription.Subscription, *product.Product, error)
	CreateSessionForPaymentMethod(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	CreateSessionForCustomerPortal(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
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
