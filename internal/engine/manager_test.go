package engine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/invariant"
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
		registry.NewInvariantTable(),
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

	// Deploy invariant session
	deployCfg := &invariant.DeployConfig{
		InvType:   core.BalanceEnforcement,
		Network:   core.Layer1,
		Stateful:  true,
		StateKey:  &core.StateKey{},
		AlertDest: core.Slack,
		InvParams: isp,
		PUUID:     testPUUID,
	}

	suuid, err := em.DeployInvariantSession(deployCfg)
	assert.NoError(t, err)
	assert.NotNil(t, suuid)

	// Construct invariant input
	invInput := core.InvariantInput{
		PUUID: testPUUID,
		Input: core.TransitData{
			Type:    core.AccountBalance,
			Address: common.HexToAddress("0x69"),
			Value:   float64(666),
		},
	}

	// Send invariant input to event loop
	ingress <- invInput
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
