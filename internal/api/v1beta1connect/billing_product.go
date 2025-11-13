package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
)

func (h *ConnectHandler) ListProducts(ctx context.Context, request *connect.Request[frontierv1beta1.ListProductsRequest]) (*connect.Response[frontierv1beta1.ListProductsResponse], error) {
	errorLogger := NewErrorLogger()

	var products []*frontierv1beta1.Product
	productsList, err := h.productService.List(ctx, product.Filter{})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProducts.List", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range productsList {
		productPB, err := transformProductToPB(v)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProducts", v.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		products = append(products, productPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListProductsResponse{
		Products: products,
	}), nil
}

func (h *ConnectHandler) GetProduct(ctx context.Context, request *connect.Request[frontierv1beta1.GetProductRequest]) (*connect.Response[frontierv1beta1.GetProductResponse], error) {
	errorLogger := NewErrorLogger()

	product, err := h.productService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetProduct.GetByID", err,
			zap.String("product_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	productPB, err := transformProductToPB(product)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetProduct", product.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetProductResponse{
		Product: productPB,
	}), nil
}

func (h *ConnectHandler) CreateProduct(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProductRequest]) (*connect.Response[frontierv1beta1.CreateProductResponse], error) {
	errorLogger := NewErrorLogger()

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	// parse price
	var productPrices []product.Price
	for _, v := range request.Msg.GetBody().GetPrices() {
		productPrices = append(productPrices, product.Price{
			Name:             v.GetName(),
			Amount:           v.GetAmount(),
			Currency:         v.GetCurrency(),
			UsageType:        product.BuildPriceUsageType(v.GetUsageType()),
			BillingScheme:    product.BuildBillingScheme(v.GetBillingScheme()),
			MeteredAggregate: v.GetMeteredAggregate(),
			Interval:         v.GetInterval(),
			Metadata:         metadata.Build(v.GetMetadata().AsMap()),
		})
	}
	// parse features
	var productFeatures []product.Feature
	for _, v := range request.Msg.GetBody().GetFeatures() {
		productFeatures = append(productFeatures, product.Feature{
			Name:       v.GetName(),
			Title:      v.GetTitle(),
			ProductIDs: v.GetProductIds(),
			Metadata:   metadata.Build(v.GetMetadata().AsMap()),
		})
	}

	behaviorConfig := product.BehaviorConfig{}
	if request.Msg.GetBody().GetBehaviorConfig() != nil {
		behaviorConfig = product.BehaviorConfig{
			CreditAmount: request.Msg.GetBody().GetBehaviorConfig().GetCreditAmount(),
			SeatLimit:    request.Msg.GetBody().GetBehaviorConfig().GetSeatLimit(),
			MinQuantity:  request.Msg.GetBody().GetBehaviorConfig().GetMinQuantity(),
			MaxQuantity:  request.Msg.GetBody().GetBehaviorConfig().GetMaxQuantity(),
		}
	}
	newProduct, err := h.productService.Create(ctx, product.Product{
		PlanIDs:     []string{request.Msg.GetBody().GetPlanId()},
		Name:        request.Msg.GetBody().GetName(),
		Title:       request.Msg.GetBody().GetTitle(),
		Description: request.Msg.GetBody().GetDescription(),
		Prices:      productPrices,
		Config:      behaviorConfig,
		Behavior:    product.Behavior(request.Msg.GetBody().GetBehavior()),
		Features:    productFeatures,
		Metadata:    metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateProduct.Create", err,
			zap.String("product_name", request.Msg.GetBody().GetName()),
			zap.String("product_title", request.Msg.GetBody().GetTitle()),
			zap.String("plan_id", request.Msg.GetBody().GetPlanId()),
			zap.String("behavior", request.Msg.GetBody().GetBehavior()),
			zap.Int("price_count", len(productPrices)),
			zap.Int("feature_count", len(productFeatures)))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	productPB, err := transformProductToPB(newProduct)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "CreateProduct", newProduct.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateProductResponse{
		Product: productPB,
	}), nil
}

