package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Component_ID(t *testing.T) {

	expectedPID := ComponentPID([4]byte{1, 1, 1, 1})
	actualID := MakeCUUID(1, 1, 1, 1)

	assert.Equal(t, expectedPID, actualID.PID)

	expectedStr := "layer1:backtest:oracle:account_balance"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_Pipeline_ID(t *testing.T) {
	expectedID := PipelinePID([9]byte{1, 1, 1, 1, 1, 1, 1, 1, 1})
	actualID := MakePUUID(1,
		MakeCUUID(1, 1, 1, 1),
		MakeCUUID(1, 1, 1, 1))

	assert.Equal(t, expectedID, actualID.PID)

	expectedStr := "backtest::layer1:backtest:oracle:account_balance::layer1:backtest:oracle:account_balance"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_InvSession_ID(t *testing.T) {
	expectedID := InvSessionPID([3]byte{1, 2, 1})
	actualID := MakeSUUID(1, 2, 1)

	assert.Equal(t, expectedID, actualID.PID)

	expectedStr := "layer1:live:unknown"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}
