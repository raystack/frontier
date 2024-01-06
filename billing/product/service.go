package product

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/mcuadros/go-defaults"

	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Product, error)
	GetByName(ctx context.Context, name string) (Product, error)
	Create(ctx context.Context, feature Product) (Product, error)
	UpdateByName(ctx context.Context, feature Product) (Product, error)
	List(ctx context.Context, flt Filter) ([]Product, error)
}

type PriceRepository interface {
	GetByID(ctx context.Context, id string) (Price, error)
	GetByName(ctx context.Context, name string) (Price, error)
	Create(ctx context.Context, price Price) (Price, error)
	UpdateByID(ctx context.Context, price Price) (Price, error)
	List(ctx context.Context, flt Filter) ([]Price, error)
}

type Service struct {
	stripeClient    *client.API
	repository      Repository
	priceRepository PriceRepository
}

func NewService(stripeClient *client.API, repository Repository,
	priceRepository PriceRepository) *Service {
	return &Service{
		stripeClient:    stripeClient,
		priceRepository: priceRepository,
		repository:      repository,
	}
}

func (s *Service) Create(ctx context.Context, feature Product) (Product, error) {
	// create a product in stripe for each feature in plan
	if feature.ID == "" {
		feature.ID = uuid.New().String()
		feature.ProviderID = feature.ID
	}
	defaults.SetDefaults(&feature)
	if feature.CreditAmount > 0 {
		feature.Behavior = CreditBehavior
	}

	_, err := s.stripeClient.Products.New(&stripe.ProductParams{
		Params: stripe.Params{
			Context: ctx,
		},
		ID:          &feature.ProviderID,
		Name:        &feature.Title,
		Description: &feature.Description,
		Metadata: map[string]string{
			"name":          feature.Name,
			"credit_amount": fmt.Sprintf("%d", feature.CreditAmount),
			"behavior":      feature.Behavior.String(),
			"managed_by":    "frontier",
		},
	})
	if err != nil {
		return Product{}, err
	}

	featureOb, err := s.repository.Create(ctx, feature)
	if err != nil {
		return Product{}, err
	}

	// create prices if provided
	for _, price := range feature.Prices {
		price.ProductID = featureOb.ID
		priceOb, err := s.CreatePrice(ctx, price)
		if err != nil {
			return Product{}, fmt.Errorf("failed to create price for feature %s: %w", featureOb.ID, err)
		}
		featureOb.Prices = append(featureOb.Prices, priceOb)
	}

	return featureOb, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Product, error) {
	var fetchedProduct Product
	var err error
	if utils.IsValidUUID(id) {
		fetchedProduct, err = s.repository.GetByID(ctx, id)
		if err != nil {
			return Product{}, err
		}
	} else {
		fetchedProduct, err = s.repository.GetByName(ctx, id)
		if err != nil {
			return Product{}, err
		}
	}

	if fetchedProduct.Prices, err = s.GetPriceByProductID(ctx, fetchedProduct.ID); err != nil {
		return Product{}, fmt.Errorf("failed to fetch prices for feature %s: %w", fetchedProduct.ID, err)
	}
	return fetchedProduct, nil
}

// Update updates a feature, but it doesn't update all fields
// ideally we should keep it immutable and create a new feature
func (s *Service) Update(ctx context.Context, feature Product) (Product, error) {
	// update product in stripe
	_, err := s.stripeClient.Products.Update(feature.ProviderID, &stripe.ProductParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Name:        &feature.Title,
		Description: &feature.Description,
		Metadata: map[string]string{
			"name":       feature.Name,
			"plan_ids":   strings.Join(feature.PlanIDs, ","),
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Product{}, err
	}
	return s.repository.UpdateByName(ctx, feature)
}

func (s *Service) AddPlan(ctx context.Context, planID string, featureOb Product) error {
	if !slices.Contains(featureOb.PlanIDs, planID) {
		featureOb.PlanIDs = append(featureOb.PlanIDs, planID)
	}
	featureOb, err := s.Update(ctx, featureOb)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) CreatePrice(ctx context.Context, price Price) (Price, error) {
	// set defaults
	if price.BillingScheme == "" {
		price.BillingScheme = BillingSchemeFlat
	}
	if price.Currency == "" {
		price.Currency = "usd"
	}
	if price.UsageType == "" {
		price.UsageType = PriceUsageTypeLicensed
	}

	providerParams := &stripe.PriceParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Product:       &price.ProductID,
		Nickname:      &price.Name,
		BillingScheme: stripe.String(price.BillingScheme.ToStripe()),
		Currency:      &price.Currency,
		UnitAmount:    &price.Amount,
		Metadata: map[string]string{
			"name":       price.Name,
			"managed_by": "frontier",
		},
	}
	if price.Interval != "" {
		providerParams.Recurring = &stripe.PriceRecurringParams{
			Interval:  stripe.String(price.Interval),
			UsageType: stripe.String(price.UsageType.ToStripe()),
		}
		if price.UsageType == PriceUsageTypeMetered {
			providerParams.Recurring.AggregateUsage = stripe.String(price.MeteredAggregate)
		}
	}
	stripePrice, err := s.stripeClient.Prices.New(providerParams)
	if err != nil {
		return Price{}, err
	}

	price.ProviderID = stripePrice.ID
	return s.priceRepository.Create(ctx, price)
}

func (s *Service) GetPriceByID(ctx context.Context, id string) (Price, error) {
	if utils.IsValidUUID(id) {
		return s.priceRepository.GetByID(ctx, id)
	}
	return s.priceRepository.GetByName(ctx, id)
}

func (s *Service) GetPriceByProductID(ctx context.Context, id string) ([]Price, error) {
	return s.priceRepository.List(ctx, Filter{
		ProductIDs: []string{id},
	})
}

// UpdatePrice updates a price, but it doesn't update all fields
// ideally we should keep it immutable and create a new price
func (s *Service) UpdatePrice(ctx context.Context, price Price) (Price, error) {
	_, err := s.stripeClient.Prices.Update(price.ProviderID, &stripe.PriceParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Nickname: &price.Name,
		Metadata: map[string]string{
			"name":       price.Name,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Price{}, err
	}
	return s.priceRepository.UpdateByID(ctx, price)
}

func (s *Service) List(ctx context.Context, flt Filter) ([]Product, error) {
	listedProducts, err := s.repository.List(ctx, flt)
	if err != nil {
		return nil, err
	}

	// enrich with prices
	for i, listedProduct := range listedProducts {
		// TODO(kushsharma): we can do this in one query
		price, err := s.GetPriceByProductID(ctx, listedProduct.ID)
		if err != nil {
			return nil, err
		}
		listedProducts[i].Prices = price
	}
	return listedProducts, nil
}
