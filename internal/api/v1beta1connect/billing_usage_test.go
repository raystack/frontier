package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/usage"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_CreateBillingUsage(t *testing.T) {
	tests := []struct {
		name          string
		customerSetup func(cs *mocks.CustomerService)
		setup         func(us *mocks.UsageService)
		request       *connect.Request[frontierv1beta1.CreateBillingUsageRequest]
		want          *connect.Response[frontierv1beta1.CreateBillingUsageResponse]
		wantErr       error
		errCode       connect.Code
	}{
		{
			name: "should return internal server error when usage service returns generic error",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Type:        "credit",
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.CreditType,
						Amount:      100,
						Source:      "api",
						Description: "API usage",
						UserID:      "user-123",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return invalid argument error when insufficient credits",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Type:        "credit",
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.CreditType,
						Amount:      100,
						Source:      "api",
						Description: "API usage",
						UserID:      "user-123",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(credit.ErrInsufficientCredits)
			},
			want:    nil,
			wantErr: credit.ErrInsufficientCredits,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return already exists error when usage already applied",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Type:        "credit",
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.CreditType,
						Amount:      100,
						Source:      "api",
						Description: "API usage",
						UserID:      "user-123",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(credit.ErrAlreadyApplied)
			},
			want:    nil,
			wantErr: credit.ErrAlreadyApplied,
			errCode: connect.CodeAlreadyExists,
		},
		{
			name: "should successfully create billing usage with default credit type",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.CreditType,
						Amount:      100,
						Source:      "api",
						Description: "API usage",
						UserID:      "user-123",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(nil)
			},
			want:    connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully create billing usage with custom type",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Type:        "custom_type",
						Source:      "Dashboard",
						Description: "Dashboard usage",
						UserId:      "user-456",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.Type("custom_type"),
						Amount:      100,
						Source:      "dashboard",
						Description: "Dashboard usage",
						UserID:      "user-456",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(nil)
			},
			want:    connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully create multiple billing usages",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Type:        "credit",
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					{
						Id:          "usage-2",
						Amount:      200,
						Type:        "debit",
						Source:      "Dashboard",
						Description: "Dashboard usage",
						UserId:      "user-456",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.Type("credit"),
						Amount:      100,
						Source:      "api",
						Description: "API usage",
						UserID:      "user-123",
						Metadata:    map[string]interface{}{},
					},
					{
						ID:          "usage-2",
						CustomerID:  "billing-123",
						Type:        usage.Type("debit"),
						Amount:      200,
						Source:      "dashboard",
						Description: "Dashboard usage",
						UserID:      "user-456",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(nil)
			},
			want:    connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should handle empty source by lowercasing",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				OrgId: "org-123",
				Usages: []*frontierv1beta1.Usage{
					{
						Id:          "usage-1",
						Amount:      100,
						Source:      "",
						Description: "Empty source usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
			}),
			customerSetup: func(cs *mocks.CustomerService) {
				cs.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(us *mocks.UsageService) {
				expectedUsages := []usage.Usage{
					{
						ID:          "usage-1",
						CustomerID:  "billing-123",
						Type:        usage.CreditType,
						Amount:      100,
						Source:      "",
						Description: "Empty source usage",
						UserID:      "user-123",
						Metadata:    map[string]interface{}{},
					},
				}
				us.EXPECT().Report(mock.Anything, expectedUsages).Return(nil)
			},
			want:    connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCustomerSvc := new(mocks.CustomerService)
			mockUsageSvc := new(mocks.UsageService)
			if tt.customerSetup != nil {
				tt.customerSetup(mockCustomerSvc)
			}
			if tt.setup != nil {
				tt.setup(mockUsageSvc)
			}
			h := &ConnectHandler{
				customerService: mockCustomerSvc,
				usageService:    mockUsageSvc,
			}
			got, err := h.CreateBillingUsage(context.Background(), tt.request)
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

func TestConnectHandler_ListBillingTransactions(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	testTransactions := []credit.Transaction{
		{
			ID:          "txn-1",
			CustomerID:  "billing-123",
			Amount:      100,
			Type:        credit.CreditType,
			Source:      "API",
			Description: "API usage",
			UserID:      "user-123",
			Metadata:    metadata.Metadata{"key": "value"},
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
		},
		{
			ID:          "txn-2",
			CustomerID:  "billing-123",
			Amount:      -50,
			Type:        credit.DebitType,
			Source:      "Dashboard",
			Description: "Dashboard usage",
			UserID:      "user-456",
			Metadata:    metadata.Metadata{"key2": "value2"},
			CreatedAt:   testTime.Add(time.Hour),
			UpdatedAt:   testTime.Add(time.Hour),
		},
	}

	tests := []struct {
		name          string
		setup         func(cs *mocks.CreditService)
		customerSetup func(custSvc *mocks.CustomerService)
		request       *connect.Request[frontierv1beta1.ListBillingTransactionsRequest]
		want          *connect.Response[frontierv1beta1.ListBillingTransactionsResponse]
		wantErr       error
		errCode       connect.Code
	}{
		{
			name: "should return empty list when billing account not found",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{}, customer.ErrNotFound)
			},
			setup: func(cs *mocks.CreditService) {},
			want: connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: []*frontierv1beta1.BillingTransaction{},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return internal server error when credit service returns error",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, credit.Filter{
					CustomerID: "billing-123",
					StartRange: time.Time{},
					EndRange:   time.Time{},
				}).Return(nil, errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should successfully list transactions with basic request",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, credit.Filter{
					CustomerID: "billing-123",
					StartRange: time.Time{},
					EndRange:   time.Time{},
				}).Return(testTransactions, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: []*frontierv1beta1.BillingTransaction{
					{
						Id:          "txn-1",
						CustomerId:  "billing-123",
						Amount:      100,
						Type:        string(credit.CreditType),
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key": {Kind: &structpb.Value_StringValue{StringValue: "value"}}}},
						CreatedAt:   timestamppb.New(testTime),
						UpdatedAt:   timestamppb.New(testTime),
					},
					{
						Id:          "txn-2",
						CustomerId:  "billing-123",
						Amount:      -50,
						Type:        string(credit.DebitType),
						Source:      "Dashboard",
						Description: "Dashboard usage",
						UserId:      "user-456",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key2": {Kind: &structpb.Value_StringValue{StringValue: "value2"}}}},
						CreatedAt:   timestamppb.New(testTime.Add(time.Hour)),
						UpdatedAt:   timestamppb.New(testTime.Add(time.Hour)),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully list transactions with since parameter",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId: "org-123",
				Since: timestamppb.New(testTime),
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, credit.Filter{
					CustomerID: "billing-123",
					StartRange: testTime,
					EndRange:   time.Time{},
				}).Return(testTransactions, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: []*frontierv1beta1.BillingTransaction{
					{
						Id:          "txn-1",
						CustomerId:  "billing-123",
						Amount:      100,
						Type:        string(credit.CreditType),
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key": {Kind: &structpb.Value_StringValue{StringValue: "value"}}}},
						CreatedAt:   timestamppb.New(testTime),
						UpdatedAt:   timestamppb.New(testTime),
					},
					{
						Id:          "txn-2",
						CustomerId:  "billing-123",
						Amount:      -50,
						Type:        string(credit.DebitType),
						Source:      "Dashboard",
						Description: "Dashboard usage",
						UserId:      "user-456",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key2": {Kind: &structpb.Value_StringValue{StringValue: "value2"}}}},
						CreatedAt:   timestamppb.New(testTime.Add(time.Hour)),
						UpdatedAt:   timestamppb.New(testTime.Add(time.Hour)),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully list transactions with start and end range",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId:      "org-123",
				StartRange: timestamppb.New(testTime),
				EndRange:   timestamppb.New(testTime.Add(2 * time.Hour)),
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, credit.Filter{
					CustomerID: "billing-123",
					StartRange: testTime,
					EndRange:   testTime.Add(2 * time.Hour),
				}).Return(testTransactions, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: []*frontierv1beta1.BillingTransaction{
					{
						Id:          "txn-1",
						CustomerId:  "billing-123",
						Amount:      100,
						Type:        string(credit.CreditType),
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key": {Kind: &structpb.Value_StringValue{StringValue: "value"}}}},
						CreatedAt:   timestamppb.New(testTime),
						UpdatedAt:   timestamppb.New(testTime),
					},
					{
						Id:          "txn-2",
						CustomerId:  "billing-123",
						Amount:      -50,
						Type:        string(credit.DebitType),
						Source:      "Dashboard",
						Description: "Dashboard usage",
						UserId:      "user-456",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key2": {Kind: &structpb.Value_StringValue{StringValue: "value2"}}}},
						CreatedAt:   timestamppb.New(testTime.Add(time.Hour)),
						UpdatedAt:   timestamppb.New(testTime.Add(time.Hour)),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully list transactions with empty result",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, credit.Filter{
					CustomerID: "billing-123",
					StartRange: time.Time{},
					EndRange:   time.Time{},
				}).Return([]credit.Transaction{}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: nil,
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should prioritize start_range over since when both provided",
			request: connect.NewRequest(&frontierv1beta1.ListBillingTransactionsRequest{
				OrgId:      "org-123",
				Since:      timestamppb.New(testTime),
				StartRange: timestamppb.New(testTime.Add(time.Hour)),
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().List(mock.Anything, credit.Filter{
					CustomerID: "billing-123",
					StartRange: testTime.Add(time.Hour), // start_range should take precedence over since
					EndRange:   time.Time{},
				}).Return(testTransactions, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: []*frontierv1beta1.BillingTransaction{
					{
						Id:          "txn-1",
						CustomerId:  "billing-123",
						Amount:      100,
						Type:        string(credit.CreditType),
						Source:      "API",
						Description: "API usage",
						UserId:      "user-123",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key": {Kind: &structpb.Value_StringValue{StringValue: "value"}}}},
						CreatedAt:   timestamppb.New(testTime),
						UpdatedAt:   timestamppb.New(testTime),
					},
					{
						Id:          "txn-2",
						CustomerId:  "billing-123",
						Amount:      -50,
						Type:        string(credit.DebitType),
						Source:      "Dashboard",
						Description: "Dashboard usage",
						UserId:      "user-456",
						Metadata:    &structpb.Struct{Fields: map[string]*structpb.Value{"key2": {Kind: &structpb.Value_StringValue{StringValue: "value2"}}}},
						CreatedAt:   timestamppb.New(testTime.Add(time.Hour)),
						UpdatedAt:   timestamppb.New(testTime.Add(time.Hour)),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCreditSvc := new(mocks.CreditService)
			mockCustomerSvc := new(mocks.CustomerService)
			if tt.setup != nil {
				tt.setup(mockCreditSvc)
			}
			if tt.customerSetup != nil {
				tt.customerSetup(mockCustomerSvc)
			}
			h := &ConnectHandler{
				creditService:   mockCreditSvc,
				customerService: mockCustomerSvc,
			}
			got, err := h.ListBillingTransactions(context.Background(), tt.request)
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

func TestConnectHandler_TotalDebitedTransactions(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(cs *mocks.CreditService)
		customerSetup func(custSvc *mocks.CustomerService)
		request       *connect.Request[frontierv1beta1.TotalDebitedTransactionsRequest]
		want          *connect.Response[frontierv1beta1.TotalDebitedTransactionsResponse]
		wantErr       error
		errCode       connect.Code
	}{
		{
			name: "should return empty response when billing account not found",
			request: connect.NewRequest(&frontierv1beta1.TotalDebitedTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{}, customer.ErrNotFound)
			},
			setup: func(cs *mocks.CreditService) {},
			want: connect.NewResponse(&frontierv1beta1.TotalDebitedTransactionsResponse{
				Debited: &frontierv1beta1.BillingAccount_Balance{
					Amount:   0,
					Currency: "VC",
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return internal server error when credit service returns error",
			request: connect.NewRequest(&frontierv1beta1.TotalDebitedTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().GetTotalDebitedAmount(mock.Anything, "billing-123").Return(int64(0), errors.New("service error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should successfully get total debited amount",
			request: connect.NewRequest(&frontierv1beta1.TotalDebitedTransactionsRequest{
				OrgId: "org-123",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-123").Return(customer.Customer{ID: "billing-123"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().GetTotalDebitedAmount(mock.Anything, "billing-123").Return(int64(500), nil)
			},
			want: connect.NewResponse(&frontierv1beta1.TotalDebitedTransactionsResponse{
				Debited: &frontierv1beta1.BillingAccount_Balance{
					Amount:   500,
					Currency: "VC",
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully get zero debited amount",
			request: connect.NewRequest(&frontierv1beta1.TotalDebitedTransactionsRequest{
				OrgId: "org-456",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-456").Return(customer.Customer{ID: "billing-456"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().GetTotalDebitedAmount(mock.Anything, "billing-456").Return(int64(0), nil)
			},
			want: connect.NewResponse(&frontierv1beta1.TotalDebitedTransactionsResponse{
				Debited: &frontierv1beta1.BillingAccount_Balance{
					Amount:   0,
					Currency: "VC",
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should successfully get large debited amount",
			request: connect.NewRequest(&frontierv1beta1.TotalDebitedTransactionsRequest{
				OrgId: "org-789",
			}),
			customerSetup: func(custSvc *mocks.CustomerService) {
				custSvc.EXPECT().GetByOrgID(mock.Anything, "org-789").Return(customer.Customer{ID: "billing-789"}, nil)
			},
			setup: func(cs *mocks.CreditService) {
				cs.EXPECT().GetTotalDebitedAmount(mock.Anything, "billing-789").Return(int64(999999), nil)
			},
			want: connect.NewResponse(&frontierv1beta1.TotalDebitedTransactionsResponse{
				Debited: &frontierv1beta1.BillingAccount_Balance{
					Amount:   999999,
					Currency: "VC",
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCreditSvc := new(mocks.CreditService)
			mockCustomerSvc := new(mocks.CustomerService)
			if tt.setup != nil {
				tt.setup(mockCreditSvc)
			}
			if tt.customerSetup != nil {
				tt.customerSetup(mockCustomerSvc)
			}
			h := &ConnectHandler{
				creditService:   mockCreditSvc,
				customerService: mockCustomerSvc,
			}
			got, err := h.TotalDebitedTransactions(context.Background(), tt.request)
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
