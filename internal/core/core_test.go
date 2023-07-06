package core_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_TransitData(t *testing.T) {
	// Verify construction
	td := core.NewTransitData(
		core.GethBlock,
		nil,
	)

	assert.NotNil(t, td, "TransitData should not be nil")
	assert.NotNil(t, td.Timestamp, "TransitData timestamp should not be nil")

	// Verify addressing
	addressed := td.Addressed()
	assert.False(t, addressed, "TransitData should not be addressed")

	td.Address = common.HexToAddress("0x456")
	addressed = td.Addressed()
	assert.True(t, addressed, "TransitData should be addressed")
}

func Test_EngineRelay(t *testing.T) {
	outChan := make(chan core.InvariantInput)

	eir := core.NewEngineRelay(core.NilPUUID(), outChan)
	dummyTD := core.NewTransitData(core.AccountBalance, nil)

	// Verify relay and wrapping

	go eir.RelayTransitData(dummyTD)
	invInput := <-outChan

	assert.NotNil(t, invInput, "InvariantInput should not be nil")
	assert.Equal(t, invInput.PUUID, core.NilPUUID(), "InvariantInput PUUID should be nil")
	assert.Equal(t, invInput.Input, dummyTD, "InvariantInput Input should be dummyTD")
}

func Test_InvSessionParams(t *testing.T) {
	isp := core.NewSessionParams()
	assert.NotNil(t, isp, "InvSessionParams should not be nil")

	isp.SetValue("tst", "tst")
	val, err := isp.Value("tst")
	assert.Nil(t, err, "Value should not return an error")
	assert.Equal(t, val, "tst", "Value should return the correct value")

	isp.SetNestedArg("bland(1,2,3)")
	nestedArgs := isp.NestedArgs()
	assert.Equal(t, nestedArgs, []interface{}{"bland(1,2,3)"}, "NestedArgs should return the correct value")

}
