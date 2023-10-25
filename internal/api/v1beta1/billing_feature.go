package v1beta1

import (
	"context"

	"github.com/raystack/frontier/pkg/metadata"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/feature"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type FeatureService interface {
	GetByID(ctx context.Context, id string) (feature.Feature, error)
	Create(ctx context.Context, feature feature.Feature) (feature.Feature, error)
	Update(ctx context.Context, feature feature.Feature) (feature.Feature, error)
	List(ctx context.Context, filter feature.Filter) ([]feature.Feature, error)
}

func (h Handler) CreateFeature(ctx context.Context, request *frontierv1beta1.CreateFeatureRequest) (*frontierv1beta1.CreateFeatureResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())
	// parse price
	featurePrices := []feature.Price{}
	for _, v := range request.GetBody().GetPrices() {
		featurePrices = append(featurePrices, feature.Price{
			Name:             v.GetName(),
			Amount:           v.GetAmount(),
			Currency:         v.GetCurrency(),
			UsageType:        feature.BuildPriceUsageType(v.GetUsageType()),
			BillingScheme:    feature.BuildBillingScheme(v.GetBillingScheme()),
			MeteredAggregate: v.GetMeteredAggregate(),
			Metadata:         metadata.Build(v.GetMetadata().AsMap()),
		})
	}

	newFeature, err := h.featureService.Create(ctx, feature.Feature{
		PlanIDs:      []string{request.GetBody().GetPlanId()},
		Name:         request.GetBody().GetName(),
		Title:        request.GetBody().GetTitle(),
		Description:  request.GetBody().GetDescription(),
		Prices:       featurePrices,
		Interval:     request.GetBody().GetInterval(),
		CreditAmount: request.GetBody().GetCreditAmount(),
		Metadata:     metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	featurePB, err := transformFeatureToPB(newFeature)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.CreateFeatureResponse{
		Feature: featurePB,
	}, nil
}

func (h Handler) UpdateFeature(ctx context.Context, request *frontierv1beta1.UpdateFeatureRequest) (*frontierv1beta1.UpdateFeatureResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())
	// parse price
	featurePrices := []feature.Price{}
	for _, v := range request.GetBody().GetPrices() {
		featurePrices = append(featurePrices, feature.Price{
			Name:     v.GetName(),
			Metadata: metadata.Build(v.GetMetadata().AsMap()),
		})
	}
	updatedFeature, err := h.featureService.Update(ctx, feature.Feature{
		ID:          request.GetId(),
		Name:        request.GetBody().GetName(),
		Title:       request.GetBody().GetTitle(),
		Description: request.GetBody().GetDescription(),
		Prices:      featurePrices,
		Metadata:    metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	featurePb, err := transformFeatureToPB(updatedFeature)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.UpdateFeatureResponse{
		Feature: featurePb,
	}, nil
}

func (h Handler) ListFeatures(ctx context.Context, request *frontierv1beta1.ListFeaturesRequest) (*frontierv1beta1.ListFeaturesResponse, error) {
	logger := grpczap.Extract(ctx)

	var features []*frontierv1beta1.Feature
	featuresList, err := h.featureService.List(ctx, feature.Filter{})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range featuresList {
		featurePB, err := transformFeatureToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		features = append(features, featurePB)
	}

	return &frontierv1beta1.ListFeaturesResponse{
		Features: features,
	}, nil
}

func (h Handler) GetFeature(ctx context.Context, request *frontierv1beta1.GetFeatureRequest) (*frontierv1beta1.GetFeatureResponse, error) {
	logger := grpczap.Extract(ctx)

	feature, err := h.featureService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	featurePB, err := transformFeatureToPB(feature)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.GetFeatureResponse{
		Feature: featurePB,
	}, nil
}
