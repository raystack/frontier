package v1beta1

import (
	"context"

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
}

type FeatureService interface {
	GetByID(ctx context.Context, id string) (feature.Feature, error)
	Create(ctx context.Context, feature feature.Feature) (feature.Feature, error)
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

	pricePB, err := transformPriceToPB(f.Price)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Feature{
		Id:          f.ID,
		Name:        f.Name,
		Title:       f.Title,
		Description: f.Description,
		PlanId:      f.PlanID,
		State:       f.State,
		Price:       pricePB,
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(f.CreatedAt),
		UpdatedAt:   timestamppb.New(f.UpdatedAt),
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
