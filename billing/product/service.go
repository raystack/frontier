package product

import (
	"context"
	"errors"
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
	Create(ctx context.Context, product Product) (Product, error)
	UpdateByName(ctx context.Context, product Product) (Product, error)
	List(ctx context.Context, flt Filter) ([]Product, error)
}

type PriceRepository interface {
	GetByID(ctx context.Context, id string) (Price, error)
	GetByName(ctx context.Context, name string) (Price, error)
	Create(ctx context.Context, price Price) (Price, error)
	UpdateByID(ctx context.Context, price Price) (Price, error)
	List(ctx context.Context, flt Filter) ([]Price, error)
}

type FeatureRepository interface {
	GetByID(ctx context.Context, id string) (Feature, error)
	GetByName(ctx context.Context, name string) (Feature, error)
	Create(ctx context.Context, feature Feature) (Feature, error)
	UpdateByName(ctx context.Context, feature Feature) (Feature, error)
	List(ctx context.Context, flt Filter) ([]Feature, error)
}

type Service struct {
	stripeClient      *client.API
	productRepository Repository
	priceRepository   PriceRepository
	featureRepository FeatureRepository
}

func NewService(stripeClient *client.API, productRepository Repository,
	priceRepository PriceRepository, featureRepository FeatureRepository) *Service {
	return &Service{
		stripeClient:      stripeClient,
		priceRepository:   priceRepository,
		productRepository: productRepository,
		featureRepository: featureRepository,
	}
}

func (s *Service) Create(ctx context.Context, product Product) (Product, error) {
	// create a product in stripe for each product in plan
	if product.ID == "" {
		product.ID = uuid.New().String()
		product.ProviderID = product.ID
	}
	defaults.SetDefaults(&product)
	if product.CreditAmount > 0 {
		product.Behavior = CreditBehavior
	}

	_, err := s.stripeClient.Products.New(&stripe.ProductParams{
		Params: stripe.Params{
			Context: ctx,
		},
		ID:          &product.ProviderID,
		Name:        &product.Title,
		Description: &product.Description,
		Metadata: map[string]string{
			"name":          product.Name,
			"credit_amount": fmt.Sprintf("%d", product.CreditAmount),
			"behavior":      product.Behavior.String(),
			"managed_by":    "frontier",
		},
	})
	if err != nil {
		return Product{}, err
	}

	productOb, err := s.productRepository.Create(ctx, product)
	if err != nil {
		return Product{}, err
	}

	// create prices if provided
	for _, price := range product.Prices {
		price.ProductID = productOb.ID
		priceOb, err := s.CreatePrice(ctx, price)
		if err != nil {
			return Product{}, fmt.Errorf("failed to create price for product %s: %w", productOb.ID, err)
		}
		productOb.Prices = append(productOb.Prices, priceOb)
	}

	// create features if provided
	for _, feature := range product.Features {
		feature.ProductIDs = append(feature.ProductIDs, productOb.ID)
		featureOb, err := s.UpsertFeature(ctx, feature)
		if err != nil {
			return Product{}, fmt.Errorf("failed to create feature for product %s: %w", productOb.ID, err)
		}
		productOb.Features = append(productOb.Features, featureOb)
	}

	return productOb, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Product, error) {
	var fetchedProduct Product
	var err error
	if utils.IsValidUUID(id) {
		fetchedProduct, err = s.productRepository.GetByID(ctx, id)
		if err != nil {
			return Product{}, err
		}
	} else {
		fetchedProduct, err = s.productRepository.GetByName(ctx, id)
		if err != nil {
			return Product{}, err
		}
	}

	if fetchedProduct.Prices, err = s.GetPriceByProductID(ctx, fetchedProduct.ID); err != nil {
		return Product{}, fmt.Errorf("failed to fetch prices for product %s: %w", fetchedProduct.ID, err)
	}
	if fetchedProduct.Features, err = s.GetFeatureByProductID(ctx, fetchedProduct.ID); err != nil {
		return Product{}, fmt.Errorf("failed to fetch features for product %s: %w", fetchedProduct.ID, err)
	}
	return fetchedProduct, nil
}

// Update updates a product, but it doesn't update all fields
// ideally we should keep it immutable and create a new product
func (s *Service) Update(ctx context.Context, product Product) (Product, error) {
	// update product in stripe
	_, err := s.stripeClient.Products.Update(product.ProviderID, &stripe.ProductParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Name:        &product.Title,
		Description: &product.Description,
		Metadata: map[string]string{
			"name":       product.Name,
			"plan_ids":   strings.Join(product.PlanIDs, ","),
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Product{}, err
	}
	return s.productRepository.UpdateByName(ctx, product)
}

func (s *Service) AddPlan(ctx context.Context, productOb Product, planID string) error {
	var err error
	if !slices.Contains(productOb.PlanIDs, planID) {
		productOb.PlanIDs = append(productOb.PlanIDs, planID)
		_, err = s.Update(ctx, productOb)
		if err != nil {
			return err
		}
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
	if len(id) == 0 {
		return []Price{}, nil
	}
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
	listedProducts, err := s.productRepository.List(ctx, flt)
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

func (s *Service) UpsertFeature(ctx context.Context, feature Feature) (Feature, error) {
	if len(feature.ID) == 0 {
		feature.ID = uuid.New().String()
	}
	if len(feature.Name) == 0 {
		return Feature{}, fmt.Errorf("feature name is required: %w", ErrInvalidFeatureDetail)
	}
	feature.ProductIDs = utils.Deduplicate(feature.ProductIDs)
	existingFeature, err := s.GetFeatureByID(ctx, feature.Name)
	if err != nil && errors.Is(err, ErrFeatureNotFound) {
		return s.featureRepository.Create(ctx, feature)
	}

	existingFeature.ProductIDs = feature.ProductIDs
	existingFeature.Metadata = feature.Metadata
	return s.featureRepository.UpdateByName(ctx, existingFeature)
}

func (s *Service) GetFeatureByID(ctx context.Context, id string) (Feature, error) {
	if utils.IsValidUUID(id) {
		return s.featureRepository.GetByID(ctx, id)
	}
	return s.featureRepository.GetByName(ctx, id)
}

func (s *Service) GetFeatureByProductID(ctx context.Context, id string) ([]Feature, error) {
	if len(id) == 0 {
		return []Feature{}, nil
	}
	return s.featureRepository.List(ctx, Filter{
		ProductID: id,
	})
}

func (s *Service) ListFeatures(ctx context.Context, flt Filter) ([]Feature, error) {
	return s.featureRepository.List(ctx, flt)
}
