package registry_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/stretchr/testify/assert"
)

func Test_Balance_Invalidate(t *testing.T) {
	upper := float64(5)
	lower := float64(1)

	bi := registry.NewBalanceInvariant(&registry.BalanceInvConfig{
		Address:    "0x123",
		UpperBound: &upper,
		LowerBound: &lower,
	})

	// No invalidation
	testData1 := core.TransitData{
		Type:  core.AccountBalance,
		Value: float64(3),
	}

	_, inval, err := bi.Invalidate(testData1)
	assert.NoError(t, err)
	assert.False(t, inval)

	// Upper bound invalidation
	testData2 := core.TransitData{
		Type:  core.AccountBalance,
		Value: float64(6),
	}

	_, inval, err = bi.Invalidate(testData2)
	assert.NoError(t, err)
	assert.True(t, inval)

	// Lower bound invalidation
	testData3 := core.TransitData{
		Type:  core.AccountBalance,
		Value: float64(0.1),
	}

	_, inval, err = bi.Invalidate(testData3)
	assert.NoError(t, err)
	assert.True(t, inval)
}
