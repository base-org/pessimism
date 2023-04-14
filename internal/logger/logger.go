package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig ... Configuration passed through to the logger constructor
type Config struct {
	UseCustom         bool
	Level             int
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	OutputPaths       []string
	ErrorOutputPaths  []string
}

// NewLogger ... initializes logger from config
func NewLogger(cfg *Config, isProduction bool) (*zap.Logger, error) {
	var zapCfg zap.Config

	if isProduction {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	if cfg.UseCustom {
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

	return zapCfg.Build()
}
