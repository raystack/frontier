package entitlement

import (
	"context"
	"errors"

	"github.com/raystack/frontier/billing/plan"

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

type PlanService interface {
	GetByID(ctx context.Context, id string) (plan.Plan, error)
}

type OrganizationService interface {
	MemberCount(ctx context.Context, orgID string) (int64, error)
}

type Service struct {
	subscriptionService SubscriptionService
	productService      ProductService
	planService         PlanService
	organizationService OrganizationService
}

func NewEntitlementService(subscriptionService SubscriptionService,
	featureService ProductService, planService PlanService,
	organizationService OrganizationService) *Service {
	return &Service{
		subscriptionService: subscriptionService,
		productService:      featureService,
		planService:         planService,
		organizationService: organizationService,
	}
}

// Check checks if the customer has access to the feature or product
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
	if err != nil {
		return false, err
	}

	// could be product ID as well
	asProduct, err := s.productService.GetByID(ctx, featureOrProductID)
	if err != nil && !errors.Is(err, product.ErrProductNotFound) {
		return false, err
	}
	products = append(products, asProduct)

	// check if the product is in any of the subscriptions
	for _, sub := range subs {
		if !sub.IsActive() {
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

// CheckPlanEligibility checks if the customer is eligible for the plan
func (s *Service) CheckPlanEligibility(ctx context.Context, customerID string) error {
	// get all subscriptions for the customer
	subs, err := s.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if !sub.IsActive() {
			continue
		}

		planOb, err := s.planService.GetByID(ctx, sub.PlanID)
		if err != nil {
			return err
		}

		// check if the product has seat based limits
		for _, prod := range planOb.Products {
			if prod.Behavior == product.PerSeatBehavior {
				count, err := s.organizationService.MemberCount(ctx, customerID)
				if err != nil {
					return err
				}
				if prod.Config.SeatLimit > 0 && count > prod.Config.SeatLimit {
					return product.ErrPerSeatLimitReached
				}
			}
		}
	}

	// default to true
	return nil
}
