package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/usage"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConnectHandler_CreateBillingUsage(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(us *mocks.UsageService)
		request *connect.Request[frontierv1beta1.CreateBillingUsageRequest]
		want    *connect.Response[frontierv1beta1.CreateBillingUsageResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when usage service returns generic error",
			request: connect.NewRequest(&frontierv1beta1.CreateBillingUsageRequest{
				BillingId: "billing-123",
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
				BillingId: "billing-123",
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
				BillingId: "billing-123",
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
				BillingId: "billing-123",
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
				BillingId: "billing-123",
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
				BillingId: "billing-123",
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
				BillingId: "billing-123",
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
			mockUsageSvc := new(mocks.UsageService)
			if tt.setup != nil {
				tt.setup(mockUsageSvc)
			}
			h := &ConnectHandler{
				usageService: mockUsageSvc,
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
