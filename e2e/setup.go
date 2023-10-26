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
	"github.com/golang/mock/gomock"

	"github.com/base-org/pessimism/internal/subsystem"

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
	ctrl                *gomock.Controller
	TestSlackSvr        *TestSlackServer
	TestPagerDutyServer *TestPagerDutyServer
	TestIndexerClient   *mocks.MockIndexerClient

	// Clients
	L1Client *ethclient.Client
	L2Client *ethclient.Client
}

// L2TestSuite ... Stores all the information needed to run an e2e L2Geth test
type L2TestSuite struct {
	t *testing.T

	L2Geth *op_e2e.OpGeth
	L2Cfg  *op_e2e.SystemConfig

	App    *app.Application
	AppCfg *config.Config
	Close  func()

	TestSlackSvr        *TestSlackServer
	TestPagerDutyServer *TestPagerDutyServer
}

// CreateSysTestSuite ... Creates a new L2Geth test suite
func CreateL2TestSuite(t *testing.T) *L2TestSuite {
	ctx := context.Background()
	nodeCfg := op_e2e.DefaultSystemConfig(t)
	logging.New(core.Development)

	node, err := op_e2e.NewOpGeth(t, ctx, &nodeCfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(os.Getenv("ENABLE_ROLLUP_LOGS")) == 0 {
		t.Log("set env 'ENABLE_ROLLUP_LOGS' to show rollup logs")
		for name, logger := range nodeCfg.Loggers {
			t.Logf("discarding logs for %s", name)
			logger.SetHandler(log.DiscardHandler())
		}
	}

	ss := state.NewMemState()

	bundle := &client.Bundle{
		L1Client: node.L2Client,
		L2Client: node.L2Client,
	}
	ctx = app.InitializeContext(ctx, ss, bundle)

	appCfg := DefaultTestConfig()

	slackServer := NewTestSlackServer("127.0.0.1", 0)

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

	go pess.ListenForShutdown(kill)

	return &L2TestSuite{
		t:      t,
		L2Geth: node,
		L2Cfg:  &nodeCfg,
		App:    pess,
		Close: func() {
			kill()
			node.Close()
			slackServer.Close()
			pagerdutyServer.Close()
		},
		AppCfg:              appCfg,
		TestSlackSvr:        slackServer,
		TestPagerDutyServer: pagerdutyServer,
	}
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
	indexerClient := mocks.NewMockIndexerClient(ctrl)

	bundle := &client.Bundle{
		L1Client:      sys.Clients["l1"],
		L2Client:      sys.Clients["sequencer"],
		L2Geth:        gethClient,
		IndexerClient: indexerClient,
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
		TestIndexerClient:   indexerClient,
	}
}

// DefaultTestConfig ... Returns a default app config for testing
func DefaultTestConfig() *config.Config {
	l1PollInterval := 900
	l2PollInterval := 300
	maxPipelines := 10
	workerCount := 4

	return &config.Config{
		Environment:   core.Development,
		BootStrapPath: "",
		AlertConfig: &alert.Config{
			PagerdutyAlertEventsURL: "",
			RoutingCfgPath:          "",
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
			MaxPipelineCount: maxPipelines,
			L2PollInterval:   l2PollInterval,
			L1PollInterval:   l1PollInterval,
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
