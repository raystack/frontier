package plan

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/raystack/frontier/billing/feature"
	"github.com/raystack/frontier/internal/store/blob"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stripe/stripe-go/v75/client"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Plan, error)
	GetByName(ctx context.Context, name string) (Plan, error)
	Create(ctx context.Context, plan Plan) (Plan, error)
	UpdateByName(ctx context.Context, plan Plan) (Plan, error)
	List(ctx context.Context, filter Filter) ([]Plan, error)
}

type FeatureService interface {
	Create(ctx context.Context, f feature.Feature) (feature.Feature, error)
	GetByID(ctx context.Context, id string) (feature.Feature, error)
	Update(ctx context.Context, f feature.Feature) (feature.Feature, error)
	AddPlan(ctx context.Context, planID string, f feature.Feature) error

	CreatePrice(ctx context.Context, price feature.Price, interval string) (feature.Price, error)
	UpdatePrice(ctx context.Context, price feature.Price) (feature.Price, error)
	GetPriceByID(ctx context.Context, id string) (feature.Price, error)
	GetPriceByFeatureID(ctx context.Context, id string) ([]feature.Price, error)

	List(ctx context.Context, flt feature.Filter) ([]feature.Feature, error)
}

type Service struct {
	planRepository Repository
	stripeClient   *client.API
	featureService FeatureService
}

func NewService(stripeClient *client.API, planRepository Repository, featureService FeatureService) *Service {
	return &Service{
		stripeClient:   stripeClient,
		planRepository: planRepository,
		featureService: featureService,
	}
}

func (s Service) Create(ctx context.Context, p Plan) (Plan, error) {
	return s.planRepository.Create(ctx, p)
}

func (s Service) GetByID(ctx context.Context, id string) (Plan, error) {
	var fetchedPlan Plan
	var err error
	if utils.IsValidUUID(id) {
		fetchedPlan, err = s.planRepository.GetByID(ctx, id)
		if err != nil {
			return Plan{}, err
		}
	}
	fetchedPlan, err = s.planRepository.GetByName(ctx, id)
	if err != nil {
		return Plan{}, err
	}

	// enrich with feature
	features, err := s.featureService.List(ctx, feature.Filter{
		PlanID: fetchedPlan.ID,
	})
	if err != nil {
		return Plan{}, err
	}
	fetchedPlan.Features = features
	return fetchedPlan, nil
}

func (s Service) List(ctx context.Context, filter Filter) ([]Plan, error) {
	listedPlans, err := s.planRepository.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	// enrich with feature
	for i, listedPlan := range listedPlans {
		// TODO(kushsharma): we can do this in one query
		features, err := s.featureService.List(ctx, feature.Filter{
			PlanID: listedPlan.ID,
		})
		if err != nil {
			return nil, err
		}
		listedPlans[i].Features = features
	}
	return listedPlans, nil
}

func (s Service) UpsertLocal(ctx context.Context, blobFile blob.PlanFile) error {
	// create features first
	for _, blobFeature := range blobFile.Features {
		featureOb, err := s.featureService.GetByID(ctx, blobFeature.Name)
		if err != nil && errors.Is(err, feature.ErrFeatureNotFound) {
			// create feature
			if featureOb, err = s.featureService.Create(ctx, feature.Feature{
				Name:        blobFeature.Name,
				Title:       blobFeature.Title,
				Description: blobFeature.Description,
				Interval:    blobFeature.Interval,
				Metadata:    metadata.FromString(blobFeature.Metadata),
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// update feature
			if _, err = s.featureService.Update(ctx, feature.Feature{
				ID:          featureOb.ID,
				ProviderID:  featureOb.ProviderID,
				Name:        blobFeature.Name,
				Title:       blobFeature.Title,
				Description: blobFeature.Description,
			}); err != nil {
				return err
			}
		}

		// ensure price exists
		for blobIdx, blobPrice := range blobFeature.Prices {
			if blobPrice.Name == "" {
				blobPrice.Name = fmt.Sprintf("default_%d", blobIdx)
			}
			priceObs, err := s.featureService.GetPriceByFeatureID(ctx, featureOb.ID)
			if err != nil {
				return fmt.Errorf("failed to get price by feature id: %w", err)
			}
			// find price by name
			var priceOb feature.Price
			for _, p := range priceObs {
				if p.Name == blobPrice.Name {
					priceOb = p
					break
				}
			}
			if priceOb.ID == "" {
				// create price
				if priceOb, err = s.featureService.CreatePrice(ctx, feature.Price{
					Name:             blobPrice.Name,
					Amount:           blobPrice.Amount,
					Currency:         blobPrice.Currency,
					BillingScheme:    feature.BillingSchemeFlat, // TODO(kushsharma): support tiered
					UsageType:        feature.PriceUsageType(blobPrice.UsageType),
					MeteredAggregate: blobPrice.MeteredAggregate,
					FeatureID:        featureOb.ID,
					Metadata:         metadata.FromString(blobPrice.Metadata),
				}, featureOb.Interval); err != nil {
					return err
				}
			} else {
				// update price
				if _, err = s.featureService.UpdatePrice(ctx, feature.Price{
					ID:         priceOb.ID,
					ProviderID: priceOb.ProviderID,
					FeatureID:  priceOb.FeatureID,
					Name:       priceOb.Name,
				}); err != nil {
					return err
				}
			}
		}
	}

	// create plans
	for _, blobPlan := range blobFile.Plans {
		// ensure plan exists
		planOb, err := s.GetByID(ctx, blobPlan.Name)
		if err != nil && errors.Is(err, ErrNotFound) {
			// create plan
			if planOb, err = s.planRepository.Create(ctx, Plan{
				Name:        blobPlan.Name,
				Title:       blobPlan.Title,
				Description: blobPlan.Description,
				Interval:    blobPlan.Interval,
				Metadata:    metadata.FromString(blobPlan.Metadata),
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// update plan
			if _, err = s.planRepository.UpdateByName(ctx, Plan{
				ID:          planOb.ID,
				Name:        blobPlan.Name,
				Title:       blobPlan.Title,
				Description: blobPlan.Description,
			}); err != nil {
				return err
			}
		}

		// ensure feature exists, if not fail
		for _, blobFeature := range blobPlan.Features {
			featureOb, err := s.featureService.GetByID(ctx, blobFeature.Name)
			if err != nil {
				return err
			}

			// ensure plan is added to feature
			if featureOb.Interval != planOb.Interval {
				return fmt.Errorf("feature %s has interval %s, while plan %s has interval %s",
					featureOb.Name, featureOb.Interval, planOb.Name, planOb.Interval)
			}
			if err = s.featureService.AddPlan(ctx, planOb.ID, featureOb); err != nil {
				return err
			}
		}
	}

	return nil
}
