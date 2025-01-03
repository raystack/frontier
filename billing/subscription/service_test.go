package subscription_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/billing/product"
	stripemock "github.com/raystack/frontier/billing/stripetest/mocks"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/billing/subscription/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
)

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(*mocks.Repository)
		want    subscription.Subscription
		wantErr error
	}{
		{
			name: "should return subscription if found",
			id:   "test-id",
			setup: func(r *mocks.Repository) {
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{
					ID:     "test-id",
					PlanID: "plan-1",
					State:  "active",
				}, nil)
			},
			want: subscription.Subscription{
				ID:     "test-id",
				PlanID: "plan-1",
				State:  "active",
			},
		},
		{
			name: "should return error if not found",
			id:   "test-id",
			setup: func(r *mocks.Repository) {
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{}, subscription.ErrNotFound)
			},
			wantErr: subscription.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}

			svc := subscription.NewService(nil, billing.Config{}, mockRepo, nil, nil, nil, nil, nil)
			got, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Cancel(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		immediate bool
		setup     func(*mocks.Repository, *stripe.Subscription, *stripe.SubscriptionSchedule)
		wantErr   error
	}{
		{
			name:      "should return error if subscription not found",
			id:        "test-id",
			immediate: true,
			setup: func(r *mocks.Repository, _ *stripe.Subscription, _ *stripe.SubscriptionSchedule) {
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{}, subscription.ErrNotFound)
			},
			wantErr: subscription.ErrNotFound,
		},
		{
			name:      "should cancel subscription immediately",
			id:        "test-id",
			immediate: true,
			setup: func(r *mocks.Repository, stripeSub *stripe.Subscription, _ *stripe.SubscriptionSchedule) {
				// Setup repository expectations
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{
					ID:         "test-id",
					State:      subscription.StateActive.String(),
					ProviderID: "stripe-sub-id",
				}, nil)

				// Setup stripe subscription response
				*stripeSub = stripe.Subscription{
					ID:         "stripe-sub-id",
					Status:     stripe.SubscriptionStatusCanceled,
					CanceledAt: time.Now().Unix(),
				}

				r.EXPECT().UpdateByID(mock.Anything, mock.MatchedBy(func(s subscription.Subscription) bool {
					return s.ID == "test-id" && s.State == subscription.StateCanceled.String()
				})).Return(subscription.Subscription{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			mockBackend := stripemock.NewBackend(t)

			// Create stripe client with mock backend
			stripeClient := client.New("key_123", &stripe.Backends{
				API: mockBackend,
			})

			stripeSub := &stripe.Subscription{}
			stripeSched := &stripe.SubscriptionSchedule{}

			if tt.setup != nil {
				tt.setup(mockRepo, stripeSub, stripeSched)
			}

			// Setup mock backend expectations
			if stripeSub.ID != "" {
				mockBackend.EXPECT().Call("GET", "/v1/subscriptions/"+stripeSub.ID, "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sub := v.(*stripe.Subscription)
					sub.ID = stripeSub.ID
					sub.Status = stripe.SubscriptionStatusActive
				}).Return(nil)

				mockBackend.EXPECT().Call("POST", "/v1/subscription_schedules", "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sched := v.(*stripe.SubscriptionSchedule)
					sched.ID = "sched_123"
				}).Return(nil)

				mockBackend.EXPECT().Call("DELETE", "/v1/subscriptions/"+stripeSub.ID, "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sub := v.(*stripe.Subscription)
					sub.ID = stripeSub.ID
					sub.Status = stripe.SubscriptionStatusCanceled
					sub.CanceledAt = time.Now().Unix()
				}).Return(nil)
			}

			svc := subscription.NewService(stripeClient, billing.Config{}, mockRepo, nil, nil, nil, nil, nil)
			_, err := svc.Cancel(context.Background(), tt.id, tt.immediate)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestService_ChangePlan(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		change  subscription.ChangeRequest
		setup   func(*mocks.Repository, *mocks.PlanService, *mocks.CustomerService, *mocks.OrganizationService)
		want    subscription.Phase
		wantErr error
	}{
		{
			name: "should return error if subscription not found",
			id:   "test-id",
			change: subscription.ChangeRequest{
				PlanID:    "new-plan",
				Immediate: true,
			},
			setup: func(r *mocks.Repository, p *mocks.PlanService, c *mocks.CustomerService, o *mocks.OrganizationService) {
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{}, subscription.ErrNotFound)
			},
			wantErr: subscription.ErrNotFound,
		},
		{
			name: "should return error if subscription not active",
			id:   "test-id",
			change: subscription.ChangeRequest{
				PlanID:    "new-plan",
				Immediate: true,
			},
			setup: func(r *mocks.Repository, p *mocks.PlanService, c *mocks.CustomerService, o *mocks.OrganizationService) {
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{
					ID:    "test-id",
					State: subscription.StateCanceled.String(),
				}, nil)
			},
			wantErr: errors.New("only active subscriptions can be changed"),
		},
		{
			name: "should return error if plan not found",
			id:   "test-id",
			change: subscription.ChangeRequest{
				PlanID:    "new-plan",
				Immediate: true,
			},
			setup: func(r *mocks.Repository, p *mocks.PlanService, c *mocks.CustomerService, o *mocks.OrganizationService) {
				r.EXPECT().GetByID(mock.Anything, "test-id").Return(subscription.Subscription{
					ID:    "test-id",
					State: subscription.StateActive.String(),
				}, nil)
				p.EXPECT().GetByID(mock.Anything, "new-plan").Return(plan.Plan{}, errors.New("plan not found"))
			},
			wantErr: errors.New("plan not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			mockPlanSvc := mocks.NewPlanService(t)
			mockCustomerSvc := mocks.NewCustomerService(t)
			mockOrgSvc := mocks.NewOrganizationService(t)

			if tt.setup != nil {
				tt.setup(mockRepo, mockPlanSvc, mockCustomerSvc, mockOrgSvc)
			}

			svc := subscription.NewService(nil, billing.Config{}, mockRepo, mockCustomerSvc, mockPlanSvc, mockOrgSvc, nil, nil)
			got, err := svc.ChangePlan(context.Background(), tt.id, tt.change)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_SyncWithProvider(t *testing.T) {
	tests := []struct {
		name    string
		cust    customer.Customer
		setup   func(*mocks.Repository, *mocks.CustomerService, *mocks.PlanService, *mocks.OrganizationService, *mocks.ProductService, *stripe.Subscription)
		wantErr error
	}{
		{
			name: "should sync active subscriptions",
			cust: customer.Customer{
				ID:    "customer-1",
				OrgID: "org-1",
				State: customer.ActiveState,
			},
			setup: func(r *mocks.Repository, c *mocks.CustomerService, p *mocks.PlanService, o *mocks.OrganizationService, prodSvc *mocks.ProductService, stripeSub *stripe.Subscription) {
				// Setup stripe subscription first since it affects the flow
				*stripeSub = stripe.Subscription{
					ID:     "stripe-sub-1",
					Status: stripe.SubscriptionStatusActive,
					Items: &stripe.SubscriptionItemList{
						Data: []*stripe.SubscriptionItem{
							{
								ID: "si_123",
								Price: &stripe.Price{
									ID: "price_123",
									Recurring: &stripe.PriceRecurring{
										Interval: stripe.PriceRecurringIntervalMonth,
									},
									Product: &stripe.Product{
										ID: "prod_123",
									},
								},
								Quantity: 1,
								Metadata: map[string]string{
									"price_id":   "price-1",
									"managed_by": "frontier",
								},
							},
						},
					},
					Schedule: &stripe.SubscriptionSchedule{
						ID: "sched_123",
						CurrentPhase: &stripe.SubscriptionScheduleCurrentPhase{
							StartDate: time.Now().Unix(),
							EndDate:   time.Now().Add(24 * time.Hour).Unix(),
						},
						Phases: []*stripe.SubscriptionSchedulePhase{
							{
								StartDate: time.Now().Unix(),
								EndDate:   time.Now().Add(24 * time.Hour).Unix(),
								Items: []*stripe.SubscriptionSchedulePhaseItem{
									{
										Price: &stripe.Price{
											ID: "price_123",
											Recurring: &stripe.PriceRecurring{
												Interval: stripe.PriceRecurringIntervalMonth,
											},
											Product: &stripe.Product{
												ID: "prod_123",
											},
										},
										Quantity: 1,
									},
								},
								Metadata: map[string]string{
									"plan_id":    "plan-1",
									"managed_by": "frontier",
								},
							},
							{
								StartDate: time.Now().Add(24 * time.Hour).Unix(),
								EndDate:   time.Now().Add(48 * time.Hour).Unix(),
								Items: []*stripe.SubscriptionSchedulePhaseItem{
									{
										Price: &stripe.Price{
											ID: "price_123",
											Recurring: &stripe.PriceRecurring{
												Interval: stripe.PriceRecurringIntervalMonth,
											},
											Product: &stripe.Product{
												ID: "prod_123",
											},
										},
										Quantity: 1,
									},
								},
								Metadata: map[string]string{
									"plan_id":    "plan-1",
									"managed_by": "frontier",
								},
							},
						},
					},
				}

				// 1. List subscriptions
				r.EXPECT().List(mock.Anything, mock.MatchedBy(func(f subscription.Filter) bool {
					return f.CustomerID == "customer-1"
				})).Return([]subscription.Subscription{
					{
						ID:         "sub-1",
						CustomerID: "customer-1",
						PlanID:     "plan-1",
						State:      subscription.StateActive.String(),
						ProviderID: "stripe-sub-1",
					},
				}, nil)

				// 2. Update subscription state from stripe
				r.EXPECT().UpdateByID(mock.Anything, mock.MatchedBy(func(s subscription.Subscription) bool {
					return s.ID == "sub-1" && s.State == subscription.StateActive.String()
				})).Return(subscription.Subscription{
					ID:         "sub-1",
					CustomerID: "customer-1",
					PlanID:     "plan-1",
					State:      subscription.StateActive.String(),
					ProviderID: "stripe-sub-1",
				}, nil)

				// 3. Get plan details for active subscription
				p.EXPECT().GetByID(mock.Anything, "plan-1").Return(plan.Plan{
					ID: "plan-1",
					Products: []product.Product{
						{
							ID:       "product-1",
							Behavior: product.PerSeatBehavior,
							Prices: []product.Price{
								{
									ID:         "price-1",
									ProviderID: "price_123",
									Interval:   string(stripe.PriceRecurringIntervalMonth),
								},
							},
						},
					},
					Metadata: map[string]interface{}{
						"price_id": "price_123",
					},
				}, nil).Times(2) // Called for both current and next phase

				// 4. Get member count for quantity update
				o.EXPECT().MemberCount(mock.Anything, "org-1").Return(int64(2), nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			mockCustomerSvc := mocks.NewCustomerService(t)
			mockPlanSvc := mocks.NewPlanService(t)
			mockOrgSvc := mocks.NewOrganizationService(t)
			mockProdSvc := mocks.NewProductService(t)
			mockBackend := stripemock.NewBackend(t)

			stripeClient := client.New("key_123", &stripe.Backends{
				API: mockBackend,
			})

			stripeSub := &stripe.Subscription{}

			if tt.setup != nil {
				tt.setup(mockRepo, mockCustomerSvc, mockPlanSvc, mockOrgSvc, mockProdSvc, stripeSub)
			}

			if stripeSub.ID != "" {
				// Mock GET subscription call
				mockBackend.EXPECT().Call("GET", "/v1/subscriptions/"+stripeSub.ID, "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sub := v.(*stripe.Subscription)
					*sub = *stripeSub
				}).Return(nil).Once()

				// Mock GET schedule call
				mockBackend.EXPECT().Call("GET", "/v1/subscription_schedules/sched_123", "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sched := v.(*stripe.SubscriptionSchedule)
					*sched = *stripeSub.Schedule
				}).Return(nil).Once()

				// Mock POST subscription update call for quantity
				mockBackend.EXPECT().Call("POST", "/v1/subscriptions/"+stripeSub.ID, "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sub := v.(*stripe.Subscription)
					*sub = *stripeSub
					sub.Items.Data[0].Quantity = 2 // Updated quantity
				}).Return(nil).Once()

				// Mock POST subscription schedule update call for both phases
				mockBackend.EXPECT().Call("POST", "/v1/subscription_schedules/sched_123", "key_123",
					mock.Anything, mock.Anything).Run(func(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) {
					sched := v.(*stripe.SubscriptionSchedule)
					*sched = *stripeSub.Schedule
					sched.Phases[0].Items[0].Quantity = 2 // Updated quantity for current phase
					sched.Phases[1].Items[0].Quantity = 2 // Updated quantity for next phase
				}).Return(nil).Once()
			}

			svc := subscription.NewService(
				stripeClient,
				billing.Config{
					ProductConfig: billing.ProductConfig{
						SeatChangeBehavior: "exact",
					},
				},
				mockRepo,
				mockCustomerSvc,
				mockPlanSvc,
				mockOrgSvc,
				mockProdSvc,
				nil,
			)

			err := svc.SyncWithProvider(context.Background(), tt.cust)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestService_HasUserSubscribedBefore(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		planID     string
		setup      func(*mocks.Repository)
		want       bool
		wantErr    error
	}{
		{
			name:       "should return true if user has active subscription",
			customerID: "customer-1",
			planID:     "plan-1",
			setup: func(r *mocks.Repository) {
				r.EXPECT().List(mock.Anything, subscription.Filter{
					CustomerID: "customer-1",
				}).Return([]subscription.Subscription{
					{
						ID:         "sub-1",
						CustomerID: "customer-1",
						PlanID:     "plan-1",
						State:      subscription.StateActive.String(),
					},
				}, nil)
			},
			want: true,
		},
		{
			name:       "should return true if user had subscription in history",
			customerID: "customer-1",
			planID:     "plan-1",
			setup: func(r *mocks.Repository) {
				r.EXPECT().List(mock.Anything, subscription.Filter{
					CustomerID: "customer-1",
				}).Return([]subscription.Subscription{
					{
						ID:         "sub-1",
						CustomerID: "customer-1",
						PlanID:     "plan-2",
						State:      subscription.StateActive.String(),
						PlanHistory: []subscription.Phase{
							{
								PlanID: "plan-1",
							},
						},
					},
				}, nil)
			},
			want: true,
		},
		{
			name:       "should return false if user never subscribed",
			customerID: "customer-1",
			planID:     "plan-1",
			setup: func(r *mocks.Repository) {
				r.EXPECT().List(mock.Anything, subscription.Filter{
					CustomerID: "customer-1",
				}).Return([]subscription.Subscription{
					{
						ID:         "sub-1",
						CustomerID: "customer-1",
						PlanID:     "plan-2",
						State:      subscription.StateActive.String(),
					},
				}, nil)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}

			svc := subscription.NewService(nil, billing.Config{}, mockRepo, nil, nil, nil, nil, nil)
			got, err := svc.HasUserSubscribedBefore(context.Background(), tt.customerID, tt.planID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		sub     subscription.Subscription
		setup   func(*mocks.Repository)
		want    subscription.Subscription
		wantErr error
	}{
		{
			name: "should create new subscription",
			sub: subscription.Subscription{
				CustomerID: "customer-1",
				PlanID:     "plan-1",
				State:      subscription.StateActive.String(),
				Metadata: map[string]interface{}{
					"test": "data",
				},
			},
			setup: func(r *mocks.Repository) {
				r.EXPECT().Create(mock.Anything, mock.MatchedBy(func(s subscription.Subscription) bool {
					return s.CustomerID == "customer-1" && s.PlanID == "plan-1"
				})).Return(subscription.Subscription{
					ID:         "sub-1",
					CustomerID: "customer-1",
					PlanID:     "plan-1",
					State:      subscription.StateActive.String(),
					Metadata: map[string]interface{}{
						"test": "data",
					},
				}, nil)
			},
			want: subscription.Subscription{
				ID:         "sub-1",
				CustomerID: "customer-1",
				PlanID:     "plan-1",
				State:      subscription.StateActive.String(),
				Metadata: map[string]interface{}{
					"test": "data",
				},
			},
		},
		{
			name: "should return error if repository fails",
			sub: subscription.Subscription{
				CustomerID: "customer-1",
				PlanID:     "plan-1",
			},
			setup: func(r *mocks.Repository) {
				r.EXPECT().Create(mock.Anything, mock.Anything).Return(subscription.Subscription{}, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}

			svc := subscription.NewService(nil, billing.Config{}, mockRepo, nil, nil, nil, nil, nil)
			got, err := svc.Create(context.Background(), tt.sub)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name    string
		filter  subscription.Filter
		setup   func(*mocks.Repository)
		want    []subscription.Subscription
		wantErr error
	}{
		{
			name: "should list all subscriptions",
			filter: subscription.Filter{
				CustomerID: "customer-1",
			},
			setup: func(r *mocks.Repository) {
				r.EXPECT().List(mock.Anything, subscription.Filter{
					CustomerID: "customer-1",
				}).Return([]subscription.Subscription{
					{
						ID:         "sub-1",
						CustomerID: "customer-1",
						PlanID:     "plan-1",
						State:      subscription.StateActive.String(),
					},
					{
						ID:         "sub-2",
						CustomerID: "customer-1",
						PlanID:     "plan-2",
						State:      subscription.StateCanceled.String(),
					},
				}, nil)
			},
			want: []subscription.Subscription{
				{
					ID:         "sub-1",
					CustomerID: "customer-1",
					PlanID:     "plan-1",
					State:      subscription.StateActive.String(),
				},
				{
					ID:         "sub-2",
					CustomerID: "customer-1",
					PlanID:     "plan-2",
					State:      subscription.StateCanceled.String(),
				},
			},
		},
		{
			name: "should return error if repository fails",
			filter: subscription.Filter{
				CustomerID: "customer-1",
			},
			setup: func(r *mocks.Repository) {
				r.EXPECT().List(mock.Anything, subscription.Filter{
					CustomerID: "customer-1",
				}).Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}

			svc := subscription.NewService(nil, billing.Config{}, mockRepo, nil, nil, nil, nil, nil)
			got, err := svc.List(context.Background(), tt.filter)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_DeleteByCustomer(t *testing.T) {
	tests := []struct {
		name    string
		cust    customer.Customer
		setup   func(*mocks.Repository, *stripemock.Backend, *mocks.PlanService, *mocks.ProductService)
		wantErr error
	}{
		{
			name: "should return error if listing subscriptions fails",
			cust: customer.Customer{
				ID:    "customer-1",
				State: customer.ActiveState,
			},
			setup: func(r *mocks.Repository, b *stripemock.Backend, p *mocks.PlanService, ps *mocks.ProductService) {
				r.EXPECT().List(mock.Anything, subscription.Filter{
					CustomerID: "customer-1",
				}).Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			mockBackend := stripemock.NewBackend(t)
			mockPlanSvc := mocks.NewPlanService(t)
			mockProdSvc := mocks.NewProductService(t)

			// Create stripe client with mock backend
			stripeClient := client.New("key_123", &stripe.Backends{
				API: mockBackend,
			})

			if tt.setup != nil {
				tt.setup(mockRepo, mockBackend, mockPlanSvc, mockProdSvc)
			}

			svc := subscription.NewService(stripeClient, billing.Config{}, mockRepo, nil, mockPlanSvc, nil, mockProdSvc, nil)
			err := svc.DeleteByCustomer(context.Background(), tt.cust)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestService_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  billing.Config
		wantErr error
	}{
		{
			name: "should initialize service with cron job",
			config: billing.Config{
				RefreshInterval: billing.RefreshInterval{
					Subscription: time.Minute,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := subscription.NewService(nil, tt.config, nil, nil, nil, nil, nil, nil)
			err := svc.Init(context.Background())

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, svc.Close())
		})
	}
}
