package logger

import (
	"os"

	"github.com/odpf/salt/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(cfg Config) *log.Zap {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(atomicLevel(cfg.Level))
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.DisableCaller = true
	consoleEncoder := zapcore.NewConsoleEncoder(zapCfg.EncoderConfig)

	opt := log.ZapWithConfig(zapCfg, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapCfg.Level)
	}))

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
