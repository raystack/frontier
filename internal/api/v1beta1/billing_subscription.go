package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/subscription"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SubscriptionService interface {
	GetByID(ctx context.Context, id string) (subscription.Subscription, error)
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
	Cancel(ctx context.Context, id string) (subscription.Subscription, error)
}

func (h Handler) ListSubscriptions(ctx context.Context, request *frontierv1beta1.ListSubscriptionsRequest) (*frontierv1beta1.ListSubscriptionsResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}
	var subscriptions []*frontierv1beta1.Subscription
	subscriptionList, err := h.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: request.GetBillingId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range subscriptionList {
		subscriptionPB, err := transformSubscriptionToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		subscriptions = append(subscriptions, subscriptionPB)
	}

	return &frontierv1beta1.ListSubscriptionsResponse{
		Subscriptions: subscriptions,
	}, nil
}

func (h Handler) GetSubscription(ctx context.Context, request *frontierv1beta1.GetSubscriptionRequest) (*frontierv1beta1.GetSubscriptionResponse, error) {
	logger := grpczap.Extract(ctx)

	subscription, err := h.subscriptionService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	subscriptionPB, err := transformSubscriptionToPB(subscription)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.GetSubscriptionResponse{
		Subscription: subscriptionPB,
	}, nil
}

func (h Handler) CancelSubscription(ctx context.Context, request *frontierv1beta1.CancelSubscriptionRequest) (*frontierv1beta1.CancelSubscriptionResponse, error) {
	logger := grpczap.Extract(ctx)

	_, err := h.subscriptionService.Cancel(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.CancelSubscriptionResponse{}, nil
}

func transformSubscriptionToPB(subs subscription.Subscription) (*frontierv1beta1.Subscription, error) {
	metaData, err := subs.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Subscription{}, err
	}
	subsPb := &frontierv1beta1.Subscription{
		Id:         subs.ID,
		CustomerId: subs.CustomerID,
		PlanId:     subs.PlanID,
		ProviderId: subs.ProviderID,
		State:      subs.State,
		Metadata:   metaData,
		CreatedAt:  timestamppb.New(subs.CreatedAt),
		UpdatedAt:  timestamppb.New(subs.UpdatedAt),
		CanceledAt: timestamppb.New(subs.CanceledAt),
	}
	return subsPb, nil
}
