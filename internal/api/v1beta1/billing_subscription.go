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
	Create(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error)
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
}

func (h Handler) CreateSubscription(ctx context.Context, request *frontierv1beta1.CreateSubscriptionRequest) (*frontierv1beta1.CreateSubscriptionResponse, error) {
	logger := grpczap.Extract(ctx)

	// create subscription
	newSubscription, err := h.subscriptionService.Create(ctx, subscription.Subscription{
		CustomerID: request.GetCustomerId(),
		PlanID:     request.GetBody().GetPlan(),
		SuccessUrl: request.GetBody().GetSuccessUrl(),
		CancelUrl:  request.GetBody().GetCancelUrl(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	subscriptionPB, err := transformSubscriptionToPB(newSubscription)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.CreateSubscriptionResponse{
		Subscription: subscriptionPB,
	}, nil
}

func (h Handler) ListSubscriptions(ctx context.Context, request *frontierv1beta1.ListSubscriptionsRequest) (*frontierv1beta1.ListSubscriptionsResponse, error) {
	logger := grpczap.Extract(ctx)

	var subscriptions []*frontierv1beta1.Subscription
	subscriptionList, err := h.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: request.GetCustomerId(),
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

func transformSubscriptionToPB(subs subscription.Subscription) (*frontierv1beta1.Subscription, error) {
	metaData, err := subs.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Subscription{}, err
	}
	checkoutURL := ""
	if url, ok := subs.Metadata["checkout_url"]; ok {
		checkoutURL = url.(string)
	}
	subsPb := &frontierv1beta1.Subscription{
		Id:          subs.ID,
		CustomerId:  subs.CustomerID,
		PlanId:      subs.PlanID,
		ProviderId:  subs.ProviderID,
		CancelUrl:   subs.CancelUrl,
		SuccessUrl:  subs.SuccessUrl,
		CheckoutUrl: checkoutURL,
		State:       subs.State,
		CreatedAt:   timestamppb.New(subs.CreatedAt),
		UpdatedAt:   timestamppb.New(subs.UpdatedAt),
		Metadata:    metaData,
	}
	if subs.CanceledAt != nil {
		subsPb.CanceledAt = timestamppb.New(*subs.CanceledAt)
	}
	return subsPb, nil
}
