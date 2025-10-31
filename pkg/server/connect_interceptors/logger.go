package connectinterceptors

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/pkg/server/consts"
	"go.uber.org/zap"
)

const loggerContextKey = "logger"

type LoggerOption struct {
	// Decider returns true if method should be logged
	Decider func(procedure string) bool
}

type LoggerOptions struct {
	decider func(procedure string) bool
}

func NewLoggerOptions(opts ...LoggerOption) *LoggerOptions {
	options := &LoggerOptions{
		decider: func(procedure string) bool { return true }, // log everything by default
	}

	for _, opt := range opts {
		if opt.Decider != nil {
			options.decider = opt.Decider
		}
	}
	return options
}

func UnaryConnectLoggerInterceptor(logger *zap.Logger, opts *LoggerOptions) connect.UnaryInterceptorFunc {
	if opts == nil {
		opts = NewLoggerOptions()
	}

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if !opts.decider(req.Spec().Procedure) {
				return next(ctx, req)
			}

			// Embed logger in context
			ctx = context.WithValue(ctx, loggerContextKey, logger)

			startTime := time.Now()
			resp, err := next(ctx, req)
			duration := time.Since(startTime)

			code := connect.Code(0)
			if connectErr, ok := err.(*connect.Error); ok {
				code = connectErr.Code()
			}

			fields := []zap.Field{
				zap.String("system", "connect_rpc"),
				zap.Time("start_time", startTime),
				zap.String("method", req.Spec().Procedure),
				zap.Int64("time_ms", duration.Milliseconds()),
				zap.String("code", code.String()),
				zap.String("request_id", req.Header().Get(consts.RequestIDHeader)),
				zap.Error(err),
			}
			if err == nil && ctx.Err() == nil {
				logger.Info("finished call", fields...)
				return resp, err
			}

			switch code {
			case connect.CodeCanceled:
				logger.Warn("client cancelled request", fields...)
			case connect.CodeDeadlineExceeded:
				logger.Warn("request timeout", fields...)
			case connect.CodeInvalidArgument,
				connect.CodeNotFound,
				connect.CodeAlreadyExists,
				connect.CodeUnauthenticated,
				connect.CodePermissionDenied,
				connect.CodeFailedPrecondition,
				connect.CodeOutOfRange:
				logger.Warn("finished call", fields...)
			default:
				logger.Error("finished call", fields...)
			}
			return resp, err
		}
	}
}
