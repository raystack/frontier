package v1beta1

import (
	"context"
	"errors"
	"testing"

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
		request       *frontierv1beta1.SetOrganizationKycRequest
		mockResponse  kyc.KYC
		mockError     error
		expectError   bool
		expectedError error
	}{
		{
			mockService: mocks.NewKycService(t),
			name:        "successful KYC update",
			request: &frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "valid-org-id",
				Status: true,
				Link:   "http://kyc-link.com",
			},
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
			request: &frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "valid-org-id",
				Status: true,
				Link:   "",
			},
			mockError:     kyc.ErrKycLinkNotSet,
			expectError:   true,
			expectedError: ErrInvalidInput(kyc.ErrKycLinkNotSet.Error()),
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "Invalid UUID error",
			request: &frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "invalid-uuid",
				Status: true,
				Link:   "http://kyc-link.com",
			},
			mockError:     kyc.ErrInvalidUUID,
			expectError:   true,
			expectedError: ErrInvalidInput(kyc.ErrInvalidUUID.Error()),
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "Organization does not exist error",
			request: &frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "non-existent-org",
				Status: true,
				Link:   "http://kyc-link.com",
			},
			mockError:     kyc.ErrOrgDoesntExist,
			expectError:   true,
			expectedError: ErrInvalidInput(kyc.ErrOrgDoesntExist.Error()),
		},
		{
			mockService: mocks.NewKycService(t),
			name:        "Unexpected internal error",
			request: &frontierv1beta1.SetOrganizationKycRequest{
				OrgId:  "valid-org-id",
				Status: true,
				Link:   "http://kyc-link.com",
			},
			mockError:     errors.New("internal error"),
			expectError:   true,
			expectedError: errors.New("internal error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handler{orgKycService: tt.mockService}
			tt.mockService.On("SetKyc", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
			resp, err := h.SetOrganizationKyc(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestGetOrganizationKyc(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		mockService   *mocks.KycService
		name          string
		request       *frontierv1beta1.GetOrganizationKycRequest
		mockResponse  kyc.KYC
		mockError     error
		expectError   error
		expectNilResp bool
	}{
		{
			mockService: mocks.NewKycService(t),
			name:        "success case",
			request: &frontierv1beta1.GetOrganizationKycRequest{
				OrgId: "valid-org-id",
			},
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
			request: &frontierv1beta1.GetOrganizationKycRequest{
				OrgId: "nonexistent-org",
			},
			mockResponse:  kyc.KYC{},
			mockError:     kyc.ErrNotExist,
			expectError:   grpcOrgKycNotFoundErr,
			expectNilResp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewKycService(t)
			h := Handler{orgKycService: mockService}

			mockService.On("GetKyc", ctx, tt.request.GetOrgId()).Return(tt.mockResponse, tt.mockError)

			resp, err := h.GetOrganizationKyc(ctx, tt.request)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.expectNilResp {
				assert.Nil(t, resp)
			} else {
				assert.NotNil(t, resp)
				assert.Equal(t, tt.request.GetOrgId(), resp.GetOrganizationKyc().GetOrgId())
			}
		})
	}
}
