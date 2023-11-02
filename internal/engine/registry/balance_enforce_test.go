package registry_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/stretchr/testify/assert"
)

func Test_Balance_Assess(t *testing.T) {
	upper := float64(5)
	lower := float64(1)

	bi, err := registry.NewBalanceHeuristic(
		&registry.BalanceInvConfig{
			Address:    "0x123",
			UpperBound: &upper,
			LowerBound: &lower,
		})

	assert.NoError(t, err)

	// No activation
	testData1 := core.TransitData{
		Type:  core.AccountBalance,
		Value: float64(3),
	}

	as, err := bi.Assess(testData1)
	assert.NoError(t, err)
	assert.False(t, as.Activated())

	// Upper bound activation
	testData2 := core.TransitData{
		Type:  core.AccountBalance,
		Value: float64(6),
	}

	as, err = bi.Assess(testData2)
	assert.NoError(t, err)
	assert.True(t, as.Activated())

	// Lower bound activation
	testData3 := core.TransitData{
		Type:  core.AccountBalance,
		Value: float64(0.1),
	}

	as, err = bi.Assess(testData3)
	assert.NoError(t, err)
	assert.True(t, as.Activated())
}
