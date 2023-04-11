package config

import (
	"fmt"
	"log"
	"math/big"

	"github.com/joho/godotenv"

	"os"
)

type FilePath string

// Config ... Application level configuration defined by `FilePath` value
type Config struct {
	L1RpcEndpoint string
	L2RpcEndpoint string
}

// OracleConfig ... Configuration passed through to an oracle component constructor
type OracleConfig struct {
	RPCEndpoint string
	StartHeight *big.Int
	EndHeight   *big.Int
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
