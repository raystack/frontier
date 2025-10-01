package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type ProductService interface {
	GetByID(ctx context.Context, id string) (product.Product, error)
	Create(ctx context.Context, product product.Product) (product.Product, error)
	Update(ctx context.Context, product product.Product) (product.Product, error)
	List(ctx context.Context, filter product.Filter) ([]product.Product, error)
	UpsertFeature(ctx context.Context, feature product.Feature) (product.Feature, error)
	GetFeatureByID(ctx context.Context, id string) (product.Feature, error)
	ListFeatures(ctx context.Context, filter product.Filter) ([]product.Feature, error)
}

func (h *ConnectHandler) ListProducts(ctx context.Context, request *connect.Request[frontierv1beta1.ListProductsRequest]) (*connect.Response[frontierv1beta1.ListProductsResponse], error) {
	var products []*frontierv1beta1.Product
	productsList, err := h.productService.List(ctx, product.Filter{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range productsList {
		productPB, err := transformProductToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		products = append(products, productPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListProductsResponse{
		Products: products,
	}), nil
}

func (h *ConnectHandler) GetProduct(ctx context.Context, request *connect.Request[frontierv1beta1.GetProductRequest]) (*connect.Response[frontierv1beta1.GetProductResponse], error) {
	product, err := h.productService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	productPB, err := transformProductToPB(product)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetProductResponse{
		Product: productPB,
	}), nil
}

func (h *ConnectHandler) CreateProduct(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProductRequest]) (*connect.Response[frontierv1beta1.CreateProductResponse], error) {
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	productPB, err := transformProductToPB(newProduct)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateProductResponse{
		Product: productPB,
	}), nil
}

func (h *ConnectHandler) UpdateProduct(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProductRequest]) (*connect.Response[frontierv1beta1.UpdateProductResponse], error) {
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	productPb, err := transformProductToPB(updatedProduct)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.UpdateProductResponse{
		Product: productPb,
	}), nil
}
