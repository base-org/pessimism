package engine_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	pathID = core.MakePathID(0,
		core.MakeProcessID(core.Live, 0, 0, 0),
		core.MakeProcessID(core.Live, 0, 0, 0))
)

func TestGetUUIDs(t *testing.T) {
	am := engine.NewAddressMap()

	id1 := core.NewUUID()
	id2 := core.NewUUID()
	address := common.HexToAddress("0x24")

	err := am.Insert(address, pathID, id1)
	assert.NoError(t, err)

	// Test for found
	ids, err := am.Get(address, pathID)
	assert.NoError(t, err)
	assert.Equal(t, id1, ids[0])

	// Test for multiple
	err = am.Insert(address, pathID, id2)
	assert.NoError(t, err)

	ids, err = am.Get(address, pathID)
	assert.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)

	// Test for not found
	ids, err = am.Get(address, core.PathID{})
	assert.Error(t, err, "should error")
	assert.Empty(t, ids, "should be empty")
}
