package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_UpdateBillingAccountDetails(t *testing.T) {
	tests := []struct {
		name              string
		request           *connect.Request[frontierv1beta1.UpdateBillingAccountDetailsRequest]
		mockUpdateDetails customer.Details
		mockUpdateError   error
		expectError       bool
		expectedError     *connect.Error
	}{
		{
			name: "successful billing account details update",
			request: connect.NewRequest(&frontierv1beta1.UpdateBillingAccountDetailsRequest{
				Id:        "billing-account-id",
				CreditMin: -100,
				DueInDays: 30,
			}),
			mockUpdateDetails: customer.Details{
				CreditMin: -100,
				DueInDays: 30,
			},
			mockUpdateError: nil,
			expectError:     false,
		},
		{
			name: "negative due in days error",
			request: connect.NewRequest(&frontierv1beta1.UpdateBillingAccountDetailsRequest{
				Id:        "billing-account-id",
				CreditMin: -100,
				DueInDays: -1, // Negative due_in_days not allowed
			}),
			expectError:   true,
			expectedError: connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot create predated invoices: due in days should be greater than 0")),
		},
		{
			name: "update details error",
			request: connect.NewRequest(&frontierv1beta1.UpdateBillingAccountDetailsRequest{
				Id:        "billing-account-id",
				CreditMin: -100,
				DueInDays: 30,
			}),
			mockUpdateError: errors.New("failed to update details"),
			expectError:     true,
			expectedError:   connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCustomerService := mocks.NewCustomerService(t)

			if tt.request.Msg.GetDueInDays() >= 0 {
				mockCustomerService.EXPECT().UpdateDetails(mock.Anything, tt.request.Msg.GetId(), mock.Anything).
					Return(tt.mockUpdateDetails, tt.mockUpdateError)
			}

			handler := &ConnectHandler{
				customerService: mockCustomerService,
			}

			ctx := context.Background()
			ctx = audit.SetContextWithService(ctx, audit.NewService("test", audit.NewNoopRepository(), audit.NewNoopWebhookService()))

			_, err := handler.UpdateBillingAccountDetails(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
