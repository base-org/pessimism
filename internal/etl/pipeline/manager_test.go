package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/state"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Manager(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() Manager
		testLogic         func(t *testing.T, m Manager)
	}{
		{
			name:        "Successful Pipe Component Construction",
			function:    "inferComponent",
			description: "inferComponent function should generate pipe component instance provided valid params",

			constructionLogic: func() Manager {
				reg := registry.NewRegistry()
				ctrl := gomock.NewController(t)

				ctx, _ := mocks.Context(context.Background(), ctrl)

				ctx = context.WithValue(ctx, core.State, state.NewMemState())

				return NewManager(ctx, NewAnalyzer(reg), reg, NewEtlStore(), NewComponentGraph(), nil)
			},

			testLogic: func(t *testing.T, m Manager) {
				cUUID := core.MakeCUUID(1, 1, 1, 1)

				register, err := registry.NewRegistry().GetRegister(core.GethBlock)

				assert.NoError(t, err)

				cc := &core.ClientConfig{}
				c, err := m.InferComponent(cc, cUUID, core.NilPUUID(), register)
				assert.NoError(t, err)

				assert.Equal(t, c.UUID(), cUUID)
				assert.Equal(t, c.Type(), register.ComponentType)
				assert.Equal(t, c.OutputType(), register.DataType)

			},
		},
		{
			name:        "Successful Pipeline Creations",
			function:    "CreateDataPipeline",
			description: "CreateDataPipeline should reuse existing pipeline if it exists",

			constructionLogic: func() Manager {
				reg := registry.NewRegistry()
				ctrl := gomock.NewController(t)

				ctx, _ := mocks.Context(context.Background(), ctrl)

				ctx = context.WithValue(ctx, core.State, state.NewMemState())

				return NewManager(ctx, NewAnalyzer(reg), reg, NewEtlStore(), NewComponentGraph(), nil)
			},

			testLogic: func(t *testing.T, m Manager) {
				pCfg := &core.PipelineConfig{
					Network:      core.Layer1,
					DataType:     core.EventLog,
					PipelineType: core.Live,
					ClientConfig: &core.ClientConfig{
						Network:      core.Layer1,
						PollInterval: time.Hour * 1,
					},
				}

				pUUID1, reuse, err := m.CreateDataPipeline(pCfg)
				assert.NoError(t, err)
				assert.False(t, reuse)
				assert.NotEqual(t, pUUID1, core.NilPUUID())

				// Now create a new pipeline with the same config
				// & ensure that the previous pipeline is reused

				pUUID2, reuse, err := m.CreateDataPipeline(pCfg)
				assert.NoError(t, err)
				assert.True(t, reuse)
				assert.Equal(t, pUUID1, pUUID2)

				// Now run the pipeline
				err = m.RunPipeline(pUUID1)
				assert.NoError(t, err)

				// Ensure shutdown works
				go func() {
					_ = m.EventLoop()
				}()
				err = m.Shutdown()
				assert.NoError(t, err)
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
