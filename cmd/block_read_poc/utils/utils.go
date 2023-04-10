package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger ... new logger initialization
func NewLogger() (*zap.Logger, error) {
	var config zap.Config
	var options []zap.Option

	config = zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	options = append(options, zap.AddStacktrace(zap.FatalLevel))

	return config.Build(options...)
}
