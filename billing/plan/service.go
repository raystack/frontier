package plan

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/raystack/frontier/billing/feature"
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

	CreatePrice(ctx context.Context, price feature.Price) (feature.Price, error)
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
	} else {
		fetchedPlan, err = s.planRepository.GetByName(ctx, id)
	}
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

func (s Service) UpsertPlans(ctx context.Context, planFile File) error {
	// create features first
	for _, featureToCreate := range planFile.Features {
		featureOb, err := s.featureService.GetByID(ctx, featureToCreate.Name)
		if err != nil && errors.Is(err, feature.ErrFeatureNotFound) {
			// create feature
			if featureOb, err = s.featureService.Create(ctx, feature.Feature{
				Name:         featureToCreate.Name,
				Title:        featureToCreate.Title,
				Description:  featureToCreate.Description,
				CreditAmount: featureToCreate.CreditAmount,
				Metadata:     metadata.Build(featureToCreate.Metadata),
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
				Name:        featureToCreate.Name,
				Title:       featureToCreate.Title,
				Description: featureToCreate.Description,
			}); err != nil {
				return err
			}
		}

		// ensure price exists
		for blobIdx, priceToCreate := range featureToCreate.Prices {
			if priceToCreate.Name == "" {
				priceToCreate.Name = fmt.Sprintf("default_%d", blobIdx)
			}
			priceObs, err := s.featureService.GetPriceByFeatureID(ctx, featureOb.ID)
			if err != nil {
				return fmt.Errorf("failed to get price by feature id: %w", err)
			}
			// find price by name
			var priceOb feature.Price
			for _, p := range priceObs {
				if p.Name == priceToCreate.Name {
					priceOb = p
					break
				}
			}
			if priceOb.ID == "" {
				// create price
				if priceOb, err = s.featureService.CreatePrice(ctx, feature.Price{
					Name:             priceToCreate.Name,
					Amount:           priceToCreate.Amount,
					Currency:         priceToCreate.Currency,
					BillingScheme:    priceToCreate.BillingScheme,
					UsageType:        priceToCreate.UsageType,
					MeteredAggregate: priceToCreate.MeteredAggregate,
					Interval:         priceToCreate.Interval,
					FeatureID:        featureOb.ID,
					Metadata:         metadata.Build(priceToCreate.Metadata),
				}); err != nil {
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
	for _, planToCreate := range planFile.Plans {
		// ensure plan exists
		planOb, err := s.GetByID(ctx, planToCreate.Name)
		if err != nil && errors.Is(err, ErrNotFound) {
			// create plan
			if planOb, err = s.planRepository.Create(ctx, Plan{
				Name:        planToCreate.Name,
				Title:       planToCreate.Title,
				Description: planToCreate.Description,
				Interval:    planToCreate.Interval,
				Metadata:    metadata.Build(planToCreate.Metadata),
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// update plan
			if _, err = s.planRepository.UpdateByName(ctx, Plan{
				ID:          planOb.ID,
				Name:        planToCreate.Name,
				Title:       planToCreate.Title,
				Description: planToCreate.Description,
			}); err != nil {
				return err
			}
		}

		// ensures only one feature has free credits
		if len(utils.Filter(planToCreate.Features, func(f feature.Feature) bool {
			return f.CreditAmount > 0
		})) > 1 {
			return fmt.Errorf("plan %s has more than one feature with free credits", planOb.Name)
		}

		// ensure feature exists, if not fail
		for _, featureToCreate := range planToCreate.Features {
			featureOb, err := s.featureService.GetByID(ctx, featureToCreate.Name)
			if err != nil {
				return err
			}

			// ensure plan can be added to feature
			hasMatchingPrice := utils.ContainsFunc(featureOb.Prices, func(p feature.Price) bool {
				return p.Interval == planOb.Interval
			})
			hasFreeCredits := featureOb.CreditAmount > 0
			if !hasMatchingPrice && !hasFreeCredits {
				return fmt.Errorf("feature %s has no prices registered with this interval, plan %s has interval %s",
					featureOb.Name, planOb.Name, planOb.Interval)
			}
			if err = s.featureService.AddPlan(ctx, planOb.ID, featureOb); err != nil {
				return err
			}
		}
	}

	return nil
}
