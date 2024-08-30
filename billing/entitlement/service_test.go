package entitlement_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/billing/entitlement"
	"github.com/raystack/frontier/billing/entitlement/mocks"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

func mockService(t *testing.T) (*entitlement.Service, *mocks.SubscriptionService,
	*mocks.ProductService, *mocks.PlanService, *mocks.OrganizationService) {
	t.Helper()
	mockSubscription := mocks.NewSubscriptionService(t)
	mockProduct := mocks.NewProductService(t)
	mockPlan := mocks.NewPlanService(t)
	mockOrg := mocks.NewOrganizationService(t)
	return entitlement.NewEntitlementService(mockSubscription, mockProduct, mockPlan, mockOrg),
		mockSubscription, mockProduct, mockPlan, mockOrg
}

func TestService_CheckPlanEligibility(t *testing.T) {
	ctx := context.Background()
	type args struct {
		customerID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func() *entitlement.Service
	}{
		{
			name: "should not return error if customer is not subscribed to any plan",
			args: args{
				customerID: "1",
			},
			wantErr: false,
			setup: func() *entitlement.Service {
				s, mockSubscription, _, _, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "1",
				}).Return([]subscription.Subscription{}, nil)
				return s
			},
		},
		{
			name: "should not return error if customer has no active subscriptions",
			args: args{
				customerID: "2",
			},
			wantErr: false,
			setup: func() *entitlement.Service {
				s, mockSubscription, _, _, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "2",
				}).Return([]subscription.Subscription{
					{
						State: subscription.StatePastDue.String(),
					},
				}, nil)

				return s
			},
		},
		{
			name: "should return error if failed to list subscriptions",
			args: args{
				customerID: "3",
			},
			wantErr: true,
			setup: func() *entitlement.Service {
				s, mockSubscription, _, _, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "3",
				}).Return(nil, fmt.Errorf("error"))
				return s
			},
		},
		{
			name: "should return error if failed to get plan by id",
			args: args{
				customerID: "4",
			},
			wantErr: true,
			setup: func() *entitlement.Service {
				s, mockSubscription, _, mockPlan, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "4",
				}).Return([]subscription.Subscription{
					{
						PlanID: "1",
						State:  subscription.StateActive.String(),
					},
				}, nil)
				mockPlan.EXPECT().GetByID(ctx, "1").Return(plan.Plan{}, fmt.Errorf("error"))
				return s
			},
		},
		{
			name: "should return error if plan member count limit is breached",
			args: args{
				customerID: "5",
			},
			wantErr: true,
			setup: func() *entitlement.Service {
				s, mockSubscription, _, mockPlan, mockOrg := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "5",
				}).Return([]subscription.Subscription{
					{
						PlanID: "1",
						State:  subscription.StateActive.String(),
					},
				}, nil)
				mockPlan.EXPECT().GetByID(ctx, "1").Return(plan.Plan{
					Products: []product.Product{
						{
							ID:       "1",
							Behavior: product.PerSeatBehavior,
							Config: product.BehaviorConfig{
								SeatLimit: 4,
							},
						},
					},
				}, nil)
				mockOrg.EXPECT().MemberCount(ctx, "5").Return(int64(5), nil)
				return s
			},
		},
		{
			name: "should not return error if plan member count is within limits",
			args: args{
				customerID: "6",
			},
			setup: func() *entitlement.Service {
				s, mockSubscription, _, mockPlan, mockOrg := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "6",
				}).Return([]subscription.Subscription{
					{
						PlanID: "1",
						State:  subscription.StateActive.String(),
					},
				}, nil)
				mockPlan.EXPECT().GetByID(ctx, "1").Return(plan.Plan{
					Products: []product.Product{
						{
							ID:       "1",
							Behavior: product.PerSeatBehavior,
							Config: product.BehaviorConfig{
								SeatLimit: 4,
							},
						},
					},
				}, nil)
				mockOrg.EXPECT().MemberCount(ctx, "6").Return(int64(3), nil)
				return s
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.CheckPlanEligibility(ctx, tt.args.customerID); (err != nil) != tt.wantErr {
				t.Errorf("CheckPlanEligibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Check(t *testing.T) {
	ctx := context.Background()
	type args struct {
		customerID         string
		featureOrProductID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    bool
		setup   func() *entitlement.Service
	}{
		{
			name: "should return false if failed to find feature by id",
			args: args{
				customerID:         "1",
				featureOrProductID: "1",
			},
			wantErr: true,
			want:    false,
			setup: func() *entitlement.Service {
				s, mockSubscription, mockProduct, _, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "1",
				}).Return([]subscription.Subscription{}, nil)
				mockProduct.EXPECT().GetFeatureByID(ctx, "1").Return(product.Feature{}, fmt.Errorf("error"))
				return s
			},
		},
		{
			name: "should return true if we find feature belong to same plan as subscription of the customer",
			args: args{
				customerID:         "2",
				featureOrProductID: "feature2",
			},
			wantErr: false,
			want:    true,
			setup: func() *entitlement.Service {
				s, mockSubscription, mockProduct, _, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "2",
				}).Return([]subscription.Subscription{
					{
						PlanID: "plan1",
						State:  subscription.StateActive.String(),
					},
				}, nil)

				mockProduct.EXPECT().GetFeatureByID(ctx, "feature2").Return(product.Feature{
					ProductIDs: []string{"1"},
				}, nil)
				mockProduct.EXPECT().List(ctx, product.Filter{
					ProductIDs: []string{"1"},
				}).Return([]product.Product{
					{
						ID:      "1",
						PlanIDs: []string{"plan1"},
					},
				}, nil)

				return s
			},
		},
		{
			name: "should return false if we find feature does not belong to same plan as subscription of the customer",
			args: args{
				customerID:         "3",
				featureOrProductID: "feature3",
			},
			wantErr: false,
			want:    false,
			setup: func() *entitlement.Service {
				s, mockSubscription, mockProduct, _, _ := mockService(t)
				mockSubscription.EXPECT().List(ctx, subscription.Filter{
					CustomerID: "3",
				}).Return([]subscription.Subscription{
					{
						PlanID: "plan1",
						State:  subscription.StateActive.String(),
					},
				}, nil)

				mockProduct.EXPECT().GetFeatureByID(ctx, "feature3").Return(product.Feature{
					ProductIDs: []string{"1"},
				}, nil)
				mockProduct.EXPECT().List(ctx, product.Filter{
					ProductIDs: []string{"1"},
				}).Return([]product.Product{
					{
						ID:      "1",
						PlanIDs: []string{"plan2"},
					},
				}, nil)

				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Check(ctx, tt.args.customerID, tt.args.featureOrProductID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Check() got = %v, want %v", got, tt.want)
			}
		})
	}
}
