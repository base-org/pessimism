package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Component_ID(t *testing.T) {

	expectedID := ComponentID([4]byte{0, 0, 0, 0})
	actualID := MakeComponentID(0, 0, 0, 0)

	assert.Equal(t, expectedID, actualID)

	expectedStr := "unknown:unknown:unknown:unknown"
	actualStr := actualID.String()

	assert.Equal(t, expectedStr, actualStr)
}

func Test_Pipeline_ID(t *testing.T) {
	expectedID := PipelineID([9]byte{0, 0, 0, 0, 0, 0, 0, 0, 0})
	actualID := MakePipelineID(0, ComponentID{0, 0, 0, 0}, ComponentID{0, 0, 0, 0})

	assert.Equal(t, expectedID, actualID)

	expectedStr := "unknown::unknown:unknown:unknown:unknown::unknown:unknown:unknown:unknown"
	actualStr := actualID.String()

	assert.Equal(t, expectedStr, actualStr)
}
