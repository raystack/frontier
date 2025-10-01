package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testCheckoutID = uuid.New().String()
	testCheckout   = checkout.Checkout{
		ID:          testCheckoutID,
		CheckoutUrl: "https://checkout.stripe.com/session123",
		SuccessUrl:  "https://example.com/success",
		CancelUrl:   "https://example.com/cancel",
		State:       "open",
		CustomerID:  "customer-123",
		PlanID:      "plan-123",
		ProductID:   "product-123",
	}
)

func TestConnectHandler_CreateCheckout(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(cs *mocks.CheckoutService)
		req         *connect.Request[frontierv1beta1.CreateCheckoutRequest]
		want        *connect.Response[frontierv1beta1.CreateCheckoutResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should create payment method setup session successfully",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().CreateSessionForPaymentMethod(mock.Anything, checkout.Checkout{
					CustomerID: "customer-123",
					SuccessUrl: "https://example.com/success",
					CancelUrl:  "https://example.com/cancel",
				}).Return(testCheckout, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SetupBody: &frontierv1beta1.CheckoutSetupBody{
					PaymentMethod: true,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
				CheckoutSession: &frontierv1beta1.CheckoutSession{
					Id:          testCheckoutID,
					CheckoutUrl: "https://checkout.stripe.com/session123",
					SuccessUrl:  "https://example.com/success",
					CancelUrl:   "https://example.com/cancel",
					State:       "open",
					Plan:        "plan-123",
					Product:     "product-123",
					CreatedAt:   &timestamppb.Timestamp{},
					UpdatedAt:   &timestamppb.Timestamp{},
					ExpireAt:    &timestamppb.Timestamp{},
				},
			}),
			wantErr: false,
		},
		{
			name: "should create customer portal setup session successfully",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().CreateSessionForCustomerPortal(mock.Anything, checkout.Checkout{
					CustomerID: "customer-123",
					SuccessUrl: "https://example.com/success",
					CancelUrl:  "https://example.com/cancel",
				}).Return(testCheckout, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SetupBody: &frontierv1beta1.CheckoutSetupBody{
					CustomerPortal: true,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
				CheckoutSession: &frontierv1beta1.CheckoutSession{
					Id:          testCheckoutID,
					CheckoutUrl: "https://checkout.stripe.com/session123",
					SuccessUrl:  "https://example.com/success",
					CancelUrl:   "https://example.com/cancel",
					State:       "open",
					Plan:        "plan-123",
					Product:     "product-123",
					CreatedAt:   &timestamppb.Timestamp{},
					UpdatedAt:   &timestamppb.Timestamp{},
					ExpireAt:    &timestamppb.Timestamp{},
				},
			}),
			wantErr: false,
		},
		{
			name: "should create subscription checkout session successfully",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Create(mock.Anything, checkout.Checkout{
					CustomerID:       "customer-123",
					SuccessUrl:       "https://example.com/success",
					CancelUrl:        "https://example.com/cancel",
					PlanID:           "plan-123",
					ProductID:        "",
					Quantity:         2,
					SkipTrial:        false,
					CancelAfterTrial: true,
				}).Return(testCheckout, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
					Plan:             "plan-123",
					SkipTrial:        false,
					CancelAfterTrial: true,
				},
				ProductBody: &frontierv1beta1.CheckoutProductBody{
					Product:  "",
					Quantity: 2,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
				CheckoutSession: &frontierv1beta1.CheckoutSession{
					Id:          testCheckoutID,
					CheckoutUrl: "https://checkout.stripe.com/session123",
					SuccessUrl:  "https://example.com/success",
					CancelUrl:   "https://example.com/cancel",
					State:       "open",
					Plan:        "plan-123",
					Product:     "product-123",
					CreatedAt:   &timestamppb.Timestamp{},
					UpdatedAt:   &timestamppb.Timestamp{},
					ExpireAt:    &timestamppb.Timestamp{},
				},
			}),
			wantErr: false,
		},
		{
			name: "should create product checkout session successfully",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Create(mock.Anything, checkout.Checkout{
					CustomerID:       "customer-123",
					SuccessUrl:       "https://example.com/success",
					CancelUrl:        "https://example.com/cancel",
					PlanID:           "",
					ProductID:        "product-123",
					Quantity:         3,
					SkipTrial:        false,
					CancelAfterTrial: false,
				}).Return(testCheckout, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				ProductBody: &frontierv1beta1.CheckoutProductBody{
					Product:  "product-123",
					Quantity: 3,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateCheckoutResponse{
				CheckoutSession: &frontierv1beta1.CheckoutSession{
					Id:          testCheckoutID,
					CheckoutUrl: "https://checkout.stripe.com/session123",
					SuccessUrl:  "https://example.com/success",
					CancelUrl:   "https://example.com/cancel",
					State:       "open",
					Plan:        "plan-123",
					Product:     "product-123",
					CreatedAt:   &timestamppb.Timestamp{},
					UpdatedAt:   &timestamppb.Timestamp{},
					ExpireAt:    &timestamppb.Timestamp{},
				},
			}),
			wantErr: false,
		},
		{
			name: "should return invalid argument error when no body provided",
			setup: func(cs *mocks.CheckoutService) {
				// No expectations set since no service call should be made
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadRequest,
		},
		{
			name: "should return per seat limit reached error for subscription checkout",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Create(mock.Anything, mock.Anything).Return(checkout.Checkout{}, product.ErrPerSeatLimitReached)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
					Plan:      "plan-123",
					SkipTrial: false,
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrPerSeatLimitReached,
		},
		{
			name: "should return per seat limit reached error for product checkout",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Create(mock.Anything, mock.Anything).Return(checkout.Checkout{}, product.ErrPerSeatLimitReached)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				ProductBody: &frontierv1beta1.CheckoutProductBody{
					Product:  "product-123",
					Quantity: 100,
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrPerSeatLimitReached,
		},
		{
			name: "should return internal server error when payment method setup fails",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().CreateSessionForPaymentMethod(mock.Anything, mock.Anything).Return(checkout.Checkout{}, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SetupBody: &frontierv1beta1.CheckoutSetupBody{
					PaymentMethod: true,
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return internal server error when customer portal setup fails",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().CreateSessionForCustomerPortal(mock.Anything, mock.Anything).Return(checkout.Checkout{}, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SetupBody: &frontierv1beta1.CheckoutSetupBody{
					CustomerPortal: true,
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return internal server error when subscription checkout fails",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Create(mock.Anything, mock.Anything).Return(checkout.Checkout{}, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
					Plan:      "plan-123",
					SkipTrial: false,
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return internal server error when product checkout fails",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Create(mock.Anything, mock.Anything).Return(checkout.Checkout{}, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.CreateCheckoutRequest{
				BillingId:  "customer-123",
				SuccessUrl: "https://example.com/success",
				CancelUrl:  "https://example.com/cancel",
				ProductBody: &frontierv1beta1.CheckoutProductBody{
					Product:  "product-123",
					Quantity: 3,
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCheckoutSvc := mocks.NewCheckoutService(t)
			tt.setup(mockCheckoutSvc)

			h := &ConnectHandler{
				checkoutService: mockCheckoutSvc,
			}

			got, err := h.CreateCheckout(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					assert.Equal(t, tt.wantErrCode, connect.CodeOf(err))
				}
				if tt.wantErrMsg != nil {
					assert.Contains(t, err.Error(), tt.wantErrMsg.Error())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetId(), got.Msg.GetCheckoutSession().GetId())
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetCheckoutUrl(), got.Msg.GetCheckoutSession().GetCheckoutUrl())
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetSuccessUrl(), got.Msg.GetCheckoutSession().GetSuccessUrl())
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetCancelUrl(), got.Msg.GetCheckoutSession().GetCancelUrl())
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetState(), got.Msg.GetCheckoutSession().GetState())
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetPlan(), got.Msg.GetCheckoutSession().GetPlan())
				assert.Equal(t, tt.want.Msg.GetCheckoutSession().GetProduct(), got.Msg.GetCheckoutSession().GetProduct())
			}
		})
	}
}

