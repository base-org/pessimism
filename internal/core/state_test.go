package core_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestStateKey(t *testing.T) {
	sk := core.MakeStateKey(0, "testId", false)

	id := core.PathID{}
	id.ID[0] = 1

	// Successfully set the key
	err := sk.SetPathID(id)
	assert.NoError(t, err)
	assert.Contains(t, sk.String(), id.String())

	// Fail to set key again
	id = core.PathID{}

	err = sk.SetPathID(id)
	assert.Error(t, err, "cannot set PathID more than once")
}
