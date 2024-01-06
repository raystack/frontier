package entitlement

import (
	"context"

	"golang.org/x/exp/slices"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

type SubscriptionService interface {
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
}

type FeatureService interface {
	GetByID(ctx context.Context, id string) (product.Product, error)
}

type Service struct {
	subscriptionService SubscriptionService
	featureService      FeatureService
}

func NewEntitlementService(subscriptionService SubscriptionService,
	featureService FeatureService) *Service {
	return &Service{
		subscriptionService: subscriptionService,
		featureService:      featureService,
	}
}

func (s *Service) Check(ctx context.Context, customerID, featureID string) (bool, error) {
	// get all subscriptions for the customer
	subs, err := s.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return false, err
	}

	// get the feature
	feature, err := s.featureService.GetByID(ctx, featureID)
	if err != nil {
		return false, err
	}

	// check if the feature is in any of the subscriptions
	for _, sub := range subs {
		if sub.State != string(subscription.StateActive) {
			continue
		}
		if slices.Contains(feature.PlanIDs, sub.PlanID) {
			return true, nil
		}
	}
	return false, nil
}
