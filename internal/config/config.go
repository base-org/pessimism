package config

import (
	"fmt"
	"log"

	"github.com/base-org/pessimism/internal/core"
	"github.com/joho/godotenv"

	"os"
)

type FilePath string

// Config ... Application level configuration defined by `FilePath` value
// TODO - Consider renaming to "environment config"
type Config struct {
	L1RpcEndpoint string
	L2RpcEndpoint string
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
	RPCEndpoint string
	StartHeight *int
	EndHeight   *int
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
		log.Printf("Config file not found for file name: %s", fileName)
		panic(err)
	}

	return &Config{
		L1RpcEndpoint: getEnv("L1_RPC_ENDPOINT"),
		L2RpcEndpoint: getEnv("L2_RPC_ENDPOINT"),
	}
}

// getEnv ... Reads env var from process environment, panics if not found
func getEnv(name string) string {
	envVar := os.Getenv(name)
	// Not found
	if envVar == "" {
		panic(fmt.Sprintf("Could not find env var for %s", name))
	}

	return envVar
}

// func convertToInt(str string) int {
// 	intRep, err := strconv.Atoi(str)

// 	if err != nil {
// 		panic(err)
// 	}

// 	return intRep

// }
