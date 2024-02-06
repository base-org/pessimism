package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/subsystem"
	"gopkg.in/yaml.v3"

	indexer_client "github.com/ethereum-optimism/optimism/indexer/client"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// TrueEnvVal ... Represents the encoded string value for true (ie. 1)
const (
	trueEnvVal               = "1"
	defaultEngineWorkerCount = 6
)

// Config ... Application level configuration defined by `FilePath` value
// TODO - Consider renaming to "environment config"
type Config struct {
	Environment   core.Env
	BootStrapPath string

	AlertConfig   *alert.Config
	ClientConfig  *client.Config
	EngineConfig  *engine.Config
	MetricsConfig *metrics.Config
	ServerConfig  *server.Config
	SystemConfig  *subsystem.Config
}

// NewConfig ... Initializer
func NewConfig(fileName core.FilePath) *Config {
	if err := godotenv.Load(string(fileName)); err != nil {
		logging.NoContext().Warn("config file not found for file: %s", zap.Any("file", fileName))
	}

	config := &Config{

		BootStrapPath: getEnvStrWithDefault("BOOTSTRAP_PATH", ""),
		Environment:   core.Env(getEnvStr("ENV")),

		AlertConfig: &alert.Config{
			RoutingCfgPath:          getEnvStrWithDefault("ALERT_ROUTE_CFG_PATH", "alerts-routing.yaml"),
			PagerdutyAlertEventsURL: getEnvStrWithDefault("PAGERDUTY_ALERT_EVENTS_URL", ""),
			RoutingParams:           nil, // This is populated after the config is created (see IngestAlertConfig)
			SNSConfig: &client.SNSConfig{
				TopicArn: getEnvStrWithDefault("SNS_TOPIC_ARN", ""),
				Endpoint: getEnvStrWithDefault("AWS_ENDPOINT", ""),
			},
		},

		ClientConfig: &client.Config{
			L1RpcEndpoint: getEnvStr("L1_RPC_ENDPOINT"),
			L2RpcEndpoint: getEnvStr("L2_RPC_ENDPOINT"),
			IndexerCfg: &indexer_client.Config{
				BaseURL:         getEnvStrWithDefault("INDEXER_URL", ""),
				PaginationLimit: getEnvIntWithDefault("INDEXER_PAGINATION_LIMIT", 0),
			},
		},

		EngineConfig: &engine.Config{
			WorkerCount: getEnvIntWithDefault("ENGINE_WORKER_COUNT", defaultEngineWorkerCount),
		},

		MetricsConfig: &metrics.Config{
			Host:              getEnvStr("METRICS_HOST"),
			Port:              getEnvInt("METRICS_PORT"),
			Enabled:           getEnvBool("ENABLE_METRICS"),
			ReadHeaderTimeout: getEnvInt("METRICS_READ_HEADER_TIMEOUT"),
		},

		ServerConfig: &server.Config{
			Host:         getEnvStr("SERVER_HOST"),
			Port:         getEnvInt("SERVER_PORT"),
			KeepAlive:    getEnvInt("SERVER_KEEP_ALIVE_TIME"),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT"),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT"),
		},

		SystemConfig: &subsystem.Config{
			MaxPathCount:   getEnvInt("MAX_PATH_COUNT"),
			L1PollInterval: getEnvInt("L1_POLL_INTERVAL"),
			L2PollInterval: getEnvInt("L2_POLL_INTERVAL"),
		},
	}

	return config
}

// IsProduction ... Returns true if the env is production
func (cfg *Config) IsProduction() bool {
	return cfg.Environment == core.Production
}

// IsDevelopment ... Returns true if the env is development
func (cfg *Config) IsDevelopment() bool {
	return cfg.Environment == core.Development
}

// IsLocal ... Returns true if the env is local
func (cfg *Config) IsLocal() bool {
	return cfg.Environment == core.Local
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
func getEnvStrWithDefault(key, defaultValue string) string {
	envVar, ok := os.LookupEnv(key)

	// Not found
	if !ok {
		return defaultValue
	}

	return envVar
}

// getEnvIntWithDefault ... Reads env var from process environment, returns default if not found
func getEnvIntWithDefault(key string, defaultValue int) int {
	envVar, ok := os.LookupEnv(key)

	// Not found
	if !ok {
		return defaultValue
	}

	intRep, err := strconv.Atoi(envVar)
	if err != nil {
		log.Fatalf("env val is not int; got: %s=%s; err: %s", key, envVar, err.Error())
	}

	return intRep
}

// getEnvBool ... Reads env vars and converts to booleans
func getEnvBool(key string) bool {
	return getEnvStr(key) == trueEnvVal
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

// IngestAlertConfig ... Ingests an alerting config provided a file path
func (cfg *Config) IngestAlertConfig() error {
	// (1) Error if no routing config path is provided
	if cfg.AlertConfig.RoutingCfgPath == "" && cfg.AlertConfig.RoutingParams == nil {
		return fmt.Errorf("alert routing config path is empty")
	}

	// (2) Return nil if a routing param struct is already provided
	if cfg.AlertConfig.RoutingParams != nil {
		return nil
	}

	// (3) Read the YAML file contents into a routing param struct
	f, err := os.ReadFile(filepath.Clean(cfg.AlertConfig.RoutingCfgPath))
	if err != nil {
		return err
	}

	var params *core.AlertRoutingParams
	err = yaml.Unmarshal(f, &params)
	if err != nil {
		return err
	}

	// (4) Set the routing params and return
	if params != nil {
		cfg.AlertConfig.RoutingParams = params
	}

	return nil
}
