package product_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/product/mocks"
	stripemock "github.com/raystack/frontier/billing/stripetest/mocks"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

func mockService(t *testing.T) (*client.API, *stripemock.Backend, *mocks.Repository, *mocks.PriceRepository, *mocks.FeatureRepository) {
	t.Helper()
	mockBackend := stripemock.NewBackend(t)
	stripeClient := client.New("key_123", &stripe.Backends{
		API: mockBackend,
	})
	mockProductRepo := mocks.NewRepository(t)
	mockPriceRepo := mocks.NewPriceRepository(t)
	mockFeatureRepo := mocks.NewFeatureRepository(t)
	return stripeClient, mockBackend, mockProductRepo, mockPriceRepo, mockFeatureRepo
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	type args struct {
		product product.Product
	}
	tests := []struct {
		name    string
		args    args
		want    product.Product
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should create product in repo and billing provider with no price and features",
			args: args{
				product: product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
				},
			},
			want: product.Product{
				ID:          "1",
				Name:        "product1",
				Description: "product 1",
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, mockStripeBackend, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockProductRepo.EXPECT().Create(ctx, product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
					Behavior:    product.BasicBehavior,
				}).Return(product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
				}, nil)
				mockStripeBackend.EXPECT().Call("POST", "/v1/products", "key_123", &stripe.ProductParams{
					Params: stripe.Params{
						Context: ctx,
					},
					ID:          stripe.String(""),
					Name:        stripe.String(""),
					Description: stripe.String("product 1"),
					Features:    nil,
					Metadata: map[string]string{
						"behavior":      "basic",
						"credit_amount": "0",
						"managed_by":    "frontier",
						"name":          "product1",
						"product_id":    "1",
					},
				}, &stripe.Product{}).Return(nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
		{
			name: "should create product in repo and billing provider with price and features",
			args: args{
				product: product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
					Behavior:    product.BasicBehavior,
					Prices: []product.Price{
						{
							ID:            "1",
							Name:          "price1",
							Amount:        100,
							Currency:      "usd",
							UsageType:     product.PriceUsageTypeLicensed,
							BillingScheme: product.BillingSchemeFlat,
							Interval:      "month",
						},
					},
					Features: []product.Feature{
						{
							ID:   "1",
							Name: "feature1",
						},
					},
				},
			},
			want: product.Product{
				ID:          "1",
				Name:        "product1",
				Description: "product 1",
				Behavior:    product.BasicBehavior,
				Prices: []product.Price{
					{
						ID:            "1",
						Name:          "price1",
						Amount:        100,
						Currency:      "usd",
						UsageType:     product.PriceUsageTypeLicensed,
						BillingScheme: product.BillingSchemeFlat,
						Interval:      "month",
						ProductID:     "1",
					},
				},
				Features: []product.Feature{
					{
						ID:         "1",
						Name:       "feature1",
						ProductIDs: []string{"1"},
					},
				},
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, mockStripeBackend, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockProductRepo.EXPECT().Create(ctx, product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
					Behavior:    product.BasicBehavior,
					Prices: []product.Price{
						{
							ID:               "1",
							Name:             "price1",
							Amount:           100,
							Currency:         "usd",
							BillingScheme:    product.BillingSchemeFlat,
							UsageType:        product.PriceUsageTypeLicensed,
							MeteredAggregate: "sum",
							TierMode:         product.PriceTierModeGraduated,
							Interval:         "month",
						},
					},
					Features: []product.Feature{
						{
							ID:   "1",
							Name: "feature1",
						},
					},
				}).Return(product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
					Behavior:    product.BasicBehavior,
				}, nil)
				mockStripeBackend.EXPECT().Call("POST", "/v1/products", "key_123", &stripe.ProductParams{
					Params: stripe.Params{
						Context: ctx,
					},
					ID:          stripe.String(""),
					Name:        stripe.String(""),
					Description: stripe.String("product 1"),
					Features:    nil,
					Metadata: map[string]string{
						"behavior":      "basic",
						"credit_amount": "0",
						"managed_by":    "frontier",
						"name":          "product1",
						"product_id":    "1",
					},
				}, &stripe.Product{}).Return(nil)

				mockPriceRepo.EXPECT().Create(ctx, product.Price{
					ID:               "1",
					Name:             "price1",
					Amount:           100,
					Currency:         "usd",
					ProductID:        "1",
					BillingScheme:    product.BillingSchemeFlat,
					UsageType:        product.PriceUsageTypeLicensed,
					TierMode:         product.PriceTierModeGraduated,
					Interval:         "month",
					MeteredAggregate: "sum",
				}).Return(product.Price{
					ID:            "1",
					Name:          "price1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					Interval:      "month",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
				}, nil)
				mockStripeBackend.EXPECT().Call("POST", "/v1/prices", "key_123", &stripe.PriceParams{
					Params: stripe.Params{
						Context: ctx,
					},
					Product:    stripe.String("1"),
					Currency:   stripe.String("usd"),
					UnitAmount: stripe.Int64(100),
					Metadata: map[string]string{
						"name":       "price1",
						"product_id": "1",
						"price_id":   "1",
						"managed_by": "frontier",
					},
					BillingScheme: stripe.String("per_unit"),
					Nickname:      stripe.String("price1"),
					Recurring: &stripe.PriceRecurringParams{
						Interval:  stripe.String("month"),
						UsageType: stripe.String("licensed"),
					},
				}, &stripe.Price{
					ID: "",
				}).Return(nil)

				mockFeatureRepo.EXPECT().GetByName(ctx, "feature1").Return(product.Feature{
					ID:   "1",
					Name: "feature1",
				}, nil)
				mockFeatureRepo.EXPECT().UpdateByName(ctx, product.Feature{
					ID:         "1",
					Name:       "feature1",
					ProductIDs: []string{"1"},
				}).Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					ProductIDs: []string{"1"},
				}, nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(ctx, tt.args.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-)want (+)got:\n%s", diff)
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    product.Product
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should get product by name",
			args: args{
				id: "product1",
			},
			want: product.Product{
				ID:          "1",
				Name:        "product1",
				Description: "product 1",
				Behavior:    product.BasicBehavior,
				Prices: []product.Price{
					{
						ID:        "1",
						Name:      "price1",
						Amount:    100,
						Currency:  "usd",
						UsageType: product.PriceUsageTypeLicensed,
					},
				},
				Features: []product.Feature{
					{
						ID:   "1",
						Name: "feature1",
					},
				},
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, _, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockProductRepo.EXPECT().GetByName(ctx, "product1").Return(product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
					Behavior:    product.BasicBehavior,
				}, nil)

				mockPriceRepo.EXPECT().List(ctx, product.Filter{
					ProductID: "1",
				}).Return([]product.Price{
					{
						ID:        "1",
						Name:      "price1",
						Amount:    100,
						Currency:  "usd",
						UsageType: product.PriceUsageTypeLicensed,
					},
				}, nil)

				mockFeatureRepo.EXPECT().List(ctx, product.Filter{
					ProductID: "1",
				}).Return([]product.Feature{
					{
						ID:   "1",
						Name: "feature1",
					},
				}, nil)

				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.GetByID(ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()
	type args struct {
		product product.Product
	}
	tests := []struct {
		name    string
		args    args
		want    product.Product
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should update product description in repo and billing provider",
			args: args{
				product: product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1 new description",
					Behavior:    product.BasicBehavior,
				},
			},
			want: product.Product{
				ID:          "1",
				Name:        "product1",
				Description: "product 1 new description",
				Behavior:    product.BasicBehavior,
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, mockStripeBackend, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)

				mockProductRepo.EXPECT().GetByID(ctx, "1").Return(product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1",
					Behavior:    product.BasicBehavior,
				}, nil)
				mockProductRepo.EXPECT().UpdateByName(ctx, product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1 new description",
					Behavior:    product.BasicBehavior,
				}).Return(product.Product{
					ID:          "1",
					Name:        "product1",
					Description: "product 1 new description",
					Behavior:    product.BasicBehavior,
				}, nil)
				mockStripeBackend.EXPECT().Call("POST", "/v1/products/", "key_123", &stripe.ProductParams{
					Params: stripe.Params{
						Context: ctx,
					},
					Name:        stripe.String(""),
					Description: stripe.String("product 1 new description"),
					Features:    nil,
					Metadata: map[string]string{
						"behavior":   "basic",
						"managed_by": "frontier",
						"name":       "product1",
						"product_id": "1",
						"plan_ids":   "",
					},
				}, &stripe.Product{}).Return(nil)

				mockPriceRepo.EXPECT().List(ctx, product.Filter{
					ProductID: "1",
				}).Return([]product.Price{}, nil)

				mockFeatureRepo.EXPECT().List(ctx, product.Filter{
					ProductID: "1",
				}).Return([]product.Feature{}, nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Update(ctx, tt.args.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := reflect.DeepEqual(tt.want, got); diff {
				t.Errorf("Update() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_CreatePrice(t *testing.T) {
	ctx := context.Background()
	type args struct {
		price product.Price
	}
	tests := []struct {
		name    string
		args    args
		want    product.Price
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should create price in repo and billing provider",
			args: args{
				price: product.Price{
					ID:            "1",
					Name:          "price1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
					Interval:      "month",
				},
			},
			want: product.Price{
				ID:            "1",
				Name:          "price1",
				Amount:        100,
				Currency:      "usd",
				ProductID:     "1",
				BillingScheme: product.BillingSchemeFlat,
				UsageType:     product.PriceUsageTypeLicensed,
				Interval:      "month",
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, mockStripeBackend, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockPriceRepo.EXPECT().Create(ctx, product.Price{
					ID:            "1",
					Name:          "price1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
					Interval:      "month",
				}).Return(product.Price{
					ID:            "1",
					Name:          "price1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
					Interval:      "month",
				}, nil)
				mockStripeBackend.EXPECT().Call("POST", "/v1/prices", "key_123", &stripe.PriceParams{
					Params: stripe.Params{
						Context: ctx,
					},
					Product:    stripe.String("1"),
					Currency:   stripe.String("usd"),
					UnitAmount: stripe.Int64(100),
					Metadata: map[string]string{
						"name":       "price1",
						"product_id": "1",
						"price_id":   "1",
						"managed_by": "frontier",
					},
					BillingScheme: stripe.String("per_unit"),
					Nickname:      stripe.String("price1"),
					Recurring: &stripe.PriceRecurringParams{
						Interval:  stripe.String("month"),
						UsageType: stripe.String("licensed"),
					},
				}, &stripe.Price{
					ID: "",
				}).Return(nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.CreatePrice(ctx, tt.args.price)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CreatePrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_UpdatePrice(t *testing.T) {
	ctx := context.Background()
	type args struct {
		price product.Price
	}
	tests := []struct {
		name    string
		args    args
		want    product.Price
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should update price name in repo and billing provider",
			args: args{
				price: product.Price{
					ID:            "1",
					Name:          "price1.1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
				},
			},
			want: product.Price{
				ID:            "1",
				Name:          "price1.1",
				Amount:        100,
				Currency:      "usd",
				ProductID:     "1",
				BillingScheme: product.BillingSchemeFlat,
				UsageType:     product.PriceUsageTypeLicensed,
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, mockStripeBackend, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockPriceRepo.EXPECT().GetByID(ctx, "1").Return(product.Price{
					ID:            "1",
					Name:          "price1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
				}, nil)

				mockPriceRepo.EXPECT().UpdateByID(ctx, product.Price{
					ID:            "1",
					Name:          "price1.1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
				}).Return(product.Price{
					ID:            "1",
					Name:          "price1.1",
					Amount:        100,
					Currency:      "usd",
					ProductID:     "1",
					BillingScheme: product.BillingSchemeFlat,
					UsageType:     product.PriceUsageTypeLicensed,
				}, nil)

				mockStripeBackend.EXPECT().Call("POST", "/v1/prices/", "key_123", &stripe.PriceParams{
					Params: stripe.Params{
						Context: ctx,
					},
					Metadata: map[string]string{
						"name":       "price1.1",
						"product_id": "1",
						"price_id":   "1",
						"managed_by": "frontier",
					},
					Nickname: stripe.String("price1.1"),
				}, &stripe.Price{
					ID: "",
				}).Return(nil)

				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.UpdatePrice(ctx, tt.args.price)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("UpdatePrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_UpsertFeature(t *testing.T) {
	ctx := context.Background()
	type args struct {
		feature product.Feature
	}
	tests := []struct {
		name    string
		args    args
		want    product.Feature
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should create feature in repo if doesn't exists",
			args: args{
				feature: product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				},
			},
			want: product.Feature{
				ID:         "1",
				Name:       "feature1",
				Title:      "Feature 1",
				ProductIDs: []string{"1"},
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, _, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockFeatureRepo.EXPECT().GetByName(ctx, "feature1").Return(product.Feature{}, product.ErrFeatureNotFound)
				mockFeatureRepo.EXPECT().Create(ctx, product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				}).Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				}, nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
		{
			name: "should update feature in repo if already exists",
			args: args{
				feature: product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1.1",
					ProductIDs: []string{"1"},
				},
			},
			want: product.Feature{
				ID:         "1",
				Name:       "feature1",
				Title:      "Feature 1.1",
				ProductIDs: []string{"1"},
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, _, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockFeatureRepo.EXPECT().GetByName(ctx, "feature1").Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				}, nil)
				mockFeatureRepo.EXPECT().UpdateByName(ctx, product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1.1",
					ProductIDs: []string{"1"},
				}).Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1.1",
					ProductIDs: []string{"1"},
				}, nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.UpsertFeature(ctx, tt.args.feature)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpsertFeature() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_AddFeatureToProduct(t *testing.T) {
	ctx := context.Background()
	type args struct {
		feature   product.Feature
		productID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should add feature to product",
			args: args{
				feature: product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				},
				productID: "2",
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, _, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockFeatureRepo.EXPECT().GetByName(ctx, "feature1").Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				}, nil)
				mockFeatureRepo.EXPECT().UpdateByName(ctx, product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1", "2"},
				}).Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1", "2"},
				}, nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.AddFeatureToProduct(ctx, tt.args.feature, tt.args.productID); (err != nil) != tt.wantErr {
				t.Errorf("AddFeatureToProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_RemoveFeatureFromProduct(t *testing.T) {
	ctx := context.Background()
	type args struct {
		featureID string
		productID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func() *product.Service
	}{
		{
			name: "should remove feature from product",
			args: args{
				featureID: "1",
				productID: "2",
			},
			wantErr: false,
			setup: func() *product.Service {
				stripeClient, _, mockProductRepo, mockPriceRepo, mockFeatureRepo := mockService(t)
				mockFeatureRepo.EXPECT().GetByName(ctx, "1").Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1", "2"},
				}, nil)
				mockFeatureRepo.EXPECT().UpdateByName(ctx, product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				}).Return(product.Feature{
					ID:         "1",
					Name:       "feature1",
					Title:      "Feature 1",
					ProductIDs: []string{"1"},
				}, nil)
				return product.NewService(stripeClient, mockProductRepo, mockPriceRepo, mockFeatureRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.RemoveFeatureFromProduct(ctx, tt.args.featureID, tt.args.productID); (err != nil) != tt.wantErr {
				t.Errorf("RemoveFeatureFromProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
