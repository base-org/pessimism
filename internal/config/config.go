package config

import (
	"log"
	"strconv"
	"strings"

	"github.com/base-org/pessimism/internal/logger"
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

	LoggerConfig *logger.Config
}

// OracleConfig ... Configuration passed through to an oracle component constructor
type OracleConfig struct {
	RPCEndpoint string
	StartHeight *int
	EndHeight   *int
}

// NewConfig ... Initializer
func NewConfig(fileName FilePath) *Config {
	if err := godotenv.Load(string(fileName)); err != nil {
		log.Fatalf("config file not found for file: %s", fileName)
	}

	config := &Config{
		L1RpcEndpoint: getEnvStr("L1_RPC_ENDPOINT"),
		L2RpcEndpoint: getEnvStr("L2_RPC_ENDPOINT"),

		Environment: Env(getEnvStr("ENV")),
	}

	config.LoggerConfig = &logger.Config{
		UseCustom:         getEnvBool("LOGGER_USE_CUSTOM"),
		Level:             getEnvInt("LOGGER_LEVEL"),
		IsProduction:      config.Environment == Production,
		DisableCaller:     getEnvBool("LOGGER_DISABLE_CALLER"),
		DisableStacktrace: getEnvBool("LOGGER_DISABLE_STACKTRACE"),
		Encoding:          getEnvStr("LOGGER_ENCODING"),
		OutputPaths:       getEnvSlice("LOGGER_OUTPUT_PATHS"),
		ErrorOutputPaths:  getEnvSlice("LOGGER_ERROR_OUTPUT_PATHS"),

		EncoderTimeKey:          getEnvStr("LOGGER_ENCODER_TIME_KEY"),
		EncoderLevelKey:         getEnvStr("LOGGER_ENCODER_LEVEL_KEY"),
		EncoderNameKey:          getEnvStr("LOGGER_ENCODER_NAME_KEY"),
		EncoderCallerKey:        getEnvStr("LOGGER_ENCODER_CALLER_KEY"),
		EncoderFunctionKey:      getEnvStr("LOGGER_ENCODER_FUNCTION_KEY"),
		EncoderMessageKey:       getEnvStr("LOGGER_ENCODER_MESSAGE_KEY"),
		EncoderStacktraceKey:    getEnvStr("LOGGER_ENCODER_STACKTRACE_KEY"),
		EncoderSkipLineEnding:   getEnvBool("LOGGER_ENCODER_SKIP_LINE_ENDING"),
		EncoderLineEnding:       getEnvStr("LOGGER_ENCODER_LINE_ENDING"),
		EncoderConsoleSeparator: getEnvStr("LOGGER_ENCODER_CONSOLE_SEPARATOR"),
	}

	return config
}

// getEnvStr ... Reads env var from process environment, panics if not found
func getEnvStr(key string) string {
	envVar, ok := os.LookupEnv(key)

	// Not found
	if !ok {
		log.Fatalf("could not find env var given key: %s", key)
	}

	return envVar
}

// getEnvBool .. Reads env vars and converts to booleans, panics if incorrect input
func getEnvBool(key string) bool {
	if val := getEnvStr(key); val == "1" {
		return true
	} else if val == "0" {
		return false
	} else {
		log.Fatalf("env val is not boolean (0 or 1); got: %s=%s", key, val)
	}
	return false
}

// getEnvSlice .. Reads env vars and converts to string slice
func getEnvSlice(key string) []string {
	return strings.Split(getEnvStr(key), ",")
}

func getEnvInt(key string) int {
	val := getEnvStr(key)
	intRep, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("env val is not int; got: %s=%s", key, val)
	}
	return intRep
}
