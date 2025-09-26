package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_ListInvoices(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	emptyStruct, _ := structpb.NewStruct(map[string]interface{}{})

	tests := []struct {
		name    string
		setup   func(is *mocks.InvoiceService)
		request *connect.Request[frontierv1beta1.ListInvoicesRequest]
		want    *connect.Response[frontierv1beta1.ListInvoicesResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when invoice service returns error",
			setup: func(is *mocks.InvoiceService) {
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "customer-id",
					NonZeroOnly: false,
				}).Return(nil, errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "customer-id",
				NonzeroAmountOnly: false,
			}),
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should successfully list invoices with empty result",
			setup: func(is *mocks.InvoiceService) {
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "customer-id",
					NonZeroOnly: false,
				}).Return([]invoice.Invoice{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "customer-id",
				NonzeroAmountOnly: false,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListInvoicesResponse{
				Invoices: nil,
			}),
		},
		{
			name: "should successfully list invoices with basic invoice data",
			setup: func(is *mocks.InvoiceService) {
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "customer-id",
					NonZeroOnly: false,
				}).Return([]invoice.Invoice{
					{
						ID:         "invoice-1",
						CustomerID: "customer-id",
						ProviderID: "provider-1",
						State:      invoice.DraftState,
						Currency:   "USD",
						Amount:     1000,
						HostedURL:  "https://example.com/invoice/1",
						Metadata:   nil,
						CreatedAt:  fixedTime,
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "customer-id",
				NonzeroAmountOnly: false,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListInvoicesResponse{
				Invoices: []*frontierv1beta1.Invoice{
					{
						Id:         "invoice-1",
						CustomerId: "customer-id",
						ProviderId: "provider-1",
						State:      "draft",
						Currency:   "USD",
						Amount:     1000,
						HostedUrl:  "https://example.com/invoice/1",
						Metadata:   emptyStruct,
						CreatedAt:  timestamppb.New(fixedTime),
					},
				},
			}),
		},
		{
			name: "should successfully list invoices with nonzero_amount_only filter",
			setup: func(is *mocks.InvoiceService) {
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "customer-id",
					NonZeroOnly: true,
				}).Return([]invoice.Invoice{
					{
						ID:         "invoice-2",
						CustomerID: "customer-id",
						ProviderID: "provider-1",
						State:      invoice.PaidState,
						Currency:   "USD",
						Amount:     2500,
						HostedURL:  "https://example.com/invoice/2",
						Metadata:   nil,
						DueAt:      fixedTime.Add(30 * 24 * time.Hour),
						CreatedAt:  fixedTime,
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "customer-id",
				NonzeroAmountOnly: true,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListInvoicesResponse{
				Invoices: []*frontierv1beta1.Invoice{
					{
						Id:         "invoice-2",
						CustomerId: "customer-id",
						ProviderId: "provider-1",
						State:      "paid",
						Currency:   "USD",
						Amount:     2500,
						HostedUrl:  "https://example.com/invoice/2",
						Metadata:   emptyStruct,
						DueDate:    timestamppb.New(fixedTime.Add(30 * 24 * time.Hour)),
						CreatedAt:  timestamppb.New(fixedTime),
					},
				},
			}),
		},
		{
			name: "should successfully list multiple invoices with all timestamp fields",
			setup: func(is *mocks.InvoiceService) {
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "customer-id",
					NonZeroOnly: false,
				}).Return([]invoice.Invoice{
					{
						ID:            "invoice-3",
						CustomerID:    "customer-id",
						ProviderID:    "provider-1",
						State:         invoice.OpenState,
						Currency:      "USD",
						Amount:        1500,
						HostedURL:     "https://example.com/invoice/3",
						Metadata:      metadata.Metadata{},
						DueAt:         fixedTime.Add(15 * 24 * time.Hour),
						EffectiveAt:   fixedTime.Add(-24 * time.Hour),
						CreatedAt:     fixedTime,
						PeriodStartAt: fixedTime.Add(-30 * 24 * time.Hour),
						PeriodEndAt:   fixedTime,
					},
					{
						ID:         "invoice-4",
						CustomerID: "customer-id",
						ProviderID: "provider-2",
						State:      invoice.State("void"),
						Currency:   "EUR",
						Amount:     0,
						HostedURL:  "https://example.com/invoice/4",
						Metadata:   nil,
						CreatedAt:  fixedTime.Add(time.Hour),
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "customer-id",
				NonzeroAmountOnly: false,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListInvoicesResponse{
				Invoices: []*frontierv1beta1.Invoice{
					{
						Id:            "invoice-3",
						CustomerId:    "customer-id",
						ProviderId:    "provider-1",
						State:         "open",
						Currency:      "USD",
						Amount:        1500,
						HostedUrl:     "https://example.com/invoice/3",
						Metadata:      emptyStruct,
						DueDate:       timestamppb.New(fixedTime.Add(15 * 24 * time.Hour)),
						EffectiveAt:   timestamppb.New(fixedTime.Add(-24 * time.Hour)),
						CreatedAt:     timestamppb.New(fixedTime),
						PeriodStartAt: timestamppb.New(fixedTime.Add(-30 * 24 * time.Hour)),
						PeriodEndAt:   timestamppb.New(fixedTime),
					},
					{
						Id:         "invoice-4",
						CustomerId: "customer-id",
						ProviderId: "provider-2",
						State:      "void",
						Currency:   "EUR",
						Amount:     0,
						HostedUrl:  "https://example.com/invoice/4",
						Metadata:   emptyStruct,
						CreatedAt:  timestamppb.New(fixedTime.Add(time.Hour)),
					},
				},
			}),
		},
		{
			name: "should handle empty billing_id gracefully",
			setup: func(is *mocks.InvoiceService) {
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "",
					NonZeroOnly: false,
				}).Return([]invoice.Invoice{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "",
				NonzeroAmountOnly: false,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListInvoicesResponse{
				Invoices: nil,
			}),
		},
		{
			name: "should return internal error when transformInvoiceToPB fails due to metadata error",
			setup: func(is *mocks.InvoiceService) {
				// Create invoice with metadata that will fail ToStructPB conversion
				invalidMetadata := metadata.Metadata{"invalid": make(chan int)} // channels can't be converted to protobuf
				is.On("List", mock.Anything, invoice.Filter{
					CustomerID:  "customer-id",
					NonZeroOnly: false,
				}).Return([]invoice.Invoice{
					{
						ID:         "invoice-5",
						CustomerID: "customer-id",
						ProviderID: "provider-1",
						State:      invoice.DraftState,
						Currency:   "USD",
						Amount:     1000,
						HostedURL:  "https://example.com/invoice/5",
						Metadata:   invalidMetadata,
						CreatedAt:  fixedTime,
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListInvoicesRequest{
				BillingId:         "customer-id",
				NonzeroAmountOnly: false,
			}),
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceService := &mocks.InvoiceService{}
			if tt.setup != nil {
				tt.setup(mockInvoiceService)
			}

			handler := &ConnectHandler{
				invoiceService: mockInvoiceService,
			}

			got, err := handler.ListInvoices(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.errCode, connect.CodeOf(err))
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockInvoiceService.AssertExpectations(t)
		})
	}
}
