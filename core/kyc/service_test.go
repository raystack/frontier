package kyc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/core/kyc/mocks"
)

func TestService_GetKyc(t *testing.T) {
	tests := []struct {
		name        string
		orgID       string
		mockReturn  kyc.KYC
		mockError   error
		expectError bool
	}{
		{
			name:  "successful fetch",
			orgID: "org-123",
			mockReturn: kyc.KYC{
				OrgID:  "org-123",
				Status: true,
				Link:   "abcd",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "error fetching",
			orgID:       "org-123",
			mockReturn:  kyc.KYC{},
			mockError:   errors.New("some error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			svc := kyc.NewService(mockRepo)
			ctx := context.Background()

			mockRepo.On("GetByOrgID", ctx, tt.orgID).Return(tt.mockReturn, tt.mockError)

			result, err := svc.GetKyc(ctx, tt.orgID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.mockError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_SetKyc(t *testing.T) {
	tests := []struct {
		name        string
		input       kyc.KYC
		mockReturn  kyc.KYC
		mockError   error
		expectError bool
	}{
		{
			name: "successful upsert",
			input: kyc.KYC{
				OrgID:  "org-123",
				Status: true,
				Link:   "abcd",
			},
			mockReturn: kyc.KYC{
				OrgID:  "org-123",
				Status: true,
				Link:   "abcd",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name: "error upserting",
			input: kyc.KYC{
				OrgID:  "org-123",
				Status: true,
				Link:   "abcd",
			},
			mockReturn:  kyc.KYC{},
			mockError:   errors.New("some error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			svc := kyc.NewService(mockRepo)
			ctx := context.Background()

			mockRepo.On("Upsert", ctx, tt.input).Return(tt.mockReturn, tt.mockError)

			result, err := svc.SetKyc(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.mockError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
