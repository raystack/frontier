package logger

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(appConfig *config.Shield) *log.Zap {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(atomicLevel(appConfig.Log.Level))
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableCaller = true

	opt := log.ZapWithConfig(cfg)
	logger := log.NewZap(opt)
	return logger
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
