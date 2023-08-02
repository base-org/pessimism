package engine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_EventLoop(t *testing.T) {
	// Setup test dependencies
	alertChan := make(chan core.Alert)
	testPUUID := core.NilPUUID()

	ctx := context.Background()
	ss := state.NewMemState()

	ctx = context.WithValue(ctx, core.State, ss)

	em := engine.NewManager(ctx,
		engine.NewHardCodedEngine(),
		engine.NewAddressingMap(),
		engine.NewSessionStore(),
		registry.NewHeuristicTable(),
		alertChan,
	)

	ingress := em.Transit()

	// Spinup event loop routine w/ closure
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = em.EventLoop()
	}()

	defer func() {
		_ = em.Shutdown()
		wg.Wait()
	}()

	isp := core.NewSessionParams()
	isp.SetValue("address", common.HexToAddress("0x69").String())
	isp.SetValue("upper", 420)

	// Deploy heuristic session
	deployCfg := &heuristic.DeployConfig{
		HeuristicType: core.BalanceEnforcement,
		Network:       core.Layer1,
		Stateful:      true,
		StateKey:      &core.StateKey{},
		AlertingPolicy: &core.AlertPolicy{
			Dest: core.Slack.String(),
		},
		Params: isp,
		PUUID:  testPUUID,
	}

	suuid, err := em.DeployHeuristicSession(deployCfg)
	assert.NoError(t, err)
	assert.NotNil(t, suuid)

	// Construct heuristic input
	hi := core.HeuristicInput{
		PUUID: testPUUID,
		Input: core.TransitData{
			Type:    core.AccountBalance,
			Address: common.HexToAddress("0x69"),
			Value:   float64(666),
		},
	}

	// Send heuristic input to event loop
	ingress <- hi
	ticker := time.NewTicker(1 * time.Second)

	// Receive alert from event loop
	select {
	case <-ticker.C:
		assert.FailNow(t, "Timed out waiting for alert data")

	case alert := <-alertChan:
		assert.NotNil(t, alert)
		assert.Equal(t, alert.PUUID, testPUUID)
	}
}
