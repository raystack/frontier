package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_errWrapper(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "status errors should pass through",
			err:     status.Error(codes.InvalidArgument, "bad value"),
			wantErr: status.Error(codes.InvalidArgument, "bad value"),
		},
		{
			name:    "context.Canceled should be converted to Canceled",
			err:     context.Canceled,
			wantErr: status.Error(codes.Canceled, context.Canceled.Error()),
		},
		{
			name:    "default to internal server error",
			err:     errors.New("some error"),
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errWrapper(context.Background(), tt.err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
