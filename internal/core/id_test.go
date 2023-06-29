package core_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_Component_ID(t *testing.T) {

	expectedPID := core.ComponentPID([4]byte{1, 1, 1, 1})
	actualID := core.MakeCUUID(1, 1, 1, 1)

	assert.Equal(t, expectedPID, actualID.PID)

	expectedStr := "layer1:backtest:oracle:account_balance"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_Pipeline_ID(t *testing.T) {
	expectedID := core.PipelinePID([9]byte{1, 1, 1, 1, 1, 1, 1, 1, 1})
	actualID := core.MakePUUID(1,
		core.MakeCUUID(1, 1, 1, 1),
		core.MakeCUUID(1, 1, 1, 1))

	assert.Equal(t, expectedID, actualID.PID)

	expectedStr := "backtest::layer1:backtest:oracle:account_balance::layer1:backtest:oracle:account_balance"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_InvSession_ID(t *testing.T) {
	expectedID := core.InvSessionPID([3]byte{1, 2, 1})
	actualID := core.MakeSUUID(1, 2, 1)

	assert.Equal(t, expectedID, actualID.PID)

	expectedStr := "layer1:live:balance_enforcement"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}
