package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TODO(#33): No Unit Tests for Pipeline & ETL Manager Logic
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

				ctx := mocks.Context(context.Background(), ctrl)

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
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testPipeline := tc.constructionLogic()
			tc.testLogic(t, testPipeline)
		})

	}

}