func TestConnectHandler_DelegatedCheckout(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(cs *mocks.CheckoutService)
		req         *connect.Request[frontierv1beta1.DelegatedCheckoutRequest]
		want        *connect.Response[frontierv1beta1.DelegatedCheckoutResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should delegate subscription checkout successfully",
			setup: func(cs *mocks.CheckoutService) {
				testSubs := &subscription.Subscription{
					ID:         "sub-123",
					CustomerID: "customer-123",
					State:      "active",
				}
				testProd := &product.Product{
					ID:   "product-123",
					Name: "test-product",
				}
				cs.EXPECT().Apply(mock.Anything, checkout.Checkout{
					CustomerID:       "customer-123",
					PlanID:           "plan-123",
					ProductID:        "",
					Quantity:         0,
					SkipTrial:        false,
					CancelAfterTrial: true,
					ProviderCouponID: "coupon-123",
				}).Return(testSubs, testProd, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.DelegatedCheckoutRequest{
				BillingId: "customer-123",
				SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
					Plan:             "plan-123",
					SkipTrial:        false,
					CancelAfterTrial: true,
					ProviderCouponId: "coupon-123",
				},
			}),
			wantErr: false,
		},
		{
			name: "should delegate product checkout successfully",
			setup: func(cs *mocks.CheckoutService) {
				testProd := &product.Product{
					ID:   "product-123",
					Name: "test-product",
				}
				cs.EXPECT().Apply(mock.Anything, checkout.Checkout{
					CustomerID:       "customer-123",
					PlanID:           "",
					ProductID:        "product-123",
					Quantity:         5,
					SkipTrial:        false,
					CancelAfterTrial: false,
					ProviderCouponID: "",
				}).Return(nil, testProd, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.DelegatedCheckoutRequest{
				BillingId: "customer-123",
				ProductBody: &frontierv1beta1.CheckoutProductBody{
					Product:  "product-123",
					Quantity: 5,
				},
			}),
			wantErr: false,
		},
		{
			name: "should return internal server error when apply fails",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().Apply(mock.Anything, mock.Anything).Return(nil, nil, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.DelegatedCheckoutRequest{
				BillingId: "customer-123",
				SubscriptionBody: &frontierv1beta1.CheckoutSubscriptionBody{
					Plan: "plan-123",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCheckoutSvc := mocks.NewCheckoutService(t)
			tt.setup(mockCheckoutSvc)

			h := &ConnectHandler{
				checkoutService: mockCheckoutSvc,
			}

			got, err := h.DelegatedCheckout(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					assert.Equal(t, tt.wantErrCode, connect.CodeOf(err))
				}
				if tt.wantErrMsg != nil {
					assert.Contains(t, err.Error(), tt.wantErrMsg.Error())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestConnectHandler_ListCheckouts(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(cs *mocks.CheckoutService)
		req         *connect.Request[frontierv1beta1.ListCheckoutsRequest]
		want        *connect.Response[frontierv1beta1.ListCheckoutsResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if org_id is empty",
			setup: func(cs *mocks.CheckoutService) {
			},
			req: connect.NewRequest(&frontierv1beta1.ListCheckoutsRequest{
				OrgId:     "",
				BillingId: "test-billing-id",
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadRequest,
		},
		{
			name: "should return error if service returns error",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().List(mock.Anything, checkout.Filter{
					CustomerID: "test-billing-id",
				}).Return(nil, errors.New("service error"))
			},
			req: connect.NewRequest(&frontierv1beta1.ListCheckoutsRequest{
				OrgId:     "test-org-id",
				BillingId: "test-billing-id",
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return empty list if no checkouts found",
			setup: func(cs *mocks.CheckoutService) {
				cs.EXPECT().List(mock.Anything, checkout.Filter{
					CustomerID: "test-billing-id",
				}).Return([]checkout.Checkout{}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListCheckoutsRequest{
				OrgId:     "test-org-id",
				BillingId: "test-billing-id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListCheckoutsResponse{
				CheckoutSessions: []*frontierv1beta1.CheckoutSession{},
			}),
			wantErr: false,
		},
		{
			name: "should return list of checkouts successfully",
			setup: func(cs *mocks.CheckoutService) {
				createdAt := time.Now()
				updatedAt := time.Now()
				expireAt := time.Now()
				cs.EXPECT().List(mock.Anything, checkout.Filter{
					CustomerID: "test-billing-id",
				}).Return([]checkout.Checkout{
					{
						ID:          "test-checkout-id-1",
						CheckoutUrl: "https://checkout.stripe.com/session-1",
						SuccessUrl:  "https://example.com/success",
						CancelUrl:   "https://example.com/cancel",
						State:       "active",
						PlanID:      "plan-1",
						ProductID:   "",
						CreatedAt:   createdAt,
						UpdatedAt:   updatedAt,
						ExpireAt:    expireAt,
					},
					{
						ID:          "test-checkout-id-2",
						CheckoutUrl: "https://checkout.stripe.com/session-2",
						SuccessUrl:  "https://example.com/success",
						CancelUrl:   "https://example.com/cancel",
						State:       "expired",
						PlanID:      "",
						ProductID:   "product-1",
						CreatedAt:   createdAt,
						UpdatedAt:   updatedAt,
						ExpireAt:    expireAt,
					},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListCheckoutsRequest{
				OrgId:     "test-org-id",
				BillingId: "test-billing-id",
			}),
			want: func() *connect.Response[frontierv1beta1.ListCheckoutsResponse] {
				createdAt := time.Now()
				updatedAt := time.Now()
				expireAt := time.Now()
				return connect.NewResponse(&frontierv1beta1.ListCheckoutsResponse{
					CheckoutSessions: []*frontierv1beta1.CheckoutSession{
						{
							Id:          "test-checkout-id-1",
							CheckoutUrl: "https://checkout.stripe.com/session-1",
							SuccessUrl:  "https://example.com/success",
							CancelUrl:   "https://example.com/cancel",
							State:       "active",
							Plan:        "plan-1",
							Product:     "",
							CreatedAt:   timestamppb.New(createdAt),
							UpdatedAt:   timestamppb.New(updatedAt),
							ExpireAt:    timestamppb.New(expireAt),
						},
						{
							Id:          "test-checkout-id-2",
							CheckoutUrl: "https://checkout.stripe.com/session-2",
							SuccessUrl:  "https://example.com/success",
							CancelUrl:   "https://example.com/cancel",
							State:       "expired",
							Plan:        "",
							Product:     "product-1",
							CreatedAt:   timestamppb.New(createdAt),
							UpdatedAt:   timestamppb.New(updatedAt),
							ExpireAt:    timestamppb.New(expireAt),
						},
					},
				})
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkoutService := mocks.NewCheckoutService(t)
			if tt.setup != nil {
				tt.setup(checkoutService)
			}
			h := &ConnectHandler{
				checkoutService: checkoutService,
			}
			got, err := h.ListCheckouts(context.Background(), tt.req)
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
				assert.Equal(t, len(tt.want.Msg.GetCheckoutSessions()), len(got.Msg.GetCheckoutSessions()))
				for i, wantSession := range tt.want.Msg.GetCheckoutSessions() {
					gotSession := got.Msg.GetCheckoutSessions()[i]
					assert.Equal(t, wantSession.GetId(), gotSession.GetId())
					assert.Equal(t, wantSession.GetCheckoutUrl(), gotSession.GetCheckoutUrl())
					assert.Equal(t, wantSession.GetSuccessUrl(), gotSession.GetSuccessUrl())
					assert.Equal(t, wantSession.GetCancelUrl(), gotSession.GetCancelUrl())
					assert.Equal(t, wantSession.GetState(), gotSession.GetState())
					assert.Equal(t, wantSession.GetPlan(), gotSession.GetPlan())
					assert.Equal(t, wantSession.GetProduct(), gotSession.GetProduct())
				}
			}
		})
	}
}
