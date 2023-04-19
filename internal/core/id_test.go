package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Component_ID(t *testing.T) {

	expectedID := ComponentID([4]byte{1, 1, 1, 1})
	actualID := MakeComponentID(1, 1, 1, 1)

	assert.Equal(t, expectedID, actualID)

	expectedStr := "layer1:backtest:oracle:geth.block"
	actualStr := actualID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_Pipeline_ID(t *testing.T) {
	expectedID := PipelineID([9]byte{1, 2, 3, 4, 1, 2, 3, 4, 0})
	actualID := MakePipelineID(1, ComponentID{2, 3, 4, 1}, ComponentID{2, 3, 4, 0})

	assert.Equal(t, expectedID, actualID)

	expectedStr := "backtest::layer2:mocktest:unknown:geth.block::layer2:mocktest:unknown:unknown"
	actualStr := actualID.String()

	assert.Equal(t, expectedStr, actualStr)
}
