package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mcuadros/go-defaults"
	"github.com/stripe/stripe-go/v79"

	"slices"

	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stripe/stripe-go/v79/client"
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

	// read and validate the desired prices before mutating anything, so an
	// invalid price fails the whole update rather than leaving the product
	// half-changed. An empty list leaves the prices untouched.
	var currentPrices []Price
	if len(product.Prices) > 0 {
		currentPrices, err = s.GetPriceByProductID(ctx, existingProduct.ID)
		if err != nil {
			return Product{}, err
		}
		if err := validateDesiredPrices(currentPrices, product.Prices); err != nil {
			return Product{}, err
		}
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
	existingProduct.Config.MinQuantity = product.Config.MinQuantity
	existingProduct.Config.MaxQuantity = product.Config.MaxQuantity
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
	featureErr := s.updateProductFeatures(ctx, existingProduct, product)
	if featureErr != nil {
		return Product{}, featureErr
	}

	// update in db
	updatedProduct, err := s.productRepository.UpdateByName(ctx, existingProduct)
	if err != nil {
		return Product{}, err
	}

	// apply the validated price convergence. The desired list is authoritative:
	// a new name is created, an inactive name listed again is activated, and an
	// active price the list no longer names is deactivated.
	if len(product.Prices) > 0 {
		if err := s.applyPriceConvergence(ctx, updatedProduct.ID, currentPrices, product.Prices); err != nil {
			return Product{}, err
		}
	}

	// populate product with price and features
	updatedProduct, err = s.populateProduct(ctx, updatedProduct)
	if err != nil {
		return Product{}, err
	}

	return updatedProduct, nil
}

