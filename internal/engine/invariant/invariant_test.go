package invariant_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/stretchr/testify/assert"
)

func Test_BaseInvariant(t *testing.T) {
	testSUUID := core.MakeSUUID(1, 1, 1)
	bi := invariant.NewBaseInvariant(core.RegisterType(0))

	// Test SUUID
	bi.SetSUUID(testSUUID)
	actualSUUID := bi.SUUID()

	assert.Equal(t, testSUUID, actualSUUID, "SUUIDs should match")

	// Test InputType
	actualInputType := bi.InputType()
	assert.Equal(t, core.RegisterType(0), actualInputType, "Input types should match")

	// Test validate

	err := bi.ValidateInput(core.TransitData{
		Type: core.RegisterType(0),
	})

	assert.Nil(t, err, "Error should be nil")

	err = bi.ValidateInput(core.TransitData{
		Type: core.RegisterType(1),
	})

	assert.NotNil(t, err, "Error should not be nil")
}
