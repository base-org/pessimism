package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogKey = string

type loggerKeyType int

const loggerKey loggerKeyType = iota

// NOTE - Logger is set to Nop as default to avoid redundant testing
var logger *zap.Logger = zap.NewNop()

// Config ... Configuration passed through to the logging constructor
type Config struct {
	UseCustom         bool
	Level             int
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	OutputPaths       []string
	ErrorOutputPaths  []string
}

// NewLogger ... initializes logging from config
func NewLogger(cfg *Config, env string) {
	var zapCfg zap.Config

	if env == "production" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	if cfg != nil && cfg.UseCustom {
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))
		zapCfg.DisableCaller = cfg.DisableCaller
		zapCfg.DisableStacktrace = cfg.DisableStacktrace
		// Sampling not defined in cfg
		zapCfg.Encoding = cfg.Encoding
		// EncoderConfig not defined in cfg
		zapCfg.OutputPaths = cfg.OutputPaths
		zapCfg.ErrorOutputPaths = cfg.ErrorOutputPaths
		// InitialFields not defined in cfg
	}

	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err = zapCfg.Build()
	if err != nil {
		panic("could not initialize logging")
	}
}

// NewContext ... A helper for middleware to create requestId or other context fields
// and return a context which logger can understand.
func NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, loggerKey, WithContext(ctx).With(fields...))
}

// WithContext ... Pass in a context containing values to add to each log message
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(zap.Logger); ok {
		return &ctxLogger
	}
	return logger
}

// NoContext ... A log helper to log when there's no context. Rare case usage
func NoContext() *zap.Logger {
	return logger
}
