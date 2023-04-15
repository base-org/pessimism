package config

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/joho/godotenv"

	"os"
)

type FilePath string

type Env string

const (
	Development Env = "development"
	Production      = "production"
	Local           = "local"
)

// Config ... Application level configuration defined by `FilePath` value
// TODO - Consider renaming to "environment config"
type Config struct {
	L1RpcEndpoint string
	L2RpcEndpoint string
	Environment   Env
	LoggerConfig  *logging.Config
}

func (c *Config) GetEndpointForNetwork(n core.Network) (string, error) {
	switch n {
	case core.Layer1:
		return c.L1RpcEndpoint, nil

	case core.Layer2:
		return c.L2RpcEndpoint, nil
	}

	return "", fmt.Errorf("could not find endpoint for network: %s", n.String())
}

// OracleConfig ... Configuration passed through to an oracle component constructor
type OracleConfig struct {
	RPCEndpoint  string
	StartHeight  *big.Int
	EndHeight    *big.Int
	NumOfRetries int
}

// PipelineConfig ... Configuration passed through to a pipeline constructor
type PipelineConfig struct {
	Network      core.Network
	DataType     core.RegisterType
	PipelineType core.PipelineType
	OracleCfg    *OracleConfig
}

func (oc *OracleConfig) Backfill() bool {
	return oc.StartHeight != nil
}

func (oc *OracleConfig) Backtest() bool {
	return oc.EndHeight != nil
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

		LoggerConfig: &logging.Config{
			UseCustom:         getEnvBool("LOGGER_USE_CUSTOM"),
			Level:             getEnvInt("LOGGER_LEVEL"),
			DisableCaller:     getEnvBool("LOGGER_DISABLE_CALLER"),
			DisableStacktrace: getEnvBool("LOGGER_DISABLE_STACKTRACE"),
			Encoding:          getEnvStr("LOGGER_ENCODING"),
			OutputPaths:       getEnvSlice("LOGGER_OUTPUT_PATHS"),
			ErrorOutputPaths:  getEnvSlice("LOGGER_ERROR_OUTPUT_PATHS"),
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

// getEnvStr ... Reads env var from process environment, panics if not found
func getEnvStr(key string) string {
	envVar, ok := os.LookupEnv(key)

	// Not found
	if !ok {
		log.Fatalf("could not find env var given key: %s", key)
	}

	return envVar
}

// getEnvBool ... Reads env vars and converts to booleans
func getEnvBool(key string) bool {
	if val := getEnvStr(key); val == "1" {
		return true
	}
	return false
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
