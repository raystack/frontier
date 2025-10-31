package connectinterceptors

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

// UnaryConnectErrorResponseInterceptor handles error processing after the handler executes
// but before the logger interceptor logs the response. This interceptor normalizes
// context-related errors and wraps them in proper Connect error types.
func UnaryConnectErrorResponseInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				// handler error
				if ctx.Err() != nil {
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						return resp, connect.NewError(connect.CodeDeadlineExceeded, ctx.Err())
					}
					return resp, connect.NewError(connect.CodeCanceled, ctx.Err())
				} else {
					// no context error, only handler error
					return resp, err
				}
			} else {
				// no handler error
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					return resp, connect.NewError(connect.CodeDeadlineExceeded, ctx.Err())
				} else if errors.Is(ctx.Err(), context.Canceled) {
					return resp, connect.NewError(connect.CodeCanceled, ctx.Err())
				} else {
					// no context error, no handler error
					return resp, err
				}
			}
		}
	}
}
