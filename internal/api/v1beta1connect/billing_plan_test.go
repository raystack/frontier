package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_CreatePlan(t *testing.T) {
	testMetadata := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"key1": {Kind: &structpb.Value_StringValue{StringValue: "value1"}},
			"key2": {Kind: &structpb.Value_StringValue{StringValue: "value2"}},
		},
	}

	tests := []struct {
		name    string
		setup   func(ps *mocks.PlanService)
		request *connect.Request[frontierv1beta1.CreatePlanRequest]
		want    *connect.Response[frontierv1beta1.CreatePlanResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when UpsertPlans fails",
			request: connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
				Body: &frontierv1beta1.PlanRequestBody{
					Name:        "basic-plan",
					Title:       "Basic Plan",
					Description: "A basic plan",
					Interval:    "monthly",
					Products:    []*frontierv1beta1.Product{},
					Metadata:    testMetadata,
				},
			}),
			setup: func(ps *mocks.PlanService) {
				ps.On("UpsertPlans", mock.Anything, mock.Anything).Return(errors.New("upsert failed"))
			},
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return internal server error when GetByID fails after upsert",
			request: connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
				Body: &frontierv1beta1.PlanRequestBody{
					Name:        "basic-plan",
					Title:       "Basic Plan",
					Description: "A basic plan",
					Interval:    "monthly",
					Products:    []*frontierv1beta1.Product{},
					Metadata:    testMetadata,
				},
			}),
			setup: func(ps *mocks.PlanService) {
				ps.On("UpsertPlans", mock.Anything, mock.Anything).Return(nil)
				ps.On("GetByID", mock.Anything, "basic-plan").Return(plan.Plan{}, errors.New("get failed"))
			},
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should successfully create plan with basic data",
			request: connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
				Body: &frontierv1beta1.PlanRequestBody{
					Name:           "basic-plan",
					Title:          "Basic Plan",
					Description:    "A basic plan",
					Interval:       "monthly",
					OnStartCredits: 1000,
					TrialDays:      30,
					Products:       nil,
					Metadata:       testMetadata,
				},
			}),
			setup: func(ps *mocks.PlanService) {
				ps.On("UpsertPlans", mock.Anything, mock.Anything).Return(nil)
				ps.On("GetByID", mock.Anything, "basic-plan").Return(plan.Plan{
					ID:             "plan-123",
					Name:           "basic-plan",
					Title:          "Basic Plan",
					Description:    "A basic plan",
					Interval:       "monthly",
					OnStartCredits: 1000,
					TrialDays:      30,
					Products:       []product.Product{},
					Metadata: metadata.Metadata{
						"key1": "value1",
						"key2": "value2",
					},
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.CreatePlanResponse{
				Plan: &frontierv1beta1.Plan{
					Id:             "plan-123",
					Name:           "basic-plan",
					Title:          "Basic Plan",
					Description:    "A basic plan",
					Interval:       "monthly",
					OnStartCredits: 1000,
					TrialDays:      30,
					Products:       nil,
					Metadata:       testMetadata,
					CreatedAt:      timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:      timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
			}),
		},
		{
			name: "should successfully create plan with products, prices and features",
			request: connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
				Body: &frontierv1beta1.PlanRequestBody{
					Name:        "premium-plan",
					Title:       "Premium Plan",
					Description: "A premium plan",
					Interval:    "yearly",
					Products: []*frontierv1beta1.Product{
						{
							Id:          "product-1",
							Name:        "Premium Product",
							Title:       "Premium Product Title",
							Description: "Premium product description",
							Prices: []*frontierv1beta1.Price{
								{
									Name:             "monthly-price",
									Amount:           2999,
									Currency:         "USD",
									UsageType:        "licensed",
									BillingScheme:    "flat",
									MeteredAggregate: "sum",
									Interval:         "month",
									Metadata:         testMetadata,
								},
							},
							Features: []*frontierv1beta1.Feature{
								{
									Name:       "feature-1",
									ProductIds: []string{"product-1"},
									Metadata:   testMetadata,
								},
							},
							BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
								CreditAmount: 5000,
								SeatLimit:    10,
							},
							Behavior: "credits",
							Metadata: testMetadata,
						},
					},
					Metadata: testMetadata,
				},
			}),
			setup: func(ps *mocks.PlanService) {
				ps.On("UpsertPlans", mock.Anything, mock.Anything).Return(nil)
				ps.On("GetByID", mock.Anything, "premium-plan").Return(plan.Plan{
					ID:          "plan-456",
					Name:        "premium-plan",
					Title:       "Premium Plan",
					Description: "A premium plan",
					Interval:    "yearly",
					Products: []product.Product{
						{
							ID:          "product-1",
							Name:        "Premium Product",
							Title:       "Premium Product Title",
							Description: "Premium product description",
							Prices: []product.Price{
								{
									ID:               "price-1",
									ProductID:        "product-1",
									ProviderID:       "stripe-price-1",
									Name:             "monthly-price",
									Amount:           2999,
									Currency:         "USD",
									UsageType:        product.PriceUsageTypeLicensed,
									BillingScheme:    product.BillingSchemeFlat,
									MeteredAggregate: "sum",
									Interval:         "month",
									Metadata: metadata.Metadata{
										"key1": "value1",
										"key2": "value2",
									},
									CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
								},
							},
							Features: []product.Feature{
								{
									ID:         "feature-1",
									Name:       "feature-1",
									Title:      "Feature 1",
									ProductIDs: []string{"product-1"},
									Metadata: metadata.Metadata{
										"key1": "value1",
										"key2": "value2",
									},
									CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
								},
							},
							Config: product.BehaviorConfig{
								CreditAmount: 5000,
								SeatLimit:    10,
							},
							Behavior: product.CreditBehavior,
							Metadata: metadata.Metadata{
								"key1": "value1",
								"key2": "value2",
							},
							CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
						},
					},
					Metadata: metadata.Metadata{
						"key1": "value1",
						"key2": "value2",
					},
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.CreatePlanResponse{
				Plan: &frontierv1beta1.Plan{
					Id:          "plan-456",
					Name:        "premium-plan",
					Title:       "Premium Plan",
					Description: "A premium plan",
					Interval:    "yearly",
					Products: []*frontierv1beta1.Product{
						{
							Id:          "product-1",
							Name:        "Premium Product",
							Title:       "Premium Product Title",
							Description: "Premium product description",
							Prices: []*frontierv1beta1.Price{
								{
									Id:               "price-1",
									ProductId:        "product-1",
									ProviderId:       "stripe-price-1",
									Name:             "monthly-price",
									Amount:           2999,
									Currency:         "USD",
									UsageType:        "licensed",
									BillingScheme:    "flat",
									MeteredAggregate: "sum",
									Interval:         "month",
									Metadata:         testMetadata,
									CreatedAt:        timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
									UpdatedAt:        timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
								},
							},
							Features: []*frontierv1beta1.Feature{
								{
									Id:         "feature-1",
									Name:       "feature-1",
									Title:      "Feature 1",
									ProductIds: []string{"product-1"},
									Metadata:   testMetadata,
									CreatedAt:  timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
									UpdatedAt:  timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
								},
							},
							BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
								CreditAmount: 5000,
								SeatLimit:    10,
							},
							Behavior:  "credits",
							Metadata:  testMetadata,
							CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
							UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
						},
					},
					Metadata:  testMetadata,
					CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
			}),
		},
		{
			name: "should handle empty metadata gracefully",
			request: connect.NewRequest(&frontierv1beta1.CreatePlanRequest{
				Body: &frontierv1beta1.PlanRequestBody{
					Name:        "no-metadata-plan",
					Title:       "No Metadata Plan",
					Description: "Plan without metadata",
					Interval:    "monthly",
					Products:    []*frontierv1beta1.Product{},
				},
			}),
			setup: func(ps *mocks.PlanService) {
				ps.On("UpsertPlans", mock.Anything, mock.Anything).Return(nil)
				ps.On("GetByID", mock.Anything, "no-metadata-plan").Return(plan.Plan{
					ID:          "plan-789",
					Name:        "no-metadata-plan",
					Title:       "No Metadata Plan",
					Description: "Plan without metadata",
					Interval:    "monthly",
					Products:    []product.Product{},
					Metadata:    metadata.Metadata{},
					CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.CreatePlanResponse{
				Plan: &frontierv1beta1.Plan{
					Id:          "plan-789",
					Name:        "no-metadata-plan",
					Title:       "No Metadata Plan",
					Description: "Plan without metadata",
					Interval:    "monthly",
					Products:    nil,
					CreatedAt:   timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:   timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := mocks.NewPlanService(t)
			if tt.setup != nil {
				tt.setup(mockPlanService)
			}

			handler := &ConnectHandler{
				planService: mockPlanService,
			}

			got, err := handler.CreatePlan(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
				connectErr := err.(*connect.Error)
				assert.Equal(t, tt.errCode, connectErr.Code())
				assert.Contains(t, connectErr.Message(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}

func TestConnectHandler_ListPlans(t *testing.T) {
	testMetadata := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"key1": {Kind: &structpb.Value_StringValue{StringValue: "value1"}},
		},
	}

	tests := []struct {
		name    string
		setup   func(ps *mocks.PlanService)
		request *connect.Request[frontierv1beta1.ListPlansRequest]
		want    *connect.Response[frontierv1beta1.ListPlansResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name:    "should return internal server error when plan service fails",
			request: connect.NewRequest(&frontierv1beta1.ListPlansRequest{}),
			setup: func(ps *mocks.PlanService) {
				ps.On("List", mock.Anything, plan.Filter{}).Return([]plan.Plan{}, errors.New("service failed"))
			},
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name:    "should successfully list plans with empty result",
			request: connect.NewRequest(&frontierv1beta1.ListPlansRequest{}),
			setup: func(ps *mocks.PlanService) {
				ps.On("List", mock.Anything, plan.Filter{}).Return([]plan.Plan{}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListPlansResponse{
				Plans: nil,
			}),
		},
		{
			name:    "should successfully list multiple plans",
			request: connect.NewRequest(&frontierv1beta1.ListPlansRequest{}),
			setup: func(ps *mocks.PlanService) {
				ps.On("List", mock.Anything, plan.Filter{}).Return([]plan.Plan{
					{
						ID:             "plan-1",
						Name:           "basic-plan",
						Title:          "Basic Plan",
						Description:    "A basic plan",
						Interval:       "monthly",
						OnStartCredits: 1000,
						Products:       []product.Product{},
						Metadata: metadata.Metadata{
							"key1": "value1",
						},
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "plan-2",
						Name:        "premium-plan",
						Title:       "Premium Plan",
						Description: "A premium plan",
						Interval:    "yearly",
						Products:    []product.Product{},
						Metadata:    metadata.Metadata{},
						CreatedAt:   time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListPlansResponse{
				Plans: []*frontierv1beta1.Plan{
					{
						Id:             "plan-1",
						Name:           "basic-plan",
						Title:          "Basic Plan",
						Description:    "A basic plan",
						Interval:       "monthly",
						OnStartCredits: 1000,
						Products:       nil,
						Metadata:       testMetadata,
						CreatedAt:      timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt:      timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
					{
						Id:          "plan-2",
						Name:        "premium-plan",
						Title:       "Premium Plan",
						Description: "A premium plan",
						Interval:    "yearly",
						Products:    nil,
						CreatedAt:   timestamppb.New(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt:   timestamppb.New(time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			}),
		},
		{
			name:    "should return internal error when transformPlanToPB fails",
			request: connect.NewRequest(&frontierv1beta1.ListPlansRequest{}),
			setup: func(ps *mocks.PlanService) {
				ps.On("List", mock.Anything, plan.Filter{}).Return([]plan.Plan{
					{
						ID:       "plan-1",
						Name:     "invalid-plan",
						Metadata: metadata.Metadata{"invalid": make(chan int)}, // This will cause ToStructPB to fail
					},
				}, nil)
			},
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := mocks.NewPlanService(t)
			if tt.setup != nil {
				tt.setup(mockPlanService)
			}

			handler := &ConnectHandler{
				planService: mockPlanService,
			}

			got, err := handler.ListPlans(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
				connectErr := err.(*connect.Error)
				assert.Equal(t, tt.errCode, connectErr.Code())
				assert.Contains(t, connectErr.Message(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}

func TestConnectHandler_GetPlan(t *testing.T) {
	testMetadata := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"key1": {Kind: &structpb.Value_StringValue{StringValue: "value1"}},
		},
	}

	tests := []struct {
		name    string
		setup   func(ps *mocks.PlanService)
		request *connect.Request[frontierv1beta1.GetPlanRequest]
		want    *connect.Response[frontierv1beta1.GetPlanResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name:    "should return internal server error when plan service fails",
			request: connect.NewRequest(&frontierv1beta1.GetPlanRequest{Id: "plan-123"}),
			setup: func(ps *mocks.PlanService) {
				ps.On("GetByID", mock.Anything, "plan-123").Return(plan.Plan{}, errors.New("service failed"))
			},
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name:    "should successfully get plan with basic data",
			request: connect.NewRequest(&frontierv1beta1.GetPlanRequest{Id: "plan-123"}),
			setup: func(ps *mocks.PlanService) {
				ps.On("GetByID", mock.Anything, "plan-123").Return(plan.Plan{
					ID:             "plan-123",
					Name:           "basic-plan",
					Title:          "Basic Plan",
					Description:    "A basic plan",
					Interval:       "monthly",
					OnStartCredits: 1000,
					TrialDays:      30,
					Products:       []product.Product{},
					Metadata: metadata.Metadata{
						"key1": "value1",
					},
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.GetPlanResponse{
				Plan: &frontierv1beta1.Plan{
					Id:             "plan-123",
					Name:           "basic-plan",
					Title:          "Basic Plan",
					Description:    "A basic plan",
					Interval:       "monthly",
					OnStartCredits: 1000,
					TrialDays:      30,
					Products:       nil,
					Metadata:       testMetadata,
					CreatedAt:      timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:      timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
			}),
		},
		{
			name:    "should return internal error when transformPlanToPB fails",
			request: connect.NewRequest(&frontierv1beta1.GetPlanRequest{Id: "invalid-plan"}),
			setup: func(ps *mocks.PlanService) {
				ps.On("GetByID", mock.Anything, "invalid-plan").Return(plan.Plan{
					ID:       "invalid-plan",
					Metadata: metadata.Metadata{"invalid": make(chan int)}, // This will cause ToStructPB to fail
				}, nil)
			},
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanService := mocks.NewPlanService(t)
			if tt.setup != nil {
				tt.setup(mockPlanService)
			}

			handler := &ConnectHandler{
				planService: mockPlanService,
			}

			got, err := handler.GetPlan(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
				connectErr := err.(*connect.Error)
				assert.Equal(t, tt.errCode, connectErr.Code())
				assert.Contains(t, connectErr.Message(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}
