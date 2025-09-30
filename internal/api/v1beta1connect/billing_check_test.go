package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnectHandler_CheckFeatureEntitlement(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(es *mocks.EntitlementService)
		request *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]
		want    *connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when entitlement service returns error",
			request: connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
				BillingId: "billing-123",
				Feature:   "feature-abc",
			}),
			setup: func(es *mocks.EntitlementService) {
				es.EXPECT().Check(mock.Anything, "billing-123", "feature-abc").Return(false, errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return false when feature is not entitled",
			setup: func(es *mocks.EntitlementService) {
				es.EXPECT().Check(mock.Anything, "billing-123", "feature-abc").Return(false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
				BillingId: "billing-123",
				Feature:   "feature-abc",
			}),
			want: connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
				Status: false,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return true when feature is entitled",
			setup: func(es *mocks.EntitlementService) {
				es.EXPECT().Check(mock.Anything, "billing-123", "feature-abc").Return(true, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
				BillingId: "billing-123",
				Feature:   "feature-abc",
			}),
			want: connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
				Status: true,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should handle empty billing id",
			setup: func(es *mocks.EntitlementService) {
				es.EXPECT().Check(mock.Anything, "", "feature-abc").Return(false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
				BillingId: "",
				Feature:   "feature-abc",
			}),
			want: connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
				Status: false,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should handle empty feature",
			setup: func(es *mocks.EntitlementService) {
				es.EXPECT().Check(mock.Anything, "billing-123", "").Return(false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckFeatureEntitlementRequest{
				BillingId: "billing-123",
				Feature:   "",
			}),
			want: connect.NewResponse(&frontierv1beta1.CheckFeatureEntitlementResponse{
				Status: false,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEntitlementSvc := new(mocks.EntitlementService)
			if tt.setup != nil {
				tt.setup(mockEntitlementSvc)
			}
			h := &ConnectHandler{
				entitlementService: mockEntitlementSvc,
			}
			got, err := h.CheckFeatureEntitlement(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnectHandler_CheckCreditEntitlement(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(cs *mocks.CustomerService, crs *mocks.CreditService)
		request *connect.Request[frontierv1beta1.CheckCreditEntitlementRequest]
		want    *connect.Response[frontierv1beta1.CheckCreditEntitlementResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when customer service list returns error",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return(nil, errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return not found error when no customers exist for organization",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return([]customer.Customer{}, nil)
			},
			want:    nil,
			wantErr: ErrNotFound,
			errCode: connect.CodeNotFound,
		},
		{
			name: "should return internal server error when customer details service returns error",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return([]customer.Customer{{ID: "customer-123", OrgID: "org-123"}}, nil)
				cs.EXPECT().GetDetails(mock.Anything, "customer-123").Return(customer.Details{}, errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return internal server error when credit service returns error",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return([]customer.Customer{{ID: "customer-123", OrgID: "org-123"}}, nil)
				cs.EXPECT().GetDetails(mock.Anything, "customer-123").Return(customer.Details{
					CreditMin: 50,
				}, nil)
				crs.EXPECT().GetBalance(mock.Anything, "customer-123").Return(int64(0), errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return true when sufficient credits available",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return([]customer.Customer{{ID: "customer-123", OrgID: "org-123"}}, nil)
				cs.EXPECT().GetDetails(mock.Anything, "customer-123").Return(customer.Details{
					CreditMin: 50,
				}, nil)
				crs.EXPECT().GetBalance(mock.Anything, "customer-123").Return(int64(200), nil)
			},
			want: connect.NewResponse(&frontierv1beta1.CheckCreditEntitlementResponse{
				Status: true,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return false when insufficient credits available",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return([]customer.Customer{{ID: "customer-123", OrgID: "org-123"}}, nil)
				cs.EXPECT().GetDetails(mock.Anything, "customer-123").Return(customer.Details{
					CreditMin: 50,
				}, nil)
				crs.EXPECT().GetBalance(mock.Anything, "customer-123").Return(int64(120), nil)
			},
			want: connect.NewResponse(&frontierv1beta1.CheckCreditEntitlementResponse{
				Status: false,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return true when exactly at credit minimum after deduction",
			request: connect.NewRequest(&frontierv1beta1.CheckCreditEntitlementRequest{
				OrgId:  "org-123",
				Amount: 100,
			}),
			setup: func(cs *mocks.CustomerService, crs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: "org-123",
				}).Return([]customer.Customer{{ID: "customer-123", OrgID: "org-123"}}, nil)
				cs.EXPECT().GetDetails(mock.Anything, "customer-123").Return(customer.Details{
					CreditMin: 50,
				}, nil)
				crs.EXPECT().GetBalance(mock.Anything, "customer-123").Return(int64(150), nil)
			},
			want: connect.NewResponse(&frontierv1beta1.CheckCreditEntitlementResponse{
				Status: true,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.CustomerService)
			mockCreditSvc := new(mocks.CreditService)
			if tt.setup != nil {
				tt.setup(mockCustomerSvc, mockCreditSvc)
			}
			h := &ConnectHandler{
				customerService: mockCustomerSvc,
				creditService:   mockCreditSvc,
			}
			got, err := h.CheckCreditEntitlement(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
