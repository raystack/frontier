package v1beta1

import (
	"context"

	"github.com/raystack/frontier/pkg/metadata"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/feature"
	"github.com/raystack/frontier/billing/plan"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PlanService interface {
	GetByID(ctx context.Context, id string) (plan.Plan, error)
	Create(ctx context.Context, plan plan.Plan) (plan.Plan, error)
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	UpsertPlans(ctx context.Context, planFile plan.File) error
}

func (h Handler) ListPlans(ctx context.Context, request *frontierv1beta1.ListPlansRequest) (*frontierv1beta1.ListPlansResponse, error) {
	logger := grpczap.Extract(ctx)

	var plans []*frontierv1beta1.Plan
	planList, err := h.planService.List(ctx, plan.Filter{})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range planList {
		planPB, err := transformPlanToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		plans = append(plans, planPB)
	}

	return &frontierv1beta1.ListPlansResponse{
		Plans: plans,
	}, nil
}

func (h Handler) CreatePlan(ctx context.Context, request *frontierv1beta1.CreatePlanRequest) (*frontierv1beta1.CreatePlanResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())
	// parse features
	features := []feature.Feature{}
	for _, v := range request.GetBody().GetFeatures() {
		featurePrices := []feature.Price{}
		for _, price := range v.GetPrices() {
			featurePrices = append(featurePrices, feature.Price{
				Name:             price.GetName(),
				Amount:           price.GetAmount(),
				Currency:         price.GetCurrency(),
				UsageType:        feature.BuildPriceUsageType(price.GetUsageType()),
				BillingScheme:    feature.BuildBillingScheme(price.GetBillingScheme()),
				MeteredAggregate: price.GetMeteredAggregate(),
				Metadata:         metadata.Build(price.GetMetadata().AsMap()),
			})
		}
		features = append(features, feature.Feature{
			ID:           v.GetId(),
			Name:         v.GetName(),
			Title:        v.GetTitle(),
			Description:  v.GetDescription(),
			Prices:       featurePrices,
			Interval:     v.GetInterval(),
			CreditAmount: v.GetCreditAmount(),
			Metadata:     metadata.Build(v.GetMetadata().AsMap()),
		})
	}
	planToCreate := plan.Plan{
		Name:        request.GetBody().GetName(),
		Description: request.GetBody().GetDescription(),
		Interval:    request.GetBody().GetInterval(),
		Features:    features,
		Metadata:    metaDataMap,
	}

	err := h.planService.UpsertPlans(ctx, plan.File{
		Plans:    []plan.Plan{planToCreate},
		Features: features,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	newPlan, err := h.planService.GetByID(ctx, planToCreate.Name)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	planPB, err := transformPlanToPB(newPlan)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.CreatePlanResponse{
		Plan: planPB,
	}, nil
}

func (h Handler) GetPlan(ctx context.Context, request *frontierv1beta1.GetPlanRequest) (*frontierv1beta1.GetPlanResponse, error) {
	logger := grpczap.Extract(ctx)

	planOb, err := h.planService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	planPB, err := transformPlanToPB(planOb)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.GetPlanResponse{
		Plan: planPB,
	}, nil
}

func transformPlanToPB(p plan.Plan) (*frontierv1beta1.Plan, error) {
	metaData, err := p.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Plan{}, err
	}
	var features []*frontierv1beta1.Feature
	for _, v := range p.Features {
		featurePB, err := transformFeatureToPB(v)
		if err != nil {
			return nil, err
		}
		features = append(features, featurePB)
	}

	return &frontierv1beta1.Plan{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Interval:    p.Interval,
		Features:    features,
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}, nil
}

func transformFeatureToPB(f feature.Feature) (*frontierv1beta1.Feature, error) {
	metaData, err := f.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Feature{}, err
	}

	pricePBs := make([]*frontierv1beta1.Price, len(f.Prices))
	for i, v := range f.Prices {
		pricePB, err := transformPriceToPB(v)
		if err != nil {
			return nil, err
		}
		pricePBs[i] = pricePB
	}

	return &frontierv1beta1.Feature{
		Id:           f.ID,
		Name:         f.Name,
		Title:        f.Title,
		Description:  f.Description,
		PlanIds:      f.PlanIDs,
		State:        f.State,
		Prices:       pricePBs,
		Interval:     f.Interval,
		CreditAmount: f.CreditAmount,
		Metadata:     metaData,
		CreatedAt:    timestamppb.New(f.CreatedAt),
		UpdatedAt:    timestamppb.New(f.UpdatedAt),
	}, nil
}

func transformPriceToPB(p feature.Price) (*frontierv1beta1.Price, error) {
	metaData, err := p.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Price{}, err
	}

	return &frontierv1beta1.Price{
		Id:               p.ID,
		FeatureId:        p.FeatureID,
		ProviderId:       p.ProviderID,
		Name:             p.Name,
		UsageType:        string(p.UsageType),
		BillingScheme:    string(p.BillingScheme),
		State:            p.State,
		Currency:         p.Currency,
		Amount:           p.Amount,
		MeteredAggregate: p.MeteredAggregate,
		TierMode:         p.TierMode,
		Metadata:         metaData,
		CreatedAt:        timestamppb.New(p.CreatedAt),
		UpdatedAt:        timestamppb.New(p.UpdatedAt),
	}, nil
}
