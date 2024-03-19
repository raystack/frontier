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
	if product.Config.CreditAmount > 0 {
		product.Behavior = CreditBehavior
	}
	product.Name = strings.ToLower(product.Name)

	_, err := s.stripeClient.Products.New(&stripe.ProductParams{
		Params: stripe.Params{
			Context: ctx,
		},
		ID:          &product.ProviderID,
		Name:        &product.Title,
		Description: &product.Description,
		Metadata: map[string]string{
			"name":          product.Name,
			"credit_amount": fmt.Sprintf("%d", product.Config.CreditAmount),
			"behavior":      product.Behavior.String(),
			"product_id":    product.ID,
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

	fetchedProduct, err = s.populateProduct(ctx, fetchedProduct)
	if err != nil {
		return Product{}, err
	}
	return fetchedProduct, nil
}

func (s *Service) GetByProviderID(ctx context.Context, id string) (Product, error) {
	return s.GetByID(ctx, id)
}

// populate product with price and features
func (s *Service) populateProduct(ctx context.Context, product Product) (Product, error) {
	var err error
	product.Prices, err = s.GetPriceByProductID(ctx, product.ID)
	if err != nil {
		return Product{}, fmt.Errorf("failed to fetch prices for product %s: %w", product.ID, err)
	}
	product.Features, err = s.GetFeatureByProductID(ctx, product.ID)
	if err != nil {
		return Product{}, fmt.Errorf("failed to fetch features for product %s: %w", product.ID, err)
	}
	return product, nil
}

// Update updates a product, but it doesn't update all fields
// ideally we should keep it immutable and create a new product
func (s *Service) Update(ctx context.Context, product Product) (Product, error) {
	existingProduct, err := s.productRepository.GetByID(ctx, product.ID)
	if err != nil {
		return Product{}, err
	}

	// only following fields will be updated
	if len(product.Title) > 0 {
		existingProduct.Title = product.Title
	}
	if len(product.Description) > 0 {
		existingProduct.Description = product.Description
	}
	if len(product.PlanIDs) > 0 {
		existingProduct.PlanIDs = product.PlanIDs
	}
	if product.Config.CreditAmount > 0 {
		existingProduct.Config.CreditAmount = product.Config.CreditAmount
	}
	if product.Config.SeatLimit > 0 {
		existingProduct.Config.SeatLimit = product.Config.SeatLimit
	}
	if len(product.Metadata) > 0 {
		existingProduct.Metadata = product.Metadata
	}

	// update product in stripe
	_, err = s.stripeClient.Products.Update(existingProduct.ProviderID, &stripe.ProductParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Name:        &existingProduct.Title,
		Description: &existingProduct.Description,
		Metadata: map[string]string{
			"name":       existingProduct.Name,
			"plan_ids":   strings.Join(existingProduct.PlanIDs, ","),
			"behavior":   existingProduct.Behavior.String(),
			"product_id": existingProduct.ID,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Product{}, err
	}

	// check feature updates in product
	var featureErr error
	existingFeatures, err := s.ListFeatures(ctx, Filter{
		ProductID: existingProduct.ID,
	})
	if err != nil {
		return Product{}, err
	}
	for _, existingFeature := range existingFeatures {
		_, found := utils.FindFirst(product.Features, func(f Feature) bool {
			return f.ID == existingFeature.ID
		})
		if !found {
			if err := s.RemoveFeatureFromProduct(ctx, existingFeature.ID, existingProduct.ID); err != nil {
				featureErr = errors.Join(featureErr, err)
			}
		}
	}
	for _, feature := range product.Features {
		if err := s.AddFeatureToProduct(ctx, feature, existingProduct.ID); err != nil {
			featureErr = errors.Join(featureErr, err)
		}
	}
	if featureErr != nil {
		return Product{}, fmt.Errorf("failed to update features for product %s: %w", existingProduct.ID, featureErr)
	}

	// update in db
	updatedProduct, err := s.productRepository.UpdateByName(ctx, existingProduct)
	if err != nil {
		return Product{}, err
	}

	// populate product with price and features
	updatedProduct, err = s.populateProduct(ctx, updatedProduct)
	if err != nil {
		return Product{}, err
	}

	return updatedProduct, nil
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
	price.Interval = strings.ToLower(price.Interval)
	price.Name = strings.ToLower(price.Name)

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
			"product_id": price.ProductID,
			"price_id":   price.ID,
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
	existingPrice, err := s.priceRepository.GetByID(ctx, price.ID)
	if err != nil {
		return Price{}, err
	}

	// only following fields will be updated
	if len(price.Name) > 0 {
		existingPrice.Name = strings.ToLower(price.Name)
	}
	if len(price.Metadata) > 0 {
		existingPrice.Metadata = price.Metadata
	}

	_, err = s.stripeClient.Prices.Update(existingPrice.ProviderID, &stripe.PriceParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Nickname: &existingPrice.Name,
		Metadata: map[string]string{
			"name":       existingPrice.Name,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Price{}, err
	}

	return s.priceRepository.UpdateByID(ctx, existingPrice)
}

func (s *Service) List(ctx context.Context, flt Filter) ([]Product, error) {
	listedProducts, err := s.productRepository.List(ctx, flt)
	if err != nil {
		return nil, err
	}

	// enrich with prices
	for i, listedProduct := range listedProducts {
		// TODO(kushsharma): we can do this in one query
		listedProducts[i], err = s.populateProduct(ctx, listedProduct)
		if err != nil {
			return nil, err
		}
	}
	return listedProducts, nil
}

func (s *Service) UpsertFeature(ctx context.Context, feature Feature) (Feature, error) {
	if len(feature.Name) == 0 {
		return Feature{}, fmt.Errorf("feature name is required: %w", ErrInvalidFeatureDetail)
	}
	feature.ProductIDs = utils.Deduplicate(feature.ProductIDs)
	existingFeature, err := s.GetFeatureByID(ctx, feature.Name)
	if err != nil && errors.Is(err, ErrFeatureNotFound) {
		if len(feature.ID) == 0 {
			feature.ID = uuid.New().String()
		}
		return s.featureRepository.Create(ctx, feature)
	}

	existingFeature.ProductIDs = feature.ProductIDs
	if len(feature.Title) > 0 {
		existingFeature.Title = feature.Title
	}
	if len(feature.Metadata) > 0 {
		existingFeature.Metadata = feature.Metadata
	}
	return s.featureRepository.UpdateByName(ctx, existingFeature)
}

func (s *Service) AddFeatureToProduct(ctx context.Context, feature Feature, productID string) error {
	existingFeature, err := s.GetFeatureByID(ctx, feature.Name)
	if err != nil {
		if !errors.Is(err, ErrFeatureNotFound) {
			return err
		}
		// create a new feature if not found
		feature.ProductIDs = append(feature.ProductIDs, productID)
		existingFeature, err = s.UpsertFeature(ctx, feature)
		if err != nil {
			return err
		}
	}

	if !slices.Contains(existingFeature.ProductIDs, productID) {
		existingFeature.ProductIDs = append(existingFeature.ProductIDs, productID)
		_, err = s.featureRepository.UpdateByName(ctx, existingFeature)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RemoveFeatureFromProduct(ctx context.Context, featureID, productID string) error {
	feature, err := s.GetFeatureByID(ctx, featureID)
	if err != nil {
		return err
	}
	if slices.Contains(feature.ProductIDs, productID) {
		feature.ProductIDs = slices.DeleteFunc(feature.ProductIDs, func(id string) bool {
			return id == productID
		})
		_, err = s.featureRepository.UpdateByName(ctx, feature)
		if err != nil {
			return err
		}
	}
	return nil
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
