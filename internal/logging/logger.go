package logging

import (
	"context"
	"runtime"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"path/filepath"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	tagPackage = "package"
)

type LogKey = string

type loggerKeyType int

const loggerKey loggerKeyType = iota

// NOTE - Logger is set to Nop as default to avoid redundant testing
var logger *zap.Logger = zap.NewNop()

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
	return logger
}

// NoContext ... A log helper to log when there's no context. Rare case usage
func NoContext() *zap.Logger {
	return logger
}

func New(env string) *zap.Logger {
	if env == "local" || env == "development" {
		logger = NewDevelopment()
	} else {
		logger = NewProduction()
	}
	return logger
}

func NewProduction() *zap.Logger {
	cfg := zap.NewProductionConfig()

	logger, err := cfg.Build(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		panic(err)
	}

	return logger
}

func NewDevelopment() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := cfg.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}

	return logger
}

// WithPackage adds a package tag to the logger, using the package name of the caller.
func WithPackage(logger *zap.Logger) *zap.Logger {
	const skipOffset = 1 // skip WithPackage

	_, file, _, ok := runtime.Caller(skipOffset)
	if !ok {
		return logger
	}

	packageName := filepath.Base(filepath.Dir(file))
	return WithPackageName(logger, packageName)
}

func WithPackageName(logger *zap.Logger, packageName string) *zap.Logger {
	return logger.With(zap.String(tagPackage, packageName))
}

// WithSpan adds datadog span trace id for datadog https://docs.datadoghq.com/tracing/connect_logs_and_traces/go/
func WithSpan(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		return logger.With(
			zap.String("dd.trace_id", strconv.Itoa(int(spanContext.TraceID()))),
			zap.String("dd.span_id", strconv.Itoa(int(spanContext.SpanID()))),
		)
	}

	return logger
}
