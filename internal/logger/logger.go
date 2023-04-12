package logger

import (
	"fmt"

	"github.com/base-org/pessimism/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger .. initializes logger from config
func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	if !cfg.LoggerUseDefault {
		return zap.Config{
			Level:             StringToAtomicLevel(cfg.LoggerLevel),
			Development:       cfg.Environment == config.Development,
			DisableCaller:     cfg.LoggerDisableCaller,
			DisableStacktrace: cfg.LoggerDisableStacktrace,
			// Sampling not set
			Encoding: cfg.LoggerEncoding,
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:    cfg.LoggerEncoderMessageKey,
				LevelKey:      cfg.LoggerEncoderLevelKey,
				TimeKey:       cfg.LoggerEncoderTimeKey,
				NameKey:       cfg.LoggerEncoderNameKey,
				CallerKey:     cfg.LoggerEncoderCallerKey,
				FunctionKey:   cfg.LoggerEncoderFunctionKey,
				StacktraceKey: cfg.LoggerEncoderStacktraceKey,
				LineEnding:    cfg.LoggerEncoderLineEnding,

				EncodeLevel:  zapcore.CapitalColorLevelEncoder,
				EncodeTime:   zapcore.ISO8601TimeEncoder,
				EncodeCaller: zapcore.ShortCallerEncoder,
			},
			OutputPaths:      cfg.LoggerOutputPaths,
			ErrorOutputPaths: cfg.LoggerErrorOutputPaths,
			// InitialFields not set
		}.Build()
	} else if cfg.Environment == config.Production {
		return zap.NewProductionConfig().Build()
	} else if cfg.Environment == config.Local || cfg.Environment == config.Development {
		return zap.NewDevelopmentConfig().Build()
	}
	return nil, fmt.Errorf("logger not defined")
}

// StringToAtomicLevel ... converts strings to zap levels
func StringToAtomicLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "dpanic":
		return zap.NewAtomicLevelAt(zap.DPanicLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		panic(fmt.Sprintf("error getting log level for zap logger; given: %s, expected: debug,info,warn,error,dpanic,panic,fatal", level))
	}
}
