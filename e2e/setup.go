package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/state"
	"github.com/base-org/pessimism/internal/subsystem"
	ix_node "github.com/ethereum-optimism/optimism/indexer/node"
	"github.com/golang/mock/gomock"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// SysTestSuite ... Stores all the information needed to run an e2e system test
type SysTestSuite struct {
	t *testing.T

	// Optimism
	Cfg *op_e2e.SystemConfig
	Sys *op_e2e.System

	// Pessimism
	App        *app.Application
	AppCfg     *config.Config
	Subsystems *subsystem.Manager
	Close      func()

	// Mocked services
	TestSlackSvr        *TestSlackServer
	TestPagerDutyServer *TestPagerDutyServer
	TestIxClient        *mocks.MockIxClient

	// Clients
	L1Client *ethclient.Client
	L2Client *ethclient.Client
}

// CreateSysTestSuite ... Creates a new SysTestSuite
func CreateSysTestSuite(t *testing.T) *SysTestSuite {
	t.Log("Creating system test suite")
	ctx := context.Background()
	logging.New(core.Development)

	cfg := op_e2e.DefaultSystemConfig(t)
	cfg.DeployConfig.FinalizationPeriodSeconds = 2

	// Don't output rollup service logs unless ENABLE_ROLLUP_LOGS is set
	if len(os.Getenv("ENABLE_ROLLUP_LOGS")) == 0 {
		t.Log("set env 'ENABLE_ROLLUP_LOGS' to show rollup logs")
		for name, logger := range cfg.Loggers {
			t.Logf("discarding logs for %s", name)
			logger.SetHandler(log.DiscardHandler())
		}
	}

	sys, err := cfg.Start(t)
	if err != nil {
		t.Fatal(err)
	}

	gethClient, err := client.NewGethClient(sys.EthInstances["sequencer"].HTTPEndpoint())
	if err != nil {
		t.Fatal(err)
	}

	ss := state.NewMemState()

	ctrl := gomock.NewController(t)
	ixClient := mocks.NewMockIxClient(ctrl)

	l2NodeClient, err := ix_node.DialEthClient(sys.EthInstances["sequencer"].HTTPEndpoint(), metrics.NoopMetrics)
	if err != nil {
		t.Fatal(err)
	}

	l1NodeClient, err := ix_node.DialEthClient(sys.EthInstances["l1"].HTTPEndpoint(), metrics.NoopMetrics)
	if err != nil {
		t.Fatal(err)
	}

	bundle := &client.Bundle{
		L1Node:   l1NodeClient,
		L2Node:   l2NodeClient,
		L1Client: sys.Clients["l1"],
		L2Client: sys.Clients["sequencer"],
		L2Geth:   gethClient,
		IxClient: ixClient,
	}

	ctx = app.InitializeContext(ctx, ss, bundle)

	appCfg := DefaultTestConfig()

	// Use unstructured slack server responses for testing E2E system flows
	slackServer := NewTestSlackServer("127.0.0.1", 0)
	slackServer.Unstructured = true

	pagerdutyServer := NewTestPagerDutyServer("127.0.0.1", 0)

	slackURL := fmt.Sprintf("http://127.0.0.1:%d", slackServer.Port)
	pagerdutyURL := fmt.Sprintf("http://127.0.0.1:%d", pagerdutyServer.Port)

	appCfg.AlertConfig.PagerdutyAlertEventsURL = pagerdutyURL
	appCfg.AlertConfig.RoutingParams = DefaultRoutingParams(core.StringFromEnv(slackURL))

	pess, kill, err := app.NewPessimismApp(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := pess.Start(); err != nil {
		t.Fatal(err)
	}

	t.Parallel()
	go pess.ListenForShutdown(kill)

	return &SysTestSuite{
		t:   t,
		Sys: sys,
		Cfg: &cfg,
		App: pess,
		Close: func() {
			ctrl.Finish()
			kill()
			sys.Close()
			slackServer.Close()
			pagerdutyServer.Close()
		},
		AppCfg:              appCfg,
		Subsystems:          pess.Subsystems,
		TestSlackSvr:        slackServer,
		TestPagerDutyServer: pagerdutyServer,
		L1Client:            sys.Clients["l1"],
		L2Client:            sys.Clients["sequencer"],
		TestIxClient:        ixClient,
	}
}

// DefaultTestConfig ... Returns a default app config for testing
func DefaultTestConfig() *config.Config {
	l1PollInterval := 900
	l2PollInterval := 300
	maxPaths := 10
	workerCount := 4

	return &config.Config{
		Environment:   core.Development,
		BootStrapPath: "",
		AlertConfig: &alert.Config{
			PagerdutyAlertEventsURL: "",
			RoutingCfgPath:          "",
			SNSConfig: &client.SNSConfig{
				TopicArn: "e2e-test-arn",
			},
		},
		EngineConfig: &engine.Config{
			WorkerCount: workerCount,
		},
		MetricsConfig: &metrics.Config{
			Enabled: false,
			Host:    "localhost",
			Port:    0,
		},
		ServerConfig: &server.Config{
			Host: "localhost",
			Port: 0,
		},
		SystemConfig: &subsystem.Config{
			MaxPathCount:   maxPaths,
			L2PollInterval: l2PollInterval,
			L1PollInterval: l1PollInterval,
		},
	}
}

// DefaultRoutingParams ... Returns a default routing params configuration for testing
func DefaultRoutingParams(slackURL core.StringFromEnv) *core.AlertRoutingParams {
	return &core.AlertRoutingParams{
		AlertRoutes: &core.SeverityMap{
			Low: &core.AlertClientCfg{
				Slack: map[string]*core.AlertConfig{
					"config": {
						URL:     slackURL,
						Channel: "#test-low",
					},
				},
			},
			Medium: &core.AlertClientCfg{
				Slack: map[string]*core.AlertConfig{
					"config": {
						URL:     slackURL,
						Channel: "#test-medium",
					},
				},
				PagerDuty: map[string]*core.AlertConfig{
					"config": {
						IntegrationKey: "test-medium",
					},
				},
			},
			High: &core.AlertClientCfg{
				Slack: map[string]*core.AlertConfig{
					"config": {
						URL:     slackURL,
						Channel: "#test-high",
					},
					"config_2": {
						URL:     slackURL,
						Channel: "#test-high-2",
					},
				},
				PagerDuty: map[string]*core.AlertConfig{
					"config": {
						IntegrationKey: "test-high-1",
					},
					"config_2": {
						IntegrationKey: "test-high-2",
					},
				},
			},
		},
	}
}
