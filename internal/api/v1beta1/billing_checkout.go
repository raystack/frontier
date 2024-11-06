package v1beta1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"

	"github.com/raystack/frontier/billing/checkout"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func (h Handler) DelegatedCheckout(ctx context.Context, request *frontierv1beta1.DelegatedCheckoutRequest) (*frontierv1beta1.DelegatedCheckoutResponse, error) {
	planID := ""
	var skipTrial bool
	var cancelAfterTrail bool
	var providerCouponID string
	if request.GetSubscriptionBody() != nil {
		planID = request.GetSubscriptionBody().GetPlan()
		skipTrial = request.GetSubscriptionBody().GetSkipTrial()
		cancelAfterTrail = request.GetSubscriptionBody().GetCancelAfterTrial()
		providerCouponID = request.GetSubscriptionBody().GetProviderCouponId()
	}
	productID := ""
	var productQuantity int64
	if request.GetProductBody() != nil {
		productID = request.GetProductBody().GetProduct()
		productQuantity = request.GetProductBody().GetQuantity()
	}
	subs, prod, err := h.checkoutService.Apply(ctx, checkout.Checkout{
		CustomerID:       request.GetBillingId(),
		PlanID:           planID,
		ProductID:        productID,
		Quantity:         productQuantity,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrail,
		ProviderCouponID: providerCouponID,
	})
	if err != nil {
		return nil, err
	}

	var subsPb *frontierv1beta1.Subscription
	if subs != nil {
		if subsPb, err = transformSubscriptionToPB(*subs); err != nil {
			return nil, err
		}
	}
	var productPb *frontierv1beta1.Product
	if prod != nil {
		if productPb, err = transformProductToPB(*prod); err != nil {
			return nil, err
		}
	}

	return &frontierv1beta1.DelegatedCheckoutResponse{
		Subscription: subsPb,
		Product:      productPb,
	}, nil
}

func (h Handler) CreateCheckout(ctx context.Context, request *frontierv1beta1.CreateCheckoutRequest) (*frontierv1beta1.CreateCheckoutResponse, error) {
	// check if setup requested
	if request.GetSetupBody() != nil && request.GetSetupBody().GetPaymentMethod() {
		newCheckout, err := h.checkoutService.CreateSessionForPaymentMethod(ctx, checkout.Checkout{
			CustomerID: request.GetBillingId(),
			SuccessUrl: request.GetSuccessUrl(),
			CancelUrl:  request.GetCancelUrl(),
		})
		if err != nil {
			return nil, err
		}
		return &frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}, nil
	}

	if request.GetSetupBody() != nil && request.GetSetupBody().GetCustomerPortal() {
		newCheckout, err := h.checkoutService.CreateSessionForCustomerPortal(ctx, checkout.Checkout{
			CustomerID: request.GetBillingId(),
			SuccessUrl: request.GetSuccessUrl(),
			CancelUrl:  request.GetCancelUrl(),
		})
		if err != nil {
			return nil, err
		}
		return &frontierv1beta1.CreateCheckoutResponse{
			CheckoutSession: transformCheckoutToPB(newCheckout),
		}, nil
	}

	// check if checkout requested
	planID := ""
	var skipTrial bool
	var cancelAfterTrail bool
	if request.GetSubscriptionBody() != nil {
		planID = request.GetSubscriptionBody().GetPlan()
		skipTrial = request.GetSubscriptionBody().GetSkipTrial()
		cancelAfterTrail = request.GetSubscriptionBody().GetCancelAfterTrial()
	}
	featureID := ""
	if request.GetProductBody() != nil {
		featureID = request.GetProductBody().GetProduct()
	}
	newCheckout, err := h.checkoutService.Create(ctx, checkout.Checkout{
		CustomerID:       request.GetBillingId(),
		SuccessUrl:       request.GetSuccessUrl(),
		CancelUrl:        request.GetCancelUrl(),
		PlanID:           planID,
		ProductID:        featureID,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrail,
	})
	if err != nil {
		if errors.Is(err, product.ErrPerSeatLimitReached) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, err
	}

	return &frontierv1beta1.CreateCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(newCheckout),
	}, nil
}

func (h Handler) ListCheckouts(ctx context.Context, request *frontierv1beta1.ListCheckoutsRequest) (*frontierv1beta1.ListCheckoutsResponse, error) {
	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}
	var checkouts []*frontierv1beta1.CheckoutSession
	checkoutList, err := h.checkoutService.List(ctx, checkout.Filter{
		CustomerID: request.GetBillingId(),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range checkoutList {
		checkouts = append(checkouts, transformCheckoutToPB(v))
	}

	return &frontierv1beta1.ListCheckoutsResponse{
		CheckoutSessions: checkouts,
	}, nil
}

func (h Handler) GetCheckout(ctx context.Context, request *frontierv1beta1.GetCheckoutRequest) (*frontierv1beta1.GetCheckoutResponse, error) {
	if request.GetOrgId() == "" || request.GetId() == "" {
		return nil, grpcBadBodyError
	}
	ch, err := h.GetRawCheckout(ctx, request.GetId())
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(ch),
	}, nil
}

func (h Handler) GetRawCheckout(ctx context.Context, id string) (checkout.Checkout, error) {
	return h.checkoutService.GetByID(ctx, id)
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
