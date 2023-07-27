package registry_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/stretchr/testify/assert"
)

func Test_ComponentRegistry(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() registry.Registry
		testLogic         func(*testing.T, registry.Registry)
	}{
		{
			name:        "Fetch Failure",
			function:    "GetRegister",
			description: "When trying to get an invalid register, an error should be returned",

			constructionLogic: registry.NewRegistry,
			testLogic: func(t *testing.T, testRegistry registry.Registry) {

				invalidType := core.RegisterType(255)
				register, err := testRegistry.GetRegister(invalidType)

				assert.Error(t, err)
				assert.Nil(t, register)
			},
		},
		{
			name:        "Fetch Success",
			function:    "GetRegister",
			description: "When trying to get a register provided a valid register type, a register should be returned",

			constructionLogic: registry.NewRegistry,
			testLogic: func(t *testing.T, testRegistry registry.Registry) {

				reg, err := testRegistry.GetRegister(core.GethBlock)

				assert.NoError(t, err)
				assert.NotNil(t, reg)
				assert.Equal(t, reg.DataType, core.GethBlock)
			},
		},
		{
			name:        "Fetch Dependency Path Success",
			function:    "GetRegister",
			description: "When trying to get a register dependency path provided a valid register type, a path should be returned",

			constructionLogic: registry.NewRegistry,
			testLogic: func(t *testing.T, testRegistry registry.Registry) {

				path, err := testRegistry.GetDependencyPath(core.EventLog)

				assert.NoError(t, err)
				assert.Len(t, path.Path, 2)

				assert.Equal(t, path.Path[1].DataType, core.GethBlock)
				assert.Equal(t, path.Path[0].DataType, core.EventLog)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testRouter := tc.constructionLogic()
			tc.testLogic(t, testRouter)
		})

	}
}
