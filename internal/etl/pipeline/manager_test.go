package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/stretchr/testify/assert"
)

// TODO(#33): No Unit Tests for Pipeline & ETL Manager Logic
func Test_Manager(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() *Manager
		testLogic         func(t *testing.T, m *Manager)
	}{
		{
			name:        "Successful Pipe Component Construction",
			function:    "inferComponent",
			description: "inferComponent function should generate pipe component instance provided valid params",

			constructionLogic: func() *Manager {
				m, _ := NewManager(context.Background(), nil)
				return m
			},

			testLogic: func(t *testing.T, m *Manager) {
				cUUID := core.MakeComponentUUID(1, 1, 1, 1)

				register, err := registry.GetRegister(core.ContractCreateTX)
				assert.NoError(t, err)

				c, err := inferComponent(context.Background(), nil, cUUID, register)
				assert.NoError(t, err)

				assert.Equal(t, c.ID(), cUUID)
				assert.Equal(t, c.Type(), register.ComponentType)
				assert.Equal(t, c.OutputType(), register.DataType)

			},
		},
		{
			name:        "Successful Oracle Component Construction",
			function:    "inferComponent",
			description: "inferComponent function should generate oracle component instance provided valid params",

			constructionLogic: func() *Manager {
				m, _ := NewManager(context.Background(), nil)
				return m
			},

			testLogic: func(t *testing.T, m *Manager) {
				cUUID := core.MakeComponentUUID(1, 1, 1, 1)

				register, err := registry.GetRegister(core.GethBlock)
				assert.NoError(t, err)

				c, err := inferComponent(context.Background(),
					&core.PipelineConfig{}, cUUID, register)
				assert.NoError(t, err)

				assert.Equal(t, c.ID(), cUUID)
				assert.Equal(t, c.Type(), register.ComponentType)
				assert.Equal(t, c.OutputType(), register.DataType)

			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testPipeline := tc.constructionLogic()
			tc.testLogic(t, testPipeline)
		})

	}

}
