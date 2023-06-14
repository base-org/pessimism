package engine_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	testPUUID = core.MakePUUID(0,
		core.MakeCUUID(core.Live, 0, 0, 0),
		core.MakeCUUID(core.Live, 0, 0, 0))
)

func Test_GetSessionUUIDByPair(t *testing.T) {
	am := engine.NewAddressingMap()

	pUUID := core.NilPUUID()
	sUUID := core.NilSUUID()
	address := common.HexToAddress("0x24")

	err := am.Insert(pUUID, sUUID, address)
	assert.NoError(t, err, "should not error")

	// Test for found
	sUUID, err = am.GetSessionUUIDByPair(address, pUUID)
	assert.NoError(t, err, "should not error")
	assert.Equal(t, core.NilSUUID(), sUUID, "should be equal")

}

func Test_Insert(t *testing.T) {
	am := engine.NewAddressingMap()

	pUUID := core.NilPUUID()
	sUUID := core.NilSUUID()
	address := common.HexToAddress("0x24")

	err := am.Insert(pUUID, sUUID, address)
	assert.NoError(t, err, "should not error")

	// Test for found
	sUUID, err = am.GetSessionUUIDByPair(address, pUUID)
	assert.NoError(t, err, "should not error")
	assert.Equal(t, core.NilSUUID(), sUUID, "should be equal")

	// Test for not found
	sUUID, err = am.GetSessionUUIDByPair(address, testPUUID)
	assert.Error(t, err, "should error")
	assert.Equal(t, core.NilSUUID(), sUUID, "should be equal")
}
