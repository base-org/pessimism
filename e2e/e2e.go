package e2e

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
)

// TestSuite ... Stores all the information needed to run an e2e test
type TestSuite struct {
	t *testing.T

	L2Geth *op_e2e.OpGeth
	L2Cfg  *op_e2e.SystemConfig

	App     *app.Application
	AppCfg  *config.Config
	AppKill func()

	Slack *TestServer
}

// createTestSuite ... Creates a new TestSuite
func CreateTestSuite(t *testing.T) *TestSuite {
	ctx := context.Background()
	nodeCfg := op_e2e.DefaultSystemConfig(t)

	node, err := op_e2e.NewOpGeth(t, ctx, &nodeCfg)
	if err != nil {
		t.Fatal(err)
	}

	ss := state.NewMemState()

	ctx = app.InitializeContext(ctx, ss, node.L2Client, node.L2Client)

	appCfg := DefaultTestConfig()

	slackServer := MockSlackServer()
	appCfg.SlackURL = slackServer.Server.URL

	pess, kill, err := app.NewPessimismApp(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := pess.Start(); err != nil {
		t.Fatal(err)
	}

	go pess.ListenForShutdown(kill)

	logging.NewLogger(appCfg.LoggerConfig, false)

	return &TestSuite{
		t:       t,
		L2Geth:  node,
		L2Cfg:   &nodeCfg,
		App:     pess,
		AppKill: kill,
		AppCfg:  appCfg,
		Slack:   slackServer,
	}
}

// DefaultTestConfig ... Returns a default config for testing
func DefaultTestConfig() *config.Config {
	port := 6969
	pollInterval := 300

	return &config.Config{
		Environment:   config.Development,
		BootStrapPath: "",

		SvcConfig: &service.Config{L2PollInterval: pollInterval},
		ServerConfig: &server.Config{
			Host: "localhost",
			Port: port,
		},
		LoggerConfig: &logging.Config{
			Level: -1,
		},
	}
}
