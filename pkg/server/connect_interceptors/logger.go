package connectinterceptors

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	frontierlogger "github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/pkg/server/consts"
)

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

func UnaryConnectLoggerInterceptor(logger *slog.Logger, opts *LoggerOptions) connect.UnaryInterceptorFunc {
	if opts == nil {
		opts = NewLoggerOptions()
	}

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if !opts.decider(req.Spec().Procedure) {
				return next(ctx, req)
			}

			// Store request-scoped attrs in context for downstream loggers
			requestID := req.Header().Get(consts.RequestIDHeader)
			ctx = frontierlogger.AppendCtx(ctx,
				slog.String("request_id", requestID),
				slog.String("method", req.Spec().Procedure),
			)

			startTime := time.Now()
			resp, err := next(ctx, req)
			duration := time.Since(startTime)

			code := connect.Code(0)
			if connectErr, ok := err.(*connect.Error); ok {
				code = connectErr.Code()
			}

			attrs := []any{
				"system", "connect_rpc",
				"start_time", startTime,
				"method", req.Spec().Procedure,
				"time_ms", duration.Milliseconds(),
				"code", code.String(),
				"request_id", requestID,
			}
			if err != nil {
				attrs = append(attrs, "error", err)
			}

			level := levelForCode(code, err)
			logger.Log(ctx, level, "finished call", attrs...)
			return resp, err
		}
	}
}

func levelForCode(code connect.Code, err error) slog.Level {
	if err == nil {
		return slog.LevelInfo
	}
	switch code {
	case connect.CodeCanceled,
		connect.CodeDeadlineExceeded,
		connect.CodeInvalidArgument,
		connect.CodeNotFound,
		connect.CodeAlreadyExists,
		connect.CodeUnauthenticated,
		connect.CodePermissionDenied,
		connect.CodeFailedPrecondition,
		connect.CodeOutOfRange:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}
