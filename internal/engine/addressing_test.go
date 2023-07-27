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

func Test_GetSUUIDsByPair(t *testing.T) {
	am := engine.NewAddressingMap()

	sUUID := core.NilSUUID()
	address := common.HexToAddress("0x24")

	err := am.Insert(address, testPUUID, sUUID)
	assert.NoError(t, err, "should not error")

	// Test for found
	ids, err := am.GetSUUIDsByPair(address, testPUUID)
	assert.NoError(t, err, "should not error")
	assert.Equal(t, core.NilSUUID(), ids[0], "should be equal")

	// Test for found with multiple entries
	sUUID2 := core.MakeSUUID(0, 0, 1)
	err = am.Insert(address, testPUUID, sUUID2)
	assert.NoError(t, err, "should not error")

	ids, err = am.GetSUUIDsByPair(address, testPUUID)
	assert.NoError(t, err, "should not error")
	assert.Len(t, ids, 2, "should have length of 2")
	assert.Contains(t, ids, sUUID, "should contain sUUID")
	assert.Contains(t, ids, sUUID2, "should contain sUUID2")

	// Test for not found
	ids, err = am.GetSUUIDsByPair(address, core.NilPUUID())
	assert.Error(t, err, "should error")
	assert.Empty(t, ids, "should be empty")
}
