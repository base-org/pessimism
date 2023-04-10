package config

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

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
	StartHeight *int
	EndHeight   *int
}

// NewConfig ... Initializer
func NewConfig(ctx context.Context, fileName FilePath) *Config {
	logger := ctxzap.Extract(ctx)
	if err := godotenv.Load(string(fileName)); err != nil {
		logger.Fatal("config file not found for file", zap.String("fileName", string(fileName)))
	}

	return &Config{
		L1RpcEndpoint: getEnv(ctx, "L1_RPC_ENDPOINT"),
		L2RpcEndpoint: getEnv(ctx, "L2_RPC_ENDPOINT"),
	}
}

// getEnv ... Reads env var from process environment, panics if not found
func getEnv(ctx context.Context, name string) string {
	logger := ctxzap.Extract(ctx)
	envVar := os.Getenv(name)
	// Not found
	if envVar == "" {
		logger.Fatal("could not find env var given name", zap.String("name", name))
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
