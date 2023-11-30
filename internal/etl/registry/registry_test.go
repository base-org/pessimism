package registry_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/stretchr/testify/assert"
)

func TestProcessRegistry(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		test func(*testing.T, *registry.Registry)
	}{
		{
			name:        "Fetch Failure",
			function:    "GetDataTopic",
			description: "When trying to get an invalid register, an error should be returned",

			test: func(t *testing.T, r *registry.Registry) {

				invalidType := core.TopicType(255)
				register, err := r.GetDataTopic(invalidType)

				assert.Error(t, err)
				assert.Nil(t, register)
			},
		},
		{
			name:        "Fetch Success",
			function:    "GetDataTopic",
			description: "When trying to get a register provided a valid register type, a register should be returned",

			test: func(t *testing.T, r *registry.Registry) {

				reg, err := r.GetDataTopic(core.BlockHeader)

				assert.NoError(t, err)
				assert.NotNil(t, reg)
				assert.Equal(t, reg.DataType, core.BlockHeader)
			},
		},
		{
			name:        "Fetch Dependency Path Success",
			function:    "GetDataTopic",
			description: "When trying to get a register dependency path provided a valid register type, a path should be returned",

			test: func(t *testing.T, r *registry.Registry) {

				path, err := r.TopicPath(core.Log)

				assert.NoError(t, err)
				assert.Len(t, path.Path, 2)

				assert.Equal(t, path.Path[1].DataType, core.BlockHeader)
				assert.Equal(t, path.Path[0].DataType, core.Log)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			r := registry.New()
			tc.test(t, r)
		})

	}
}