func (h *ConnectHandler) UpdateProduct(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProductRequest]) (*connect.Response[frontierv1beta1.UpdateProductResponse], error) {
	errorLogger := NewErrorLogger()

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	// parse price
	var productPrices []product.Price
	for _, v := range request.Msg.GetBody().GetPrices() {
		productPrices = append(productPrices, product.Price{
			ID:       v.GetId(),
			Name:     v.GetName(),
			Metadata: metadata.Build(v.GetMetadata().AsMap()),
		})
	}
	// parse features
	var productFeatures []product.Feature
	for _, v := range request.Msg.GetBody().GetFeatures() {
		productFeatures = append(productFeatures, product.Feature{
			ID:         v.GetId(),
			Name:       v.GetName(),
			Title:      v.GetTitle(),
			ProductIDs: v.GetProductIds(),
			Metadata:   metadata.Build(v.GetMetadata().AsMap()),
		})
	}
	behaviorConfig := product.BehaviorConfig{}
	if request.Msg.GetBody().GetBehaviorConfig() != nil {
		behaviorConfig = product.BehaviorConfig{
			CreditAmount: request.Msg.GetBody().GetBehaviorConfig().GetCreditAmount(),
			SeatLimit:    request.Msg.GetBody().GetBehaviorConfig().GetSeatLimit(),
			MinQuantity:  request.Msg.GetBody().GetBehaviorConfig().GetMinQuantity(),
			MaxQuantity:  request.Msg.GetBody().GetBehaviorConfig().GetMaxQuantity(),
		}
	}
	updatedProduct, err := h.productService.Update(ctx, product.Product{
		ID:          request.Msg.GetId(),
		Name:        request.Msg.GetBody().GetName(),
		Title:       request.Msg.GetBody().GetTitle(),
		Description: request.Msg.GetBody().GetDescription(),
		Behavior:    product.Behavior(request.Msg.GetBody().GetBehavior()),
		Config:      behaviorConfig,
		Prices:      productPrices,
		Features:    productFeatures,
		Metadata:    metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdateProduct.Update", err,
			zap.String("product_id", request.Msg.GetId()),
			zap.String("product_name", request.Msg.GetBody().GetName()),
			zap.String("product_title", request.Msg.GetBody().GetTitle()),
			zap.String("behavior", request.Msg.GetBody().GetBehavior()),
			zap.Int("price_count", len(productPrices)),
			zap.Int("feature_count", len(productFeatures)))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	productPb, err := transformProductToPB(updatedProduct)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "UpdateProduct", updatedProduct.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.UpdateProductResponse{
		Product: productPb,
	}), nil
}

func (h *ConnectHandler) ListFeatures(ctx context.Context, request *connect.Request[frontierv1beta1.ListFeaturesRequest]) (*connect.Response[frontierv1beta1.ListFeaturesResponse], error) {
	errorLogger := NewErrorLogger()

	features, err := h.productService.ListFeatures(ctx, product.Filter{})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListFeatures.ListFeatures", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var featuresPB []*frontierv1beta1.Feature
	for _, v := range features {
		f, err := transformFeatureToPB(v)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListFeatures", v.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		featuresPB = append(featuresPB, f)
	}

	return connect.NewResponse(&frontierv1beta1.ListFeaturesResponse{
		Features: featuresPB,
	}), nil
}

func (h *ConnectHandler) CreateFeature(ctx context.Context, request *connect.Request[frontierv1beta1.CreateFeatureRequest]) (*connect.Response[frontierv1beta1.CreateFeatureResponse], error) {
	errorLogger := NewErrorLogger()

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	newFeature, err := h.productService.UpsertFeature(ctx, product.Feature{
		Name:       request.Msg.GetBody().GetName(),
		Title:      request.Msg.GetBody().GetTitle(),
		ProductIDs: request.Msg.GetBody().GetProductIds(),
		Metadata:   metaDataMap,
	})
	if err != nil {
		if errors.Is(err, product.ErrInvalidFeatureDetail) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		errorLogger.LogServiceError(ctx, request, "CreateFeature.UpsertFeature", err,
			zap.String("feature_name", request.Msg.GetBody().GetName()),
			zap.String("feature_title", request.Msg.GetBody().GetTitle()),
			zap.Strings("product_ids", request.Msg.GetBody().GetProductIds()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	featurePB, err := transformFeatureToPB(newFeature)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "CreateFeature", newFeature.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateFeatureResponse{
		Feature: featurePB,
	}), nil
}

func (h *ConnectHandler) UpdateFeature(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateFeatureRequest]) (*connect.Response[frontierv1beta1.UpdateFeatureResponse], error) {
	errorLogger := NewErrorLogger()

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	updatedFeature, err := h.productService.UpsertFeature(ctx, product.Feature{
		ID:         request.Msg.GetId(),
		Name:       request.Msg.GetBody().GetName(),
		Title:      request.Msg.GetBody().GetTitle(),
		ProductIDs: request.Msg.GetBody().GetProductIds(),
		Metadata:   metaDataMap,
	})
	if err != nil {
		if errors.Is(err, product.ErrInvalidFeatureDetail) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		errorLogger.LogServiceError(ctx, request, "UpdateFeature.UpsertFeature", err,
			zap.String("feature_id", request.Msg.GetId()),
			zap.String("feature_name", request.Msg.GetBody().GetName()),
			zap.String("feature_title", request.Msg.GetBody().GetTitle()),
			zap.Strings("product_ids", request.Msg.GetBody().GetProductIds()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	featurePB, err := transformFeatureToPB(updatedFeature)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "UpdateFeature", updatedFeature.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.UpdateFeatureResponse{
		Feature: featurePB,
	}), nil
}

func (h *ConnectHandler) GetFeature(ctx context.Context, request *connect.Request[frontierv1beta1.GetFeatureRequest]) (*connect.Response[frontierv1beta1.GetFeatureResponse], error) {
	errorLogger := NewErrorLogger()

	feature, err := h.productService.GetFeatureByID(ctx, request.Msg.GetId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetFeature.GetFeatureByID", err,
			zap.String("feature_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	featurePB, err := transformFeatureToPB(feature)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetFeature", feature.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetFeatureResponse{
		Feature: featurePB,
	}), nil
}
