package entitlement

import (
	"context"
	"errors"

	"golang.org/x/exp/slices"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

type SubscriptionService interface {
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
}

type ProductService interface {
	List(ctx context.Context, f product.Filter) ([]product.Product, error)
	GetFeatureByID(ctx context.Context, id string) (product.Feature, error)
	GetByID(ctx context.Context, id string) (product.Product, error)
}

type Service struct {
	subscriptionService SubscriptionService
	productService      ProductService
}

func NewEntitlementService(subscriptionService SubscriptionService,
	featureService ProductService) *Service {
	return &Service{
		subscriptionService: subscriptionService,
		productService:      featureService,
	}
}

func (s *Service) Check(ctx context.Context, customerID, featureOrProductID string) (bool, error) {
	// get all subscriptions for the customer
	subs, err := s.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return false, err
	}

	// get the feature
	feature, err := s.productService.GetFeatureByID(ctx, featureOrProductID)
	if err != nil {
		return false, err
	}

	// get all the products this feature is in
	products, err := s.productService.List(ctx, product.Filter{
		ProductIDs: feature.ProductIDs,
	})

	// could be product ID as well
	asProduct, err := s.productService.GetByID(ctx, featureOrProductID)
	if err != nil && !errors.Is(err, product.ErrProductNotFound) {
		return false, err
	}
	products = append(products, asProduct)

	// check if the product is in any of the subscriptions
	for _, sub := range subs {
		if sub.State != string(subscription.StateActive) {
			continue
		}
		for _, p := range products {
			if slices.Contains(p.PlanIDs, sub.PlanID) {
				return true, nil
			}
		}
	}
	return false, nil
}
