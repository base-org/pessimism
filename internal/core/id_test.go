package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Component_ID(t *testing.T) {

	expectedPID := CPID([4]byte{1, 1, 1, 1})
	actualID := MakeComponentID(1, 1, 1, 1)

	assert.Equal(t, expectedPID, actualID.PID)

	expectedStr := "layer1:backtest:oracle:geth.block"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_Pipeline_ID(t *testing.T) {
	expectedID := PipelinePID([9]byte{1, 1, 1, 1, 1, 1, 1, 1, 1})
	actualID := MakePipelineID(1,
		MakeComponentID(1, 1, 1, 1),
		MakeComponentID(1, 1, 1, 1))

	assert.Equal(t, expectedID, actualID.PID)

	expectedStr := "backtest::layer1:backtest:oracle:geth.block::layer1:backtest:oracle:geth.block"
	actualStr := actualID.PID.String()

	assert.Equal(t, expectedStr, actualStr)
}
