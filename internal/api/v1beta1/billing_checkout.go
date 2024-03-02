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

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type CheckoutService interface {
	Create(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
	Apply(ctx context.Context, ch checkout.Checkout) (*subscription.Subscription, *product.Product, error)
	CreateSessionForPaymentMethod(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
}

func (h Handler) DelegatedCheckout(ctx context.Context, request *frontierv1beta1.DelegatedCheckoutRequest) (*frontierv1beta1.DelegatedCheckoutResponse, error) {
	logger := grpczap.Extract(ctx)

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
	subs, prod, err := h.checkoutService.Apply(ctx, checkout.Checkout{
		CustomerID:       request.GetBillingId(),
		PlanID:           planID,
		ProductID:        featureID,
		SkipTrial:        skipTrial,
		CancelAfterTrial: cancelAfterTrail,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var subsPb *frontierv1beta1.Subscription
	if subs != nil {
		if subsPb, err = transformSubscriptionToPB(*subs); err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
	}
	var productPb *frontierv1beta1.Product
	if prod != nil {
		if productPb, err = transformProductToPB(*prod); err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
	}

	return &frontierv1beta1.DelegatedCheckoutResponse{
		Subscription: subsPb,
		Product:      productPb,
	}, nil
}

func (h Handler) CreateCheckout(ctx context.Context, request *frontierv1beta1.CreateCheckoutRequest) (*frontierv1beta1.CreateCheckoutResponse, error) {
	logger := grpczap.Extract(ctx)

	// check if setup requested
	if request.GetSetupBody() != nil && request.GetSetupBody().GetPaymentMethod() {
		newCheckout, err := h.checkoutService.CreateSessionForPaymentMethod(ctx, checkout.Checkout{
			CustomerID: request.GetBillingId(),
			SuccessUrl: request.GetSuccessUrl(),
			CancelUrl:  request.GetCancelUrl(),
		})
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
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
		logger.Error(err.Error())
		if errors.Is(err, product.ErrPerSeatLimitReached) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.CreateCheckoutResponse{
		CheckoutSession: transformCheckoutToPB(newCheckout),
	}, nil
}

func (h Handler) ListCheckouts(ctx context.Context, request *frontierv1beta1.ListCheckoutsRequest) (*frontierv1beta1.ListCheckoutsResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}
	var checkouts []*frontierv1beta1.CheckoutSession
	checkoutList, err := h.checkoutService.List(ctx, checkout.Filter{
		CustomerID: request.GetBillingId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range checkoutList {
		checkouts = append(checkouts, transformCheckoutToPB(v))
	}

	return &frontierv1beta1.ListCheckoutsResponse{
		CheckoutSessions: checkouts,
	}, nil
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
