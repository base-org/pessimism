package core_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestProcessID(t *testing.T) {

	expected := core.ProcIdentifier([4]byte{1, 1, 1, 1})
	actual := core.MakeProcessID(1, 1, 1, 1)

	assert.Equal(t, expected, actual.ID)

	expectedStr := "layer1:reader:block_header"
	actualStr := actual.Identifier()

	assert.Equal(t, expectedStr, actualStr)
}

func TestPathID(t *testing.T) {
	expected := core.PathIdentifier([9]byte{1, 1, 1, 1, 1, 1, 1, 1, 1})
	actual := core.MakePathID(1,
		core.MakeProcessID(1, 1, 1, 1),
		core.MakeProcessID(1, 1, 1, 1))

	assert.Equal(t, expected, actual.ID)

	expectedStr := "layer1:reader:block_header::layer1:reader:block_header"
	actualStr := actual.Identifier()

	assert.Equal(t, expectedStr, actualStr)
}
