package v1beta1

import (
	"context"
	"testing"

	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Authenticate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(auth *mocks.AuthnService, session *mocks.SessionService)
		request *frontierv1beta1.AuthenticateRequest
		want    *frontierv1beta1.AuthenticateResponse
		wantErr error
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthnSrv := new(mocks.AuthnService)
			mockSessionSrv := new(mocks.SessionService)
			if tt.setup != nil {
				tt.setup(mockAuthnSrv, mockSessionSrv)
			}
			mockDep := Handler{authnService: mockAuthnSrv, sessionService: mockSessionSrv}
			resp, err := mockDep.Authenticate(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
