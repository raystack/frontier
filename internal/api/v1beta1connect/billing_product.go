package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/product"
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
