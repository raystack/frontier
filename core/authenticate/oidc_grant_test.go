package authenticate

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestIsInvalidGrantErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "invalid_grant maps to true",
			err:  &oauth2.RetrieveError{ErrorCode: "invalid_grant", Response: &http.Response{StatusCode: http.StatusBadRequest}},
			want: true,
		},
		{
			name: "wrapped invalid_grant maps to true",
			err:  fmt.Errorf("token exchange: %w", &oauth2.RetrieveError{ErrorCode: "invalid_grant"}),
			want: true,
		},
		{
			name: "provider rate limit stays false",
			err:  &oauth2.RetrieveError{ErrorCode: "temporarily_unavailable", Response: &http.Response{StatusCode: http.StatusTooManyRequests}},
			want: false,
		},
		{
			name: "provider server error stays false",
			err:  &oauth2.RetrieveError{Response: &http.Response{StatusCode: http.StatusInternalServerError}},
			want: false,
		},
		{
			name: "misconfigured client stays false",
			err:  &oauth2.RetrieveError{ErrorCode: "invalid_client", Response: &http.Response{StatusCode: http.StatusUnauthorized}},
			want: false,
		},
		{
			name: "non oauth2 error stays false",
			err:  fmt.Errorf("some network error"),
			want: false,
		},
		{
			name: "nil error stays false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isInvalidGrantErr(tt.err))
		})
	}
}
