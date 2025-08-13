package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetOrganizationKyc(t *testing.T) {
	tests := []struct {
		mockService   *mocks.KycService
		name          string
		request       *connect.Request[frontierv1beta1.SetOrganizationKycRequest]
		mockResponse  kyc.KYC
		mockError     error
		expectError   bool
		expectedError error
	}{
		{
			mockService: mocks.NewKycService(t),
			name:        "successful KYC update",
			request: connect.NewRequest(&frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "valid-org-id",
				Status: true,
				Link:   "http://kyc-link.com",
			}),
			mockResponse: kyc.KYC{
				OrgID:  "valid-org-id",
				Status: true,
				Link:   "http://kyc-link.com",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "KYC link not set error",
			request: connect.NewRequest(&frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "valid-org-id",
				Status: true,
				Link:   "",
			}),
			mockError:     kyc.ErrKycLinkNotSet,
			expectError:   true,
			expectedError: connect.NewError(connect.CodeInvalidArgument, kyc.ErrKycLinkNotSet),
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "Invalid UUID error",
			request: connect.NewRequest(&frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "invalid-uuid",
				Status: true,
				Link:   "http://kyc-link.com",
			}),
			mockError:     kyc.ErrInvalidUUID,
			expectError:   true,
			expectedError: connect.NewError(connect.CodeInvalidArgument, kyc.ErrInvalidUUID),
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "Organization does not exist error",
			request: connect.NewRequest(&frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "non-existent-org",
				Status: true,
				Link:   "http://kyc-link.com",
			}),
			mockError:     kyc.ErrOrgDoesntExist,
			expectError:   true,
			expectedError: connect.NewError(connect.CodeInvalidArgument, kyc.ErrOrgDoesntExist),
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "Unexpected internal error",
			request: connect.NewRequest(&frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "valid-org-id",
				Status: true,
				Link:   "http://kyc-link.com",
			}),
			mockError:     errors.New("internal error"),
			expectError:   true,
			expectedError: connect.NewError(connect.CodeInternal, errors.New("internal error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock behavior
			if tt.mockError != nil {
				tt.mockService.EXPECT().SetKyc(mock.Anything, mock.Anything).Return(kyc.KYC{}, tt.mockError)
			} else {
				tt.mockService.EXPECT().SetKyc(mock.Anything, mock.Anything).Return(tt.mockResponse, nil)
			}

			// Create handler with mock service
			handler := ConnectHandler{
				orgKycService: tt.mockService,
			}

			// Create context with audit service
			ctx := context.Background()
			ctx = audit.SetContextWithService(ctx, audit.NewService("test", audit.NewNoopRepository(), audit.NewNoopWebhookService()))

			// Call the handler method
			response, err := handler.SetOrganizationKyc(ctx, tt.request)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.mockResponse.OrgID, response.Msg.GetOrganizationKyc().GetOrgId())
				assert.Equal(t, tt.mockResponse.Status, response.Msg.GetOrganizationKyc().GetStatus())
				assert.Equal(t, tt.mockResponse.Link, response.Msg.GetOrganizationKyc().GetLink())
			}
		})
	}
}

func TestGetOrganizationKyc(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		mockService   *mocks.KycService
		name          string
		request       *connect.Request[frontierv1beta1.GetOrganizationKycRequest]
		mockResponse  kyc.KYC
		mockError     error
		expectError   error
		expectNilResp bool
	}{
		{
			mockService: mocks.NewKycService(t),
			name:        "success case",
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationKycRequest{
				OrgId: "valid-org-id",
			}),
			mockResponse: kyc.KYC{
				OrgID:  "valid-org-id",
				Status: true,
				Link:   "https://example.com/kyc",
			},
			mockError:     nil,
			expectError:   nil,
			expectNilResp: false,
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "error case - KYC record not found",
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationKycRequest{
				OrgId: "nonexistent-org",
			}),
			mockResponse:  kyc.KYC{},
			mockError:     kyc.ErrNotExist,
			expectError:   connect.NewError(connect.CodeNotFound, kyc.ErrNotExist),
			expectNilResp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewKycService(t)
			h := ConnectHandler{orgKycService: mockService}

			mockService.On("GetKyc", ctx, tt.request.Msg.GetOrgId()).Return(tt.mockResponse, tt.mockError)

			resp, err := h.GetOrganizationKyc(ctx, tt.request)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tt.expectNilResp {
				assert.Nil(t, resp)
			} else {
				assert.NotNil(t, resp)
				assert.Equal(t, tt.request.Msg.GetOrgId(), resp.Msg.GetOrganizationKyc().GetOrgId())
			}
		})
	}
}

func TestListOrganizationsKyc(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockService   *mocks.KycService
		mockResponse  []kyc.KYC
		mockError     error
		expectError   error
		expectNilResp bool
	}{
		{
			name:        "success case",
			mockService: mocks.NewKycService(t),
			mockResponse: []kyc.KYC{
				{
					OrgID:  "org-1",
					Status: true,
					Link:   "https://example.com/kyc1",
				},
				{
					OrgID:  "org-2",
					Status: false,
					Link:   "https://example.com/kyc2",
				},
			},
			mockError:     nil,
			expectError:   nil,
			expectNilResp: false,
		},
		{
			name:          "error case - no KYC records found",
			mockService:   mocks.NewKycService(t),
			mockResponse:  nil,
			mockError:     kyc.ErrNotExist,
			expectError:   connect.NewError(connect.CodeNotFound, kyc.ErrNotExist),
			expectNilResp: true,
		},
		{
			name:          "error case - internal error",
			mockService:   mocks.NewKycService(t),
			mockResponse:  nil,
			mockError:     errors.New("internal error"),
			expectError:   connect.NewError(connect.CodeInternal, errors.New("internal error")),
			expectNilResp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewKycService(t)
			h := ConnectHandler{orgKycService: mockService}

			mockService.On("ListKycs", ctx).Return(tt.mockResponse, tt.mockError)

			resp, err := h.ListOrganizationsKyc(ctx, connect.NewRequest(&frontierv1beta1.ListOrganizationsKycRequest{}))

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tt.expectNilResp {
				assert.Nil(t, resp)
			} else {
				assert.NotNil(t, resp)
				assert.Equal(t, len(tt.mockResponse), len(resp.Msg.GetOrganizationsKyc()))
				for i, kyc := range tt.mockResponse {
					assert.Equal(t, kyc.OrgID, resp.Msg.GetOrganizationsKyc()[i].GetOrgId())
					assert.Equal(t, kyc.Status, resp.Msg.GetOrganizationsKyc()[i].GetStatus())
					assert.Equal(t, kyc.Link, resp.Msg.GetOrganizationsKyc()[i].GetLink())
				}
			}
		})
	}
}
