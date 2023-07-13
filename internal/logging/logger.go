package logging

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type LogKey = string

type loggerKeyType int

const loggerKey loggerKeyType = iota

// NOTE - Logger is set to Nop as default to avoid redundant testing
var logger *zap.Logger = zap.NewNop()

const (
	messageKey = "message"
)

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
	_ = zap.RegisterEncoder(StringJSONEncoderName, NewStringJSONEncoder) //nolint:nolintlint

	var zapCfg zap.Config

	if env == "production" {
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = StringJSONEncoderName

	} else if env == "development" {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Encoding = StringJSONEncoderName

	} else if env == "local" {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

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

	zapCfg.EncoderConfig.MessageKey = messageKey
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err = zapCfg.Build(zap.AddStacktrace(zap.FatalLevel))
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
