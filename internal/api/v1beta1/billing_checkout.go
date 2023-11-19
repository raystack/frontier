package v1beta1

import (
	"context"

	"github.com/raystack/frontier/billing/checkout"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type CheckoutService interface {
	Create(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
}

func (h Handler) CreateCheckout(ctx context.Context, request *frontierv1beta1.CreateCheckoutRequest) (*frontierv1beta1.CreateCheckoutResponse, error) {
	logger := grpczap.Extract(ctx)

	planID := ""
	if request.GetSubscriptionBody() != nil {
		planID = request.GetSubscriptionBody().GetPlan()
	}
	featureID := ""
	if request.GetFeatureBody() != nil {
		featureID = request.GetFeatureBody().GetFeature()
	}
	newCheckout, err := h.checkoutService.Create(ctx, checkout.Checkout{
		CustomerID: request.GetBillingId(),
		SuccessUrl: request.GetSuccessUrl(),
		CancelUrl:  request.GetCancelUrl(),
		PlanID:     planID,
		FeatureID:  featureID,
	})
	if err != nil {
		logger.Error(err.Error())
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
		CreatedAt:   timestamppb.New(ch.CreatedAt),
		UpdatedAt:   timestamppb.New(ch.UpdatedAt),
		ExpireAt:    timestamppb.New(ch.ExpireAt),
	}
}
