package config

import (
	"context"
	"log"
	"strings"

	"github.com/joho/godotenv"

	"os"
)

type FilePath string

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
	Local       Env = "local"
)

// Config ... Application level configuration defined by `FilePath` value
type Config struct {
	L1RpcEndpoint string
	L2RpcEndpoint string

	Environment Env

	LoggerUseDefault        bool
	LoggerLevel             string
	LoggerDisableCaller     bool
	LoggerDisableStacktrace bool
	LoggerEncoding          string
	LoggerOutputPaths       []string
	LoggerErrorOutputPaths  []string

	LoggerEncoderTimeKey       string
	LoggerEncoderLevelKey      string
	LoggerEncoderNameKey       string
	LoggerEncoderCallerKey     string
	LoggerEncoderFunctionKey   string
	LoggerEncoderMessageKey    string
	LoggerEncoderStacktraceKey string
	LoggerEncoderLineEnding    string
}

// OracleConfig ... Configuration passed through to an oracle component constructor
type OracleConfig struct {
	RPCEndpoint string
	StartHeight *int
	EndHeight   *int
}

// NewConfig ... Initializer
func NewConfig(ctx context.Context, fileName FilePath) *Config {
	if err := godotenv.Load(string(fileName)); err != nil {
		log.Fatalf("config file not found for file: %s", fileName)
	}

	return &Config{
		L1RpcEndpoint: getEnvStr("L1_RPC_ENDPOINT"),
		L2RpcEndpoint: getEnvStr("L2_RPC_ENDPOINT"),

		Environment: Env(getEnvStr("ENV")),

		LoggerUseDefault:        getEnvBool("LOGGER_USE_DEFAULT"),
		LoggerLevel:             getEnvStr("LOGGER_LEVEL"),
		LoggerDisableCaller:     getEnvBool("LOGGER_DISABLE_CALLER"),
		LoggerDisableStacktrace: getEnvBool("LOGGER_DISABLE_STACKTRACE"),
		LoggerEncoding:          getEnvStr("LOGGER_ENCODING"),
		LoggerOutputPaths:       getEnvSlice("LOGGER_OUTPUT_PATHS"),
		LoggerErrorOutputPaths:  getEnvSlice("LOGGER_ERROR_OUTPUT_PATHS"),

		LoggerEncoderTimeKey:       getEnvStr("LOGGER_ENCODER_TIME_KEY"),
		LoggerEncoderLevelKey:      getEnvStr("LOGGER_ENCODER_LEVEL_KEY"),
		LoggerEncoderNameKey:       getEnvStr("LOGGER_ENCODER_NAME_KEY"),
		LoggerEncoderCallerKey:     getEnvStr("LOGGER_ENCODER_CALLER_KEY"),
		LoggerEncoderFunctionKey:   getEnvStr("LOGGER_ENCODER_FUNCTION_KEY"),
		LoggerEncoderMessageKey:    getEnvStr("LOGGER_ENCODER_MESSAGE_KEY"),
		LoggerEncoderStacktraceKey: getEnvStr("LOGGER_ENCODER_STACKTRACE_KEY"),
		LoggerEncoderLineEnding:    getEnvStr("LOGGER_ENCODER_LINE_ENDING"),
	}
}

// getEnvStr ... Reads env var from process environment, panics if not found
func getEnvStr(key string) string {
	envVar := os.Getenv(key)
	// Not found
	if envVar == "" {
		log.Fatalf("could not find env var given name: %s", key)
	}

	return envVar
}

// getEnvBool .. Reads env vars and converts to booleans, panics if incorrect input
func getEnvBool(key string) bool {
	if key := getEnvStr(key); key == "true" {
		return true
	} else if key == "false" {
		return false
	}
	log.Fatalf("env var given name: %s is not boolean", key)
	return false
}

// getEnvSlice .. Reads env vars and converts to string slice
func getEnvSlice(key string) []string {
	return strings.Split(getEnvStr(key), ",")
}

// func convertToInt(str string) int {
// 	intRep, err := strconv.Atoi(str)

// 	if err != nil {
// 		panic(err)
// 	}

// 	return intRep

// }
