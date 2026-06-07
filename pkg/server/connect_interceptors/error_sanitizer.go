package connectinterceptors

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

func UnaryConnectErrorSanitizerInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err == nil {
				return resp, nil
			}

			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
			}

			code := connectErr.Code()
			if code == connect.CodeInternal || code == connect.CodeUnknown {
				return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
			}

			return resp, err
		}
	}
}
