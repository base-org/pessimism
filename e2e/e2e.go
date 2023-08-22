package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"
	"github.com/base-org/pessimism/internal/subsystem"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SysTestSuite ... Stores all the information needed to run an e2e system test
type SysTestSuite struct {
	t *testing.T

	Cfg *op_e2e.SystemConfig
	Sys *op_e2e.System

	App    *app.Application
	AppCfg *config.Config
	Close  func()

	TestSlackSvr        *TestSlackServer
	TestPagerDutyServer *TestPagerDutyServer
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

	node, err := op_e2e.NewOpGeth(t, ctx, &nodeCfg)
	if err != nil {
		t.Fatal(err)
	}

	ss := state.NewMemState()
	ctx = app.InitializeContext(ctx, ss, node.L2Client, node.L2Client, nil)

	appCfg := DefaultTestConfig()

	slackServer := NewTestSlackServer()

	pagerdutyServer := NewTestPagerDutyServer()

	appCfg.AlertConfig.RoutingCfgPath = "alert-routing-cfg.yaml"
	appCfg.AlertConfig.PagerdutyAlertEventsURL = pagerdutyServer.Server.URL

	pess, kill, err := app.NewPessimismApp(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}

	UpdateSlackUrls(appCfg.AlertConfig, slackServer.Server.URL)

	if err := pess.Start(); err != nil {
		t.Fatal(err)
	}

	go pess.ListenForShutdown(kill)

	logging.New(core.Development)

	return &L2TestSuite{
		t:      t,
		L2Geth: node,
		L2Cfg:  &nodeCfg,
		App:    pess,
		Close: func() {
			kill()
			node.Close()
		},
		AppCfg:              appCfg,
		TestSlackSvr:        slackServer,
		TestPagerDutyServer: pagerdutyServer,
	}
}

// CreateSysTestSuite ... Creates a new SysTestSuite
func CreateSysTestSuite(t *testing.T) *SysTestSuite {
	ctx := context.Background()

	cfg := op_e2e.DefaultSystemConfig(t)
	cfg.DeployConfig.FinalizationPeriodSeconds = 6

	sys, err := cfg.Start()
	if err != nil {
		t.Fatal(err)
	}

	gethClient, err := client.NewGethClient(sys.Nodes["sequencer"].HTTPEndpoint())
	if err != nil {
		t.Fatal(err)
	}

	ss := state.NewMemState()
	ctx = app.InitializeContext(ctx, ss,
		sys.Clients["l1"],
		sys.Clients["sequencer"], gethClient)

	appCfg := DefaultTestConfig()

	slackServer := NewTestSlackServer()

	pagerdutyServer := NewTestPagerDutyServer()

	appCfg.AlertConfig.RoutingCfgPath = "alert-routing-cfg.yaml"
	appCfg.AlertConfig.PagerdutyAlertEventsURL = pagerdutyServer.Server.URL

	pess, kill, err := app.NewPessimismApp(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}

	UpdateSlackUrls(appCfg.AlertConfig, slackServer.Server.URL)

	if err := pess.Start(); err != nil {
		t.Fatal(err)
	}

	go pess.ListenForShutdown(kill)

	logging.New(core.Development)

	return &SysTestSuite{
		t:   t,
		Sys: sys,
		Cfg: &cfg,
		App: pess,
		Close: func() {
			kill()
			sys.Close()
		},
		AppCfg:              appCfg,
		TestSlackSvr:        slackServer,
		TestPagerDutyServer: pagerdutyServer,
	}
}

// DefaultTestConfig ... Returns a default app config for testing
func DefaultTestConfig() *config.Config {
	port := 6980
	metPort := 6300
	l1PollInterval := 900
	l2PollInterval := 300
	maxPipelines := 10

	return &config.Config{
		Environment:   core.Development,
		BootStrapPath: "",
		SystemConfig: &subsystem.Config{
			MaxPipelineCount: maxPipelines,
			L2PollInterval:   l2PollInterval,
			L1PollInterval:   l1PollInterval,
		},
		MetricsConfig: &metrics.Config{
			Enabled: false,
			Host:    "localhost",
			Port:    metPort,
		},
		ServerConfig: &server.Config{
			Host: "localhost",
			Port: port,
		},
		AlertConfig: &alert.Config{
			PagerdutyAlertEventsURL: "",
			RoutingCfgPath:          "",
		},
	}
}

// WaitForTransaction ... Waits for a transaction receipt to be generated or times out
func WaitForTransaction(hash common.Hash, client *ethclient.Client, timeout time.Duration) (*types.Receipt, error) {
	timeoutCh := time.After(timeout)
	ms100 := 100

	ticker := time.NewTicker(time.Duration(ms100) * time.Millisecond)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if receipt != nil && err == nil {
			return receipt, nil
		} else if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}

		select {
		case <-timeoutCh:
			return nil, errors.New("timeout")
		case <-ticker.C:
		}
	}
}

func UpdateSlackUrls(cfg *alert.Config, url string) {
	for i := range cfg.AlertRoutingParams.AlertRoutes.High.Slack {
		cfg.AlertRoutingParams.AlertRoutes.High.Slack[i].URL = url
	}

	for i := range cfg.AlertRoutingParams.AlertRoutes.Medium.Slack {
		cfg.AlertRoutingParams.AlertRoutes.Medium.Slack[i].URL = url
	}

	for i := range cfg.AlertRoutingParams.AlertRoutes.Low.Slack {
		cfg.AlertRoutingParams.AlertRoutes.Low.Slack[i].URL = url
	}
}
