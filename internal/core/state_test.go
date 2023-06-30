package core_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_StateKey(t *testing.T) {
	sk := core.MakeStateKey(0, "testId", false)

	puuid := core.NilPUUID()
	puuid.PID[0] = 1

	// Successfully set the key
	err := sk.SetPUUID(puuid)
	assert.NoError(t, err)
	assert.Contains(t, sk.String(), puuid.String())

	pUUID2 := core.NilPUUID()

	err = sk.SetPUUID(pUUID2)
	assert.Error(t, err, "cannot set puuid more than once")
}
