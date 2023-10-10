package plan

import (
	"context"
	"errors"

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
	Create(ctx context.Context, feature feature.Feature) (feature.Feature, error)
	GetByID(ctx context.Context, id string) (feature.Feature, error)
	Update(ctx context.Context, feature feature.Feature) (feature.Feature, error)

	CreatePrice(ctx context.Context, price feature.Price) (feature.Price, error)
	UpdatePrice(ctx context.Context, price feature.Price) (feature.Price, error)
	GetPriceByID(ctx context.Context, id string) (feature.Price, error)

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
	if utils.IsValidUUID(id) {
		return s.planRepository.GetByID(ctx, id)
	}
	return s.planRepository.GetByName(ctx, id)
}

func (s Service) List(ctx context.Context, filter Filter) ([]Plan, error) {
	listedPlans, err := s.planRepository.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	// enrich with feature
	for i, listedPlan := range listedPlans {
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

func (s Service) UpsertLocal(ctx context.Context, blobPlans []blob.Plan) error {
	for _, blobPlan := range blobPlans {
		// ensure plan exists
		plan, err := s.GetByID(ctx, blobPlan.Name)
		if err != nil && errors.Is(err, ErrNotFound) {
			// create plan
			if plan, err = s.planRepository.Create(ctx, Plan{
				Name:        blobPlan.Name,
				Title:       blobPlan.Title,
				Description: blobPlan.Description,
			}); err != nil {
				return err
			}
		} else if err == nil {
			// update plan
			if _, err = s.planRepository.UpdateByName(ctx, Plan{
				ID:          plan.ID,
				Name:        blobPlan.Name,
				Title:       blobPlan.Title,
				Description: blobPlan.Description,
				// TODO: update metadata
			}); err != nil {
				return err
			}
		}

		// ensure feature exists
		for _, blobFeature := range blobPlan.Features {
			featureOb, err := s.featureService.GetByID(ctx, blobFeature.Name)
			if err != nil && errors.Is(err, feature.ErrFeatureNotFound) {
				// create feature
				if featureOb, err = s.featureService.Create(ctx, feature.Feature{
					Name:        blobFeature.Name,
					Title:       blobFeature.Title,
					Description: blobFeature.Description,
					PlanID:      plan.ID,
				}); err != nil {
					return err
				}
			} else if err == nil {
				// update feature
				if _, err = s.featureService.Update(ctx, feature.Feature{
					ID:          featureOb.ID,
					Name:        blobFeature.Name,
					Title:       blobFeature.Title,
					Description: blobFeature.Description,
					PlanID:      plan.ID,
				}); err != nil {
					return err
				}
			}

			// ensure price exists
			for _, blobPrice := range blobFeature.Prices {
				priceOb, err := s.featureService.GetPriceByID(ctx, blobPrice.Name)
				if err != nil && errors.Is(err, feature.ErrPriceNotFound) {
					// create price
					if priceOb, err = s.featureService.CreatePrice(ctx, feature.Price{
						Name:              blobPrice.Name,
						Title:             blobFeature.Title,
						Amount:            blobPrice.Amount,
						Currency:          blobPrice.Currency,
						BillingScheme:     feature.BillingSchemeFlat, // TODO(kushsharma): support tiered
						RecurringInterval: blobPrice.RecurringInterval,
						UsageType:         feature.PriceUsageType(blobPrice.UsageType),
						MeteredAggregate:  blobPrice.MeteredAggregate,
						FeatureID:         featureOb.ID,
					}); err != nil {
						return err
					}
				} else if err == nil {
					// update price
					if _, err = s.featureService.UpdatePrice(ctx, feature.Price{
						ID:         priceOb.ID,
						ProviderID: priceOb.ProviderID,
						FeatureID:  priceOb.FeatureID,
						Name:       blobPrice.Name,
						Title:      blobFeature.Title,
					}); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
