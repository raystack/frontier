package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testCustomerID = uuid.New().String()
	testCustomers  = []customer.Customer{
		{
			ID:   uuid.New().String(),
			Name: "test-customer",
		},
		{
			ID:   uuid.New().String(),
			Name: "test-customer-2",
		},
	}
)

func TestHandler_GetRequestCustomerID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, cs *mocks.CustomerService)
		req     any
		want    string
		wantErr error
	}{
		{
			name:  "should return billing id from request as customer id",
			setup: func(os *mocks.OrganizationService, cs *mocks.CustomerService) {},
			req: &frontierv1beta1.ListInvoicesRequest{
				OrgId:     testOrgID,
				BillingId: testCustomerID,
			},
			want:    testCustomerID,
			wantErr: nil,
		},
		{
			name:  "should return billing id from request as id",
			setup: func(os *mocks.OrganizationService, cs *mocks.CustomerService) {},
			req: &frontierv1beta1.GetBillingBalanceRequest{
				OrgId: testOrgID,
				Id:    testCustomerID,
			},
			want:    testCustomerID,
			wantErr: nil,
		},
		{
			name: "should return billing id by listing customers via org id",
			setup: func(os *mocks.OrganizationService, cs *mocks.CustomerService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{
					ID: testOrgID,
				}, nil)
				cs.EXPECT().List(mock.Anything, customer.Filter{
					OrgID: testOrgID,
					State: customer.ActiveState,
				}).Return(testCustomers, nil)
			},
			req: &frontierv1beta1.CreateBillingUsageRequest{
				OrgId: testOrgID,
			},
			want:    testCustomers[0].ID,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			mockCustSrv := new(mocks.CustomerService)
			if tt.setup != nil {
				tt.setup(mockOrgSrv, mockCustSrv)
			}
			handler := Handler{
				orgService:      mockOrgSrv,
				customerService: mockCustSrv,
			}
			resp, err := handler.GetRequestCustomerID(context.Background(), tt.req)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestUpdateBillingAccountDetails(t *testing.T) {
	tests := []struct {
		name              string
		request           *frontierv1beta1.UpdateBillingAccountDetailsRequest
		mockUpdateDetails customer.Details
		mockUpdateError   error
		expectError       bool
		expectedError     error
	}{
		{
			name: "successful billing account details update",
			request: &frontierv1beta1.UpdateBillingAccountDetailsRequest{
				Id:        "billing-account-id",
				CreditMin: -100,
				DueInDays: 30,
			},
			mockUpdateDetails: customer.Details{
				CreditMin: -100,
				DueInDays: 30,
			},
			mockUpdateError: nil,
			expectError:     false,
		},
		{
			name: "negative due in days error",
			request: &frontierv1beta1.UpdateBillingAccountDetailsRequest{
				Id:        "billing-account-id",
				CreditMin: -100,
				DueInDays: -1, // Negative due_in_days not allowed
			},
			expectError: true,
		},
		{
			name: "update details error",
			request: &frontierv1beta1.UpdateBillingAccountDetailsRequest{
				Id:        "billing-account-id",
				CreditMin: -100,
				DueInDays: 30,
			},
			mockUpdateError: errors.New("failed to update details"),
			expectError:     true,
			expectedError:   errors.New("failed to update details"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCustomerService := mocks.NewCustomerService(t)

			if tt.request.GetDueInDays() >= 0 {
				mockCustomerService.EXPECT().UpdateDetails(mock.Anything, tt.request.GetId(), mock.Anything).
					Return(tt.mockUpdateDetails, tt.mockUpdateError)
			}

			handler := Handler{
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
