package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_ListProducts(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(ps *mocks.ProductService)
		req         *connect.Request[frontierv1beta1.ListProductsRequest]
		want        *connect.Response[frontierv1beta1.ListProductsResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if service returns error",
			setup: func(ps *mocks.ProductService) {
				ps.EXPECT().List(mock.Anything, product.Filter{}).Return(nil, errors.New("service error"))
			},
			req:         connect.NewRequest(&frontierv1beta1.ListProductsRequest{}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return empty list if no products found",
			setup: func(ps *mocks.ProductService) {
				ps.EXPECT().List(mock.Anything, product.Filter{}).Return([]product.Product{}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListProductsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListProductsResponse{
				Products: []*frontierv1beta1.Product{},
			}),
			wantErr: false,
		},
		{
			name: "should return list of products successfully",
			setup: func(ps *mocks.ProductService) {
				createdAt := time.Now()
				updatedAt := time.Now()
				ps.EXPECT().List(mock.Anything, product.Filter{}).Return([]product.Product{
					{
						ID:          "product-1",
						Name:        "Basic Plan",
						Title:       "Basic Plan",
						Description: "Basic subscription plan",
						PlanIDs:     []string{"plan-1"},
						State:       "active",
						Prices: []product.Price{
							{
								ID:               "price-1",
								ProductID:        "product-1",
								Name:             "Monthly Price",
								Amount:           1000,
								Currency:         "usd",
								UsageType:        product.PriceUsageTypeLicensed,
								BillingScheme:    product.BillingSchemeTiered,
								State:            "active",
								Interval:         "month",
								MeteredAggregate: "sum",
								Metadata:         metadata.Metadata{"key": "value"},
								CreatedAt:        createdAt,
								UpdatedAt:        updatedAt,
							},
						},
						Features: []product.Feature{
							{
								ID:         "feature-1",
								Name:       "feature-basic",
								Title:      "Basic Feature",
								ProductIDs: []string{"product-1"},
								Metadata:   metadata.Metadata{"type": "basic"},
								CreatedAt:  createdAt,
								UpdatedAt:  updatedAt,
							},
						},
						Config: product.BehaviorConfig{
							SeatLimit:    10,
							CreditAmount: 100,
							MinQuantity:  1,
							MaxQuantity:  100,
						},
						Behavior:  product.PerSeatBehavior,
						Metadata:  metadata.Metadata{"category": "subscription"},
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
					{
						ID:          "product-2",
						Name:        "Premium Plan",
						Title:       "Premium Plan",
						Description: "Premium subscription plan",
						PlanIDs:     []string{"plan-2"},
						State:       "active",
						Prices:      []product.Price{},
						Features:    []product.Feature{},
						Config: product.BehaviorConfig{
							SeatLimit:    50,
							CreditAmount: 500,
							MinQuantity:  1,
							MaxQuantity:  500,
						},
						Behavior:  product.CreditBehavior,
						Metadata:  metadata.Metadata{"category": "premium"},
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListProductsRequest{}),
			want: func() *connect.Response[frontierv1beta1.ListProductsResponse] {
				createdAt := time.Now()
				updatedAt := time.Now()
				return connect.NewResponse(&frontierv1beta1.ListProductsResponse{
					Products: []*frontierv1beta1.Product{
						{
							Id:          "product-1",
							Name:        "Basic Plan",
							Title:       "Basic Plan",
							Description: "Basic subscription plan",
							PlanIds:     []string{"plan-1"},
							State:       "active",
							Prices: []*frontierv1beta1.Price{
								{
									Id:               "price-1",
									ProductId:        "product-1",
									Name:             "Monthly Price",
									Amount:           1000,
									Currency:         "usd",
									UsageType:        string(product.PriceUsageTypeLicensed),
									BillingScheme:    string(product.BillingSchemeTiered),
									State:            "active",
									Interval:         "month",
									MeteredAggregate: "sum",
									TierMode:         "volume",
									CreatedAt:        timestamppb.New(createdAt),
									UpdatedAt:        timestamppb.New(updatedAt),
								},
							},
							Features: []*frontierv1beta1.Feature{
								{
									Id:         "feature-1",
									Name:       "feature-basic",
									Title:      "Basic Feature",
									ProductIds: []string{"product-1"},
									CreatedAt:  timestamppb.New(createdAt),
									UpdatedAt:  timestamppb.New(updatedAt),
								},
							},
							BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
								SeatLimit:    10,
								CreditAmount: 100,
								MinQuantity:  1,
								MaxQuantity:  100,
							},
							Behavior:  product.PerSeatBehavior.String(),
							CreatedAt: timestamppb.New(createdAt),
							UpdatedAt: timestamppb.New(updatedAt),
						},
						{
							Id:          "product-2",
							Name:        "Premium Plan",
							Title:       "Premium Plan",
							Description: "Premium subscription plan",
							PlanIds:     []string{"plan-2"},
							State:       "active",
							Prices:      []*frontierv1beta1.Price{},
							Features:    []*frontierv1beta1.Feature{},
							BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
								SeatLimit:    50,
								CreditAmount: 500,
								MinQuantity:  1,
								MaxQuantity:  500,
							},
							Behavior:  product.CreditBehavior.String(),
							CreatedAt: timestamppb.New(createdAt),
							UpdatedAt: timestamppb.New(updatedAt),
						},
					},
				})
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productService := mocks.NewProductService(t)
			if tt.setup != nil {
				tt.setup(productService)
			}
			h := &ConnectHandler{
				productService: productService,
			}
			got, err := h.ListProducts(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, len(tt.want.Msg.GetProducts()), len(got.Msg.GetProducts()))
				for i, wantProduct := range tt.want.Msg.GetProducts() {
					gotProduct := got.Msg.GetProducts()[i]
					assert.Equal(t, wantProduct.GetId(), gotProduct.GetId())
					assert.Equal(t, wantProduct.GetName(), gotProduct.GetName())
					assert.Equal(t, wantProduct.GetTitle(), gotProduct.GetTitle())
					assert.Equal(t, wantProduct.GetDescription(), gotProduct.GetDescription())
					assert.Equal(t, wantProduct.GetPlanIds(), gotProduct.GetPlanIds())
					assert.Equal(t, wantProduct.GetState(), gotProduct.GetState())
					assert.Equal(t, len(wantProduct.GetPrices()), len(gotProduct.GetPrices()))
					assert.Equal(t, len(wantProduct.GetFeatures()), len(gotProduct.GetFeatures()))
					assert.Equal(t, wantProduct.GetBehavior(), gotProduct.GetBehavior())
					if wantProduct.GetBehaviorConfig() != nil {
						assert.Equal(t, wantProduct.GetBehaviorConfig().GetSeatLimit(), gotProduct.GetBehaviorConfig().GetSeatLimit())
						assert.Equal(t, wantProduct.GetBehaviorConfig().GetCreditAmount(), gotProduct.GetBehaviorConfig().GetCreditAmount())
						assert.Equal(t, wantProduct.GetBehaviorConfig().GetMinQuantity(), gotProduct.GetBehaviorConfig().GetMinQuantity())
						assert.Equal(t, wantProduct.GetBehaviorConfig().GetMaxQuantity(), gotProduct.GetBehaviorConfig().GetMaxQuantity())
					}
				}
			}
		})
	}
}

func TestConnectHandler_GetProduct(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(ps *mocks.ProductService)
		req         *connect.Request[frontierv1beta1.GetProductRequest]
		want        *connect.Response[frontierv1beta1.GetProductResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if service returns error",
			setup: func(ps *mocks.ProductService) {
				ps.EXPECT().GetByID(mock.Anything, "product-1").Return(product.Product{}, errors.New("service error"))
			},
			req:         connect.NewRequest(&frontierv1beta1.GetProductRequest{Id: "product-1"}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return product successfully with minimal data",
			setup: func(ps *mocks.ProductService) {
				createdAt := time.Now()
				updatedAt := time.Now()
				ps.EXPECT().GetByID(mock.Anything, "product-1").Return(product.Product{
					ID:          "product-1",
					Name:        "Basic Plan",
					Title:       "Basic Plan",
					Description: "Basic subscription plan",
					PlanIDs:     []string{"plan-1"},
					State:       "active",
					Prices:      []product.Price{},
					Features:    []product.Feature{},
					Config: product.BehaviorConfig{
						SeatLimit:    10,
						CreditAmount: 100,
						MinQuantity:  1,
						MaxQuantity:  100,
					},
					Behavior:  product.BasicBehavior,
					Metadata:  metadata.Metadata{"category": "basic"},
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.GetProductRequest{Id: "product-1"}),
			want: func() *connect.Response[frontierv1beta1.GetProductResponse] {
				createdAt := time.Now()
				updatedAt := time.Now()
				return connect.NewResponse(&frontierv1beta1.GetProductResponse{
					Product: &frontierv1beta1.Product{
						Id:          "product-1",
						Name:        "Basic Plan",
						Title:       "Basic Plan",
						Description: "Basic subscription plan",
						PlanIds:     []string{"plan-1"},
						State:       "active",
						Prices:      []*frontierv1beta1.Price{},
						Features:    []*frontierv1beta1.Feature{},
						BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
							SeatLimit:    10,
							CreditAmount: 100,
							MinQuantity:  1,
							MaxQuantity:  100,
						},
						Behavior:  product.BasicBehavior.String(),
						CreatedAt: timestamppb.New(createdAt),
						UpdatedAt: timestamppb.New(updatedAt),
					},
				})
			}(),
			wantErr: false,
		},
		{
			name: "should return product successfully with complex data",
			setup: func(ps *mocks.ProductService) {
				createdAt := time.Now()
				updatedAt := time.Now()
				ps.EXPECT().GetByID(mock.Anything, "product-2").Return(product.Product{
					ID:          "product-2",
					Name:        "Premium Plan",
					Title:       "Premium Plan",
					Description: "Premium subscription plan with features",
					PlanIDs:     []string{"plan-2", "plan-3"},
					State:       "active",
					Prices: []product.Price{
						{
							ID:               "price-1",
							ProductID:        "product-2",
							Name:             "Monthly Premium Price",
							Amount:           5000,
							Currency:         "usd",
							UsageType:        product.PriceUsageTypeLicensed,
							BillingScheme:    product.BillingSchemeTiered,
							State:            "active",
							Interval:         "month",
							MeteredAggregate: "sum",
							Metadata:         metadata.Metadata{"tier": "premium"},
							CreatedAt:        createdAt,
							UpdatedAt:        updatedAt,
						},
					},
					Features: []product.Feature{
						{
							ID:         "feature-1",
							Name:       "premium-feature",
							Title:      "Premium Feature",
							ProductIDs: []string{"product-2"},
							Metadata:   metadata.Metadata{"type": "premium"},
							CreatedAt:  createdAt,
							UpdatedAt:  updatedAt,
						},
					},
					Config: product.BehaviorConfig{
						SeatLimit:    50,
						CreditAmount: 500,
						MinQuantity:  2,
						MaxQuantity:  200,
					},
					Behavior:  product.PerSeatBehavior,
					Metadata:  metadata.Metadata{"category": "premium", "tier": "enterprise"},
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.GetProductRequest{Id: "product-2"}),
			want: func() *connect.Response[frontierv1beta1.GetProductResponse] {
				createdAt := time.Now()
				updatedAt := time.Now()
				return connect.NewResponse(&frontierv1beta1.GetProductResponse{
					Product: &frontierv1beta1.Product{
						Id:          "product-2",
						Name:        "Premium Plan",
						Title:       "Premium Plan",
						Description: "Premium subscription plan with features",
						PlanIds:     []string{"plan-2", "plan-3"},
						State:       "active",
						Prices: []*frontierv1beta1.Price{
							{
								Id:               "price-1",
								ProductId:        "product-2",
								Name:             "Monthly Premium Price",
								Amount:           5000,
								Currency:         "usd",
								UsageType:        string(product.PriceUsageTypeLicensed),
								BillingScheme:    string(product.BillingSchemeTiered),
								State:            "active",
								Interval:         "month",
								MeteredAggregate: "sum",
								TierMode:         "volume",
								CreatedAt:        timestamppb.New(createdAt),
								UpdatedAt:        timestamppb.New(updatedAt),
							},
						},
						Features: []*frontierv1beta1.Feature{
							{
								Id:         "feature-1",
								Name:       "premium-feature",
								Title:      "Premium Feature",
								ProductIds: []string{"product-2"},
								CreatedAt:  timestamppb.New(createdAt),
								UpdatedAt:  timestamppb.New(updatedAt),
							},
						},
						BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
							SeatLimit:    50,
							CreditAmount: 500,
							MinQuantity:  2,
							MaxQuantity:  200,
						},
						Behavior:  product.PerSeatBehavior.String(),
						CreatedAt: timestamppb.New(createdAt),
						UpdatedAt: timestamppb.New(updatedAt),
					},
				})
			}(),
			wantErr: false,
		},
		{
			name: "should handle empty product id",
			setup: func(ps *mocks.ProductService) {
				ps.EXPECT().GetByID(mock.Anything, "").Return(product.Product{}, errors.New("not found"))
			},
			req:         connect.NewRequest(&frontierv1beta1.GetProductRequest{Id: ""}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productService := mocks.NewProductService(t)
			if tt.setup != nil {
				tt.setup(productService)
			}
			h := &ConnectHandler{
				productService: productService,
			}
			got, err := h.GetProduct(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				wantProduct := tt.want.Msg.GetProduct()
				gotProduct := got.Msg.GetProduct()
				assert.Equal(t, wantProduct.GetId(), gotProduct.GetId())
				assert.Equal(t, wantProduct.GetName(), gotProduct.GetName())
				assert.Equal(t, wantProduct.GetTitle(), gotProduct.GetTitle())
				assert.Equal(t, wantProduct.GetDescription(), gotProduct.GetDescription())
				assert.Equal(t, wantProduct.GetPlanIds(), gotProduct.GetPlanIds())
				assert.Equal(t, wantProduct.GetState(), gotProduct.GetState())
				assert.Equal(t, len(wantProduct.GetPrices()), len(gotProduct.GetPrices()))
				assert.Equal(t, len(wantProduct.GetFeatures()), len(gotProduct.GetFeatures()))
				assert.Equal(t, wantProduct.GetBehavior(), gotProduct.GetBehavior())
				if wantProduct.GetBehaviorConfig() != nil {
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetSeatLimit(), gotProduct.GetBehaviorConfig().GetSeatLimit())
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetCreditAmount(), gotProduct.GetBehaviorConfig().GetCreditAmount())
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetMinQuantity(), gotProduct.GetBehaviorConfig().GetMinQuantity())
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetMaxQuantity(), gotProduct.GetBehaviorConfig().GetMaxQuantity())
				}
				// Verify prices details if present
				for i, wantPrice := range wantProduct.GetPrices() {
					if i < len(gotProduct.GetPrices()) {
						gotPrice := gotProduct.GetPrices()[i]
						assert.Equal(t, wantPrice.GetId(), gotPrice.GetId())
						assert.Equal(t, wantPrice.GetProductId(), gotPrice.GetProductId())
						assert.Equal(t, wantPrice.GetName(), gotPrice.GetName())
						assert.Equal(t, wantPrice.GetAmount(), gotPrice.GetAmount())
						assert.Equal(t, wantPrice.GetCurrency(), gotPrice.GetCurrency())
					}
				}
				// Verify features details if present
				for i, wantFeature := range wantProduct.GetFeatures() {
					if i < len(gotProduct.GetFeatures()) {
						gotFeature := gotProduct.GetFeatures()[i]
						assert.Equal(t, wantFeature.GetId(), gotFeature.GetId())
						assert.Equal(t, wantFeature.GetName(), gotFeature.GetName())
						assert.Equal(t, wantFeature.GetTitle(), gotFeature.GetTitle())
						assert.Equal(t, wantFeature.GetProductIds(), gotFeature.GetProductIds())
					}
				}
			}
		})
	}
}
func TestConnectHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(ps *mocks.ProductService)
		req         *connect.Request[frontierv1beta1.CreateProductRequest]
		want        *connect.Response[frontierv1beta1.CreateProductResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if service returns error",
			setup: func(ps *mocks.ProductService) {
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("product.Product")).Return(product.Product{}, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProductRequest{
				Body: &frontierv1beta1.ProductRequestBody{
					Name:        "Test Product",
					Title:       "Test Product Title",
					Description: "Test product description",
					PlanId:      "plan-1",
					Behavior:    product.BasicBehavior.String(),
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should create product successfully with minimal data",
			setup: func(ps *mocks.ProductService) {
				createdAt := time.Now()
				updatedAt := time.Now()
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("product.Product")).Return(product.Product{
					ID:          "product-1",
					Name:        "Basic Product",
					Title:       "Basic Product Title",
					Description: "Basic product description",
					PlanIDs:     []string{"plan-1"},
					State:       "active",
					Prices:      []product.Price{},
					Features:    []product.Feature{},
					Config: product.BehaviorConfig{
						SeatLimit:    10,
						CreditAmount: 100,
						MinQuantity:  1,
						MaxQuantity:  100,
					},
					Behavior:  product.BasicBehavior,
					Metadata:  metadata.Metadata{},
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProductRequest{
				Body: &frontierv1beta1.ProductRequestBody{
					Name:        "Basic Product",
					Title:       "Basic Product Title",
					Description: "Basic product description",
					PlanId:      "plan-1",
					Behavior:    product.BasicBehavior.String(),
					BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
						SeatLimit:    10,
						CreditAmount: 100,
						MinQuantity:  1,
						MaxQuantity:  100,
					},
				},
			}),
			want: func() *connect.Response[frontierv1beta1.CreateProductResponse] {
				createdAt := time.Now()
				updatedAt := time.Now()
				return connect.NewResponse(&frontierv1beta1.CreateProductResponse{
					Product: &frontierv1beta1.Product{
						Id:          "product-1",
						Name:        "Basic Product",
						Title:       "Basic Product Title",
						Description: "Basic product description",
						PlanIds:     []string{"plan-1"},
						State:       "active",
						Prices:      []*frontierv1beta1.Price{},
						Features:    []*frontierv1beta1.Feature{},
						BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
							SeatLimit:    10,
							CreditAmount: 100,
							MinQuantity:  1,
							MaxQuantity:  100,
						},
						Behavior:  product.BasicBehavior.String(),
						CreatedAt: timestamppb.New(createdAt),
						UpdatedAt: timestamppb.New(updatedAt),
					},
				})
			}(),
			wantErr: false,
		},
		{
			name: "should handle empty product data",
			setup: func(ps *mocks.ProductService) {
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("product.Product")).Return(product.Product{}, errors.New("validation error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProductRequest{
				Body: &frontierv1beta1.ProductRequestBody{},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productService := mocks.NewProductService(t)
			if tt.setup != nil {
				tt.setup(productService)
			}
			h := &ConnectHandler{
				productService: productService,
			}
			got, err := h.CreateProduct(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				wantProduct := tt.want.Msg.GetProduct()
				gotProduct := got.Msg.GetProduct()
				assert.Equal(t, wantProduct.GetId(), gotProduct.GetId())
				assert.Equal(t, wantProduct.GetName(), gotProduct.GetName())
				assert.Equal(t, wantProduct.GetTitle(), gotProduct.GetTitle())
				assert.Equal(t, wantProduct.GetDescription(), gotProduct.GetDescription())
				assert.Equal(t, wantProduct.GetPlanIds(), gotProduct.GetPlanIds())
				assert.Equal(t, wantProduct.GetState(), gotProduct.GetState())
				assert.Equal(t, len(wantProduct.GetPrices()), len(gotProduct.GetPrices()))
				assert.Equal(t, len(wantProduct.GetFeatures()), len(gotProduct.GetFeatures()))
				assert.Equal(t, wantProduct.GetBehavior(), gotProduct.GetBehavior())
				if wantProduct.GetBehaviorConfig() != nil {
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetSeatLimit(), gotProduct.GetBehaviorConfig().GetSeatLimit())
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetCreditAmount(), gotProduct.GetBehaviorConfig().GetCreditAmount())
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetMinQuantity(), gotProduct.GetBehaviorConfig().GetMinQuantity())
					assert.Equal(t, wantProduct.GetBehaviorConfig().GetMaxQuantity(), gotProduct.GetBehaviorConfig().GetMaxQuantity())
				}
			}
		})
	}
}
