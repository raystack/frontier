package logger

import (
	"context"
	"os"

	"github.com/raystack/frontier/pkg/server/consts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/raystack/salt/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(cfg Config) *log.Zap {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(atomicLevel(cfg.Level))
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.DisableCaller = true
	zapCfg.DisableStacktrace = true
	consoleEncoder := zapcore.NewConsoleEncoder(zapCfg.EncoderConfig)

	opt := log.ZapWithConfig(zapCfg, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapCfg.Level)
	}))

	logger := log.NewZap(opt)
	return logger
}

func Ctx(ctx context.Context) *zap.Logger {
	return grpczap.Extract(ctx)
}

func atomicLevel(level string) zapcore.Level {
	switch level {
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func RequestLogFunc(ctx context.Context, msg string, level zapcore.Level, code codes.Code, err error, duration zapcore.Field) {
	requestID := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(consts.RequestIDHeader); len(values) > 0 && values[0] != "" {
			requestID = values[0]
		}
	}
	// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
	grpczap.Extract(ctx).Check(level, msg).Write(
		zap.Error(err),
		zap.String("grpc.code", code.String()),
		zap.String(consts.RequestIDHeader, requestID),
		duration,
	)
}