// priceKey is the canonical lookup name for a price. Names are matched
// case-insensitively and ignoring surrounding whitespace. Validation,
// convergence, and the stored name all use this key, so the three never
// disagree on what counts as the same price.
func priceKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// validateDesiredPrices checks the desired price list against the product's
// current prices without touching anything, so an invalid list fails the update
// before it mutates the product. Names must be present and unique, and a name
// that already exists must keep its immutable fields, since provider prices
// cannot be changed in place.
func validateDesiredPrices(current, desired []Price) error {
	currentByName := make(map[string]Price, len(current))
	for _, p := range current {
		currentByName[priceKey(p.Name)] = p
	}
	seen := make(map[string]struct{}, len(desired))
	for _, want := range desired {
		name := priceKey(want.Name)
		if name == "" {
			return fmt.Errorf("%w: a price must have a name", ErrInvalidDetail)
		}
		if _, dup := seen[name]; dup {
			return fmt.Errorf("%w: price %q is listed more than once", ErrInvalidDetail, name)
		}
		seen[name] = struct{}{}
		if existing, ok := currentByName[name]; ok {
			if err := checkImmutablePriceFields(existing, want); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkImmutablePriceFields rejects a change to a field a provider price cannot
// change in place. A new amount, currency, interval, billing scheme, or usage
// type has to be a new price under a new name. Both sides are normalized first,
// so a field that only differs because of a default does not read as a change.
func checkImmutablePriceFields(existing, want Price) error {
	e := normalizePrice(existing)
	w := normalizePrice(want)
	switch {
	case e.Amount != w.Amount:
		return fmt.Errorf("%w: price %q amount cannot change from %d to %d; provider prices are immutable, add a new price with a different name",
			ErrInvalidDetail, w.Name, e.Amount, w.Amount)
	case e.Currency != w.Currency:
		return fmt.Errorf("%w: price %q currency cannot change from %q to %q; add a new price with a different name",
			ErrInvalidDetail, w.Name, e.Currency, w.Currency)
	case e.Interval != w.Interval:
		return fmt.Errorf("%w: price %q interval cannot change from %q to %q; add a new price with a different name",
			ErrInvalidDetail, w.Name, e.Interval, w.Interval)
	case e.BillingScheme != w.BillingScheme:
		return fmt.Errorf("%w: price %q billing scheme cannot change; add a new price with a different name", ErrInvalidDetail, w.Name)
	case e.UsageType != w.UsageType:
		return fmt.Errorf("%w: price %q usage type cannot change; add a new price with a different name", ErrInvalidDetail, w.Name)
	case e.UsageType == PriceUsageTypeMetered && e.MeteredAggregate != w.MeteredAggregate:
		return fmt.Errorf("%w: price %q metered aggregate cannot change; add a new price with a different name", ErrInvalidDetail, w.Name)
	}
	return nil
}

// normalizePrice fills the defaults CreatePrice would apply, trims and
// lowercases the name, and lowercases the interval, so two prices compare the
// way they are stored.
func normalizePrice(p Price) Price {
	if p.BillingScheme == "" {
		p.BillingScheme = BillingSchemeFlat
	}
	if p.Currency == "" {
		p.Currency = "usd"
	}
	p.Currency = strings.ToLower(p.Currency)
	if p.UsageType == "" {
		p.UsageType = PriceUsageTypeLicensed
	}
	// metered_aggregate defaults to "sum": the create path fills it via
	// SetDefaults while the add-a-price path leaves it empty, so both sides must
	// normalize to the same value or the immutability check would falsely reject
	// one against the other.
	if p.MeteredAggregate == "" {
		p.MeteredAggregate = "sum"
	}
	p.Interval = strings.ToLower(p.Interval)
	p.Name = priceKey(p.Name)
	return p
}

// applyPriceConvergence makes the product's prices match the desired list. The
// list must already have passed validateDesiredPrices. A name the product does
// not have is created, an inactive name listed again is activated, and an active
// price the list no longer names is deactivated. Adds and activations run before
// deactivations, so the product always has the new price before an old one goes
// inactive.
func (s *Service) applyPriceConvergence(ctx context.Context, productID string, current, desired []Price) error {
	currentByName := make(map[string]Price, len(current))
	for _, p := range current {
		currentByName[priceKey(p.Name)] = p
	}
	desiredNames := make(map[string]struct{}, len(desired))
	for _, want := range desired {
		desiredNames[priceKey(want.Name)] = struct{}{}
	}

	for _, want := range desired {
		existing, ok := currentByName[priceKey(want.Name)]
		if !ok {
			want.ProductID = productID
			want.Name = priceKey(want.Name)
			if _, err := s.CreatePrice(ctx, want); err != nil {
				return err
			}
			continue
		}
		if !existing.IsActive() {
			if err := s.activatePrice(ctx, existing); err != nil {
				return err
			}
		}
	}

	for _, p := range current {
		if _, wanted := desiredNames[priceKey(p.Name)]; wanted {
			continue
		}
		if !p.IsActive() {
			continue
		}
		if err := s.deactivatePrice(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// deactivatePrice marks a price inactive. Provider prices are immutable and
// cannot be deleted, so a price is taken out of use by setting it inactive in
// the provider and the repo rather than removed.
func (s *Service) deactivatePrice(ctx context.Context, price Price) error {
	return s.setPriceActive(ctx, price, false)
}

// activatePrice brings an inactive price back into use when the desired list
// names it again.
func (s *Service) activatePrice(ctx context.Context, price Price) error {
	return s.setPriceActive(ctx, price, true)
}

// setPriceActive flips a price's active flag in the provider and its state in
// the repo. A price with no provider id (nothing was created upstream) skips the
// provider call.
func (s *Service) setPriceActive(ctx context.Context, price Price, active bool) error {
	if price.ProviderID != "" {
		if _, err := s.stripeClient.Prices.Update(price.ProviderID, &stripe.PriceParams{
			Params: stripe.Params{Context: ctx},
			Active: stripe.Bool(active),
		}); err != nil {
			return err
		}
	}
	if active {
		price.State = PriceStateActive
	} else {
		price.State = PriceStateInactive
	}
	_, err := s.priceRepository.UpdateByID(ctx, price)
	return err
}

func (s *Service) AddPlan(ctx context.Context, productOb Product, planID string) error {
	var err error
	if !slices.Contains(productOb.PlanIDs, planID) {
		productOb.PlanIDs = append(productOb.PlanIDs, planID)
		// AddPlan only links a plan to the product. Clear the populated price
		// list so Update leaves the product's prices untouched instead of
		// re-converging them.
		productOb.Prices = nil
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
	price.Currency = strings.ToLower(price.Currency)
	if price.UsageType == "" {
		price.UsageType = PriceUsageTypeLicensed
	}
	// store a consistent metered_aggregate so a price added through this path
	// matches one the create path stored via SetDefaults, and a metered price
	// sends a valid aggregate to the provider.
	if price.MeteredAggregate == "" {
		price.MeteredAggregate = "sum"
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
		ProductID: id,
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
			"product_id": price.ProductID,
			"price_id":   price.ID,
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

func (s *Service) updateProductFeatures(ctx context.Context, existingProduct Product, product Product) error {
	var featureErr error
	existingFeatures, err := s.ListFeatures(ctx, Filter{
		ProductID: existingProduct.ID,
	})
	if err != nil {
		return err
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
		return fmt.Errorf("failed to update features for product %s: %w", existingProduct.ID, featureErr)
	}

	return nil
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
