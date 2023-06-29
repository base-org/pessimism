package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/joho/godotenv"
)

type FilePath string

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
	Local       Env = "local"

	// trueEnvVal ... Represents the encoded string value for true (ie. 1)
	trueEnvVal = "1"
)

// Config ... Application level configuration defined by `FilePath` value
// TODO - Consider renaming to "environment config"
type Config struct {
	Environment   Env
	BootStrapPath string
	L1RpcEndpoint string
	L2RpcEndpoint string

	// TODO - Consider moving this URL to a more appropriate location
	SlackURL string

	SvcConfig     *service.Config
	ServerConfig  *server.Config
	MetricsConfig *metrics.Config
	LoggerConfig  *logging.Config
}

// NewConfig ... Initializer
func NewConfig(fileName FilePath) *Config {
	if err := godotenv.Load(string(fileName)); err != nil {
		log.Fatalf("config file not found for file: %s", fileName)
	}

	config := &Config{
		L1RpcEndpoint: getEnvStr("L1_RPC_ENDPOINT"),
		L2RpcEndpoint: getEnvStr("L2_RPC_ENDPOINT"),

		BootStrapPath: getEnvStrWithDefault("BOOTSTRAP_PATH", ""),
		Environment:   Env(getEnvStr("ENV")),
		SlackURL:      getEnvStrWithDefault("SLACK_URL", ""),

		SvcConfig: &service.Config{

			L1PollInterval: getEnvInt("L1_POLL_INTERVAL"),
			L2PollInterval: getEnvInt("L2_POLL_INTERVAL"),
		},

		MetricsConfig: &metrics.Config{
			Host:              getEnvStr("METRICS_HOST"),
			Port:              getEnvInt("METRICS_PORT"),
			Enabled:           getEnvBool("ENABLE_METRICS"),
			ReadHeaderTimeout: getEnvInt("METRICS_READ_HEADER_TIMEOUT"),
		},

		LoggerConfig: &logging.Config{
			UseCustom:         getEnvBool("LOGGER_USE_CUSTOM"),
			Level:             getEnvInt("LOGGER_LEVEL"),
			DisableCaller:     getEnvBool("LOGGER_DISABLE_CALLER"),
			DisableStacktrace: getEnvBool("LOGGER_DISABLE_STACKTRACE"),
			Encoding:          getEnvStr("LOGGER_ENCODING"),
			OutputPaths:       getEnvSlice("LOGGER_OUTPUT_PATHS"),
			ErrorOutputPaths:  getEnvSlice("LOGGER_ERROR_OUTPUT_PATHS"),
		},

		ServerConfig: &server.Config{
			Host:         getEnvStr("SERVER_HOST"),
			Port:         getEnvInt("SERVER_PORT"),
			KeepAlive:    getEnvInt("SERVER_KEEP_ALIVE_TIME"),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT"),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT"),
		},
	}

	return config
}

// IsProduction ... Returns true if the env is production
func (cfg *Config) IsProduction() bool {
	return cfg.Environment == Production
}

// IsDevelopment ... Returns true if the env is development
func (cfg *Config) IsDevelopment() bool {
	return cfg.Environment == Development
}

// IsLocal ... Returns true if the env is local
func (cfg *Config) IsLocal() bool {
	return cfg.Environment == Local
}

// IsBootstrap ... Returns true if a state bootstrap is required
func (cfg *Config) IsBootstrap() bool {
	return cfg.BootStrapPath != ""
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

// getEnvStrWithDefault ... Reads env var from process environment, returns default if not found
func getEnvStrWithDefault(key string, defaultValue string) string {
	envVar, ok := os.LookupEnv(key)

	// Not found
	if !ok {
		return defaultValue
	}

	return envVar
}

// getEnvBool ... Reads env vars and converts to booleans
func getEnvBool(key string) bool {
	return getEnvStr(key) == trueEnvVal
}

// getEnvSlice ... Reads env vars and converts to string slice
func getEnvSlice(key string) []string {
	return strings.Split(getEnvStr(key), ",")
}

// getEnvInt ... Reads env vars and converts to int
func getEnvInt(key string) int {
	val := getEnvStr(key)
	intRep, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("env val is not int; got: %s=%s; err: %s", key, val, err.Error())
	}
	return intRep
}
