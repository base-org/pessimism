package pipeline_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_Mergable(t *testing.T) {
	var tests = []struct {
		name            string
		function        string
		description     string
		testConstructor func() pipeline.Analyzer
		testLogic       func(t *testing.T, a pipeline.Analyzer)
	}{
		{
			name:        "Successful Pipeline Merge",
			function:    "Mergable",
			description: "Mergable function should return true if pipelines are mergable",
			testConstructor: func() pipeline.Analyzer {
				dRegistry := registry.NewRegistry()
				return pipeline.NewAnalyzer(dRegistry)
			},
			testLogic: func(t *testing.T, a pipeline.Analyzer) {
				// Setup test pipelines
				mockOracle, err := mocks.NewMockOracle(context.Background(), core.GethBlock)
				assert.NoError(t, err)

				comps := []component.Component{mockOracle}
				testPUUID := core.MakePipelineUUID(0, core.MakeComponentUUID(0, 0, 0, 0), core.MakeComponentUUID(0, 0, 0, 0))
				p1, err := pipeline.NewPipeline(&core.PipelineConfig{}, testPUUID, comps)
				assert.NoError(t, err)

				p2, err := pipeline.NewPipeline(&core.PipelineConfig{}, core.NilPipelineUUID(), comps)
				assert.NoError(t, err)

				// Test
				assert.True(t, a.Mergable(p1, p2))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.testConstructor()
			test.testLogic(t, a)
		})
	}

}
