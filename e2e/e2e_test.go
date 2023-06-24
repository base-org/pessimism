package e2e_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
)

func Default_Test_Config(t *testing.T) *config.Config {
	return &config.Config{
		Environment:   config.Development,
		BootStrapPath: "",
		SlackURL:      "",

		SvcConfig: &service.Config{L2PollInterval: 300},
		ServerConfig: &server.Config{
			Host: "localhost",
			Port: 8080,
		},
		LoggerConfig: &logging.Config{
			Level: -1,
		},
	}

}

// Test_Node_To_Pessimism
func Test_Node_To_Pessimism(t *testing.T) {
	ctx := context.Background()
	cfg := op_e2e.DefaultSystemConfig(t)

	node, err := op_e2e.NewOpGeth(t, ctx, &cfg)
	if err != nil {
		t.Fatal(err)
	}

	ctx = context.WithValue(
		ctx, core.L2Client, node.L2Client)

	cfg2 := Default_Test_Config(t)
	pess, kill, err := app.NewPessimismApp(ctx, cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer kill()

	pess.Start()

	node.AddL2Block(ctx)
	node.AddL2Block(ctx)
	node.AddL2Block(ctx)
	node.AddL2Block(ctx)
	node.AddL2Block(ctx)

}
