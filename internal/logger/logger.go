package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	UseCustom               bool
	Level                   int
	IsProduction            bool
	DisableCaller           bool
	DisableStacktrace       bool
	Encoding                string
	OutputPaths             []string
	ErrorOutputPaths        []string
	EncoderTimeKey          string
	EncoderLevelKey         string
	EncoderNameKey          string
	EncoderCallerKey        string
	EncoderFunctionKey      string
	EncoderMessageKey       string
	EncoderStacktraceKey    string
	EncoderSkipLineEnding   bool
	EncoderLineEnding       string
	EncoderConsoleSeparator string
}

// InitLoggerFromConfig .. initializes logger from config
func InitLoggerFromConfig(cfg *LoggerConfig) (*zap.Logger, error) {
	if cfg.UseCustom {
		return zap.Config{
			Level:             zap.NewAtomicLevelAt(zapcore.Level(cfg.Level)),
			Development:       !cfg.IsProduction,
			DisableCaller:     cfg.DisableCaller,
			DisableStacktrace: cfg.DisableStacktrace,
			// Sampling not set
			Encoding: cfg.Encoding,
			EncoderConfig: zapcore.EncoderConfig{
				//set by config
				MessageKey:       cfg.EncoderMessageKey,
				LevelKey:         cfg.EncoderLevelKey,
				TimeKey:          cfg.EncoderTimeKey,
				NameKey:          cfg.EncoderNameKey,
				CallerKey:        cfg.EncoderCallerKey,
				FunctionKey:      cfg.EncoderFunctionKey,
				StacktraceKey:    cfg.EncoderStacktraceKey,
				SkipLineEnding:   cfg.EncoderSkipLineEnding,
				LineEnding:       cfg.EncoderLineEnding,
				ConsoleSeparator: cfg.EncoderConsoleSeparator,

				// unset by config
				EncodeLevel:  zapcore.CapitalColorLevelEncoder,
				EncodeTime:   zapcore.ISO8601TimeEncoder,
				EncodeCaller: zapcore.ShortCallerEncoder,

				// not included
				// EncodeDuration: ...
				// EncodeName: ...
				// NewReflectedEncoder: ...
			},
			OutputPaths:      cfg.OutputPaths,
			ErrorOutputPaths: cfg.ErrorOutputPaths,
			// InitialFields not set
		}.Build()
	} else if cfg.IsProduction {
		return zap.NewProductionConfig().Build()
	} else {
		return zap.NewDevelopmentConfig().Build()
	}
}
