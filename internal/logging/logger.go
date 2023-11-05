package logging

import (
	"context"
	"encoding/json"

	"github.com/base-org/pessimism/internal/core"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type LogKey = string

type loggerKeyType int

const loggerKey loggerKeyType = iota

const (
	messageKey = "message"
	levelKey   = "level"
)

// ID keys used for logging
const (
	AddrKey LogKey = "address"

	Process LogKey = "process_id"
	Path    LogKey = "path_id"
	UUID    LogKey = "uuid"
	Session LogKey = "session_id"
)

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

	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return ctxLogger
	}

	return logger
}

// NoContext ... A log helper to log when there's no context. Rare case usage
func NoContext() *zap.Logger {
	return logger
}

// New ... A helper to create a logger based on environment
func New(env core.Env) {
	// Register our custom encoder. Ignoring this error as we want to override the default string_json encoder
	// Ignoring this error will not break the logger
	_ = zap.RegisterEncoder(StringJSONEncoderName, NewStringJSONEncoder) //nolint:nolintlint

	switch env {
	case core.Local:
		logger = NewLocal()
	case core.Development:
		logger = NewDevelopment()
	case core.Production:
		logger = NewProduction()
	default:
		panic("Invalid environment")
	}
}

// NewProduction ... A logger for production
func NewProduction() *zap.Logger {
	cfg := zap.NewProductionConfig()

	cfg.Encoding = StringJSONEncoderName
	cfg.EncoderConfig.MessageKey = messageKey
	cfg.EncoderConfig.LevelKey = levelKey
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	logger, err := cfg.Build(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		panic(err)
	}

	return logger
}

// NewDevelopment ... A logger for development
func NewDevelopment() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.Encoding = StringJSONEncoderName
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncoderConfig.MessageKey = messageKey
	cfg.EncoderConfig.LevelKey = levelKey

	logger, err := cfg.Build(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		panic(err)
	}

	return logger
}

// NewLocal ... A logger for local development
func NewLocal() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.MessageKey = messageKey

	logger, err := cfg.Build(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		panic(err)
	}

	return logger
}

// StringJSONEncoderName is used for registering this encoder with zap.
const StringJSONEncoderName string = "string_json"

type stringJSONEncoder struct {
	zapcore.Encoder
}

// NewStringJSONEncoder returns an encoder that encodes the JSON log dict as a string
// so the log processing pipeline can correctly process logs with nested JSON.
func NewStringJSONEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return newStringJSONEncoder(cfg), nil
}

func newStringJSONEncoder(cfg zapcore.EncoderConfig) *stringJSONEncoder {
	return &stringJSONEncoder{zapcore.NewJSONEncoder(cfg)}
}

// EncodeEntry ... Encodes the entry and fields, and returns a buffer with the encoded log dict as a string.
func (enc *stringJSONEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	var stringifiedFields []zapcore.Field
	for i := range fields {
		switch fields[i].Type { //nolint:exhaustive // We only care about the types we handle
		// Indicates that the field carries an interface{}
		case zapcore.ReflectType:
			marshaled, err := json.Marshal(fields[i].Interface)
			if err != nil {
				return nil, err
			}
			newField := zap.String(fields[i].Key, string(marshaled))
			stringifiedFields = append(stringifiedFields, newField)
		default:
			stringifiedFields = append(stringifiedFields, fields[i])
		}
	}
	return enc.Encoder.EncodeEntry(ent, stringifiedFields)
}
