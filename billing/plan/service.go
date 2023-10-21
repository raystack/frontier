package plan

import (
	"context"
	"errors"

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
	Create(ctx context.Context, feature feature.Feature) (feature.Feature, error)
	GetByID(ctx context.Context, id string) (feature.Feature, error)
	Update(ctx context.Context, feature feature.Feature) (feature.Feature, error)

	CreatePrice(ctx context.Context, price feature.Price, interval string) (feature.Price, error)
	UpdatePrice(ctx context.Context, price feature.Price) (feature.Price, error)
	GetPriceByID(ctx context.Context, id string) (feature.Price, error)
	GetPriceByFeatureID(ctx context.Context, id string) (feature.Price, error)

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
				Metadata:    metadata.BuildString(blobPlan.Metadata),
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

		// ensure feature exists
		for _, blobFeature := range blobPlan.Features {
			featureOb, err := s.featureService.GetByID(ctx, blobFeature.Name)
			if err != nil && errors.Is(err, feature.ErrFeatureNotFound) {
				// create feature
				if featureOb, err = s.featureService.Create(ctx, feature.Feature{
					Name:        blobFeature.Name,
					Title:       blobFeature.Title,
					Description: blobFeature.Description,
					PlanID:      planOb.ID,
					Metadata:    metadata.BuildString(blobFeature.Metadata),
				}); err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// update feature
				if _, err = s.featureService.Update(ctx, feature.Feature{
					ID:          featureOb.ID,
					Name:        blobFeature.Name,
					Title:       blobFeature.Title,
					Description: blobFeature.Description,
					PlanID:      planOb.ID,
				}); err != nil {
					return err
				}
			}

			// ensure price exists
			priceOb, err := s.featureService.GetPriceByFeatureID(ctx, featureOb.ID)
			if err != nil && errors.Is(err, feature.ErrPriceNotFound) {
				// create price
				if priceOb, err = s.featureService.CreatePrice(ctx, feature.Price{
					Name:             blobFeature.Name,
					Amount:           blobFeature.Price.Amount,
					Currency:         blobFeature.Price.Currency,
					BillingScheme:    feature.BillingSchemeFlat, // TODO(kushsharma): support tiered
					UsageType:        feature.PriceUsageType(blobFeature.Price.UsageType),
					MeteredAggregate: blobFeature.Price.MeteredAggregate,
					FeatureID:        featureOb.ID,
					Metadata:         metadata.BuildString(blobFeature.Price.Metadata),
				}, planOb.Interval); err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// update price
				if _, err = s.featureService.UpdatePrice(ctx, feature.Price{
					ID:         priceOb.ID,
					ProviderID: priceOb.ProviderID,
					FeatureID:  priceOb.FeatureID,
					Name:       featureOb.Name,
				}); err != nil {
					return err
				}
			}
		}
	}

	// separate on demand features
	for _, blobFeature := range blobFile.Features {
		featureOb, err := s.featureService.GetByID(ctx, blobFeature.Name)
		if err != nil && errors.Is(err, feature.ErrFeatureNotFound) {
			// create feature
			if featureOb, err = s.featureService.Create(ctx, feature.Feature{
				Name:        blobFeature.Name,
				Title:       blobFeature.Title,
				Description: blobFeature.Description,
				Metadata:    metadata.BuildString(blobFeature.Metadata),
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// update feature
			if _, err = s.featureService.Update(ctx, feature.Feature{
				ID:          featureOb.ID,
				Name:        blobFeature.Name,
				Title:       blobFeature.Title,
				Description: blobFeature.Description,
			}); err != nil {
				return err
			}
		}

		// ensure price exists
		priceOb, err := s.featureService.GetPriceByFeatureID(ctx, featureOb.ID)
		if err != nil && errors.Is(err, feature.ErrPriceNotFound) {
			// create price
			if priceOb, err = s.featureService.CreatePrice(ctx, feature.Price{
				Name:             blobFeature.Name,
				Amount:           blobFeature.Price.Amount,
				Currency:         blobFeature.Price.Currency,
				BillingScheme:    feature.BillingSchemeFlat, // TODO(kushsharma): support tiered
				UsageType:        feature.PriceUsageType(blobFeature.Price.UsageType),
				MeteredAggregate: blobFeature.Price.MeteredAggregate,
				FeatureID:        featureOb.ID,
				Metadata:         metadata.BuildString(blobFeature.Price.Metadata),
			}, ""); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// update price
			if _, err = s.featureService.UpdatePrice(ctx, feature.Price{
				ID:         priceOb.ID,
				ProviderID: priceOb.ProviderID,
				FeatureID:  priceOb.FeatureID,
				Name:       featureOb.Name,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
