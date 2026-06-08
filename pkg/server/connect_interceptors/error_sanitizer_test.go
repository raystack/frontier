package connectinterceptors

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestUnaryConnectErrorSanitizerInterceptor(t *testing.T) {
	sanitizer := UnaryConnectErrorSanitizerInterceptor()

	tests := []struct {
		name        string
		handlerErr  error
		wantCode    connect.Code
		wantMessage string
		wantSameErr bool
	}{
		{
			name:        "nil error passes through",
			handlerErr:  nil,
			wantSameErr: true,
		},
		{
			name:        "CodeNotFound passes through unchanged",
			handlerErr:  connect.NewError(connect.CodeNotFound, errors.New("user not found")),
			wantCode:    connect.CodeNotFound,
			wantMessage: "user not found",
			wantSameErr: true,
		},
		{
			name:        "CodeInvalidArgument passes through unchanged",
			handlerErr:  connect.NewError(connect.CodeInvalidArgument, errors.New("bad input")),
			wantCode:    connect.CodeInvalidArgument,
			wantMessage: "bad input",
			wantSameErr: true,
		},
		{
			name:        "CodePermissionDenied passes through unchanged",
			handlerErr:  connect.NewError(connect.CodePermissionDenied, errors.New("forbidden")),
			wantCode:    connect.CodePermissionDenied,
			wantMessage: "forbidden",
			wantSameErr: true,
		},
		{
			name:        "CodeUnauthenticated passes through unchanged",
			handlerErr:  connect.NewError(connect.CodeUnauthenticated, errors.New("not logged in")),
			wantCode:    connect.CodeUnauthenticated,
			wantMessage: "not logged in",
			wantSameErr: true,
		},
		{
			name:        "CodeInternal is sanitized",
			handlerErr:  connect.NewError(connect.CodeInternal, errors.New("pq: connection refused")),
			wantCode:    connect.CodeInternal,
			wantMessage: "internal server error",
		},
		{
			name:        "CodeUnknown is sanitized to CodeInternal",
			handlerErr:  connect.NewError(connect.CodeUnknown, errors.New("customer not found")),
			wantCode:    connect.CodeInternal,
			wantMessage: "internal server error",
		},
		{
			name:        "raw non-connect error is sanitized to CodeInternal",
			handlerErr:  errors.New("sql: no rows in result set"),
			wantCode:    connect.CodeInternal,
			wantMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := sanitizer(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
				return nil, tt.handlerErr
			})

			_, err := interceptor(context.Background(), nil)

			if tt.handlerErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}

			if tt.wantSameErr {
				if err != tt.handlerErr {
					t.Fatalf("expected error to pass through unchanged, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T", err)
			}
			if connectErr.Code() != tt.wantCode {
				t.Errorf("code = %v, want %v", connectErr.Code(), tt.wantCode)
			}
			if connectErr.Message() != tt.wantMessage {
				t.Errorf("message = %q, want %q", connectErr.Message(), tt.wantMessage)
			}
			if err == tt.handlerErr {
				t.Error("expected sanitized error to be a new instance, got same pointer")
			}
		})
	}
}
