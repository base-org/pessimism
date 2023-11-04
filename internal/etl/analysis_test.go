package etl_test

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
				mockOracle, err := mocks.NewReader(context.Background(), core.BlockHeader)
				assert.NoError(t, err)

				comps := []component.Process{mockOracle}
				testPathID := core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))
				testPathID2 := core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))

				testCfg := &core.PathConfig{
					PathType:     core.Live,
					ClientConfig: &core.ClientConfig{},
				}

				p1, err := pipeline.NewPath(testCfg, testPathID, comps)
				assert.NoError(t, err)

				p2, err := pipeline.NewPath(testCfg, testPathID2, comps)
				assert.NoError(t, err)

				assert.True(t, a.Mergable(p1, p2))
			},
		},
		{
			name:        "Failure Pipeline Merge",
			function:    "Mergable",
			description: "Mergable function should return false when PID's do not match",
			testConstructor: func() pipeline.Analyzer {
				dRegistry := registry.NewRegistry()
				return pipeline.NewAnalyzer(dRegistry)
			},
			testLogic: func(t *testing.T, a pipeline.Analyzer) {
				// Setup test pipelines
				mockOracle, err := mocks.NewReader(context.Background(), core.BlockHeader)
				assert.NoError(t, err)

				comps := []component.Process{mockOracle}
				testPathID := core.MakePathID(0, core.MakeProcessID(core.Backtest, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))
				testPathID2 := core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))

				testCfg := &core.PathConfig{
					PathType:     core.Live,
					ClientConfig: &core.ClientConfig{},
				}

				p1, err := pipeline.NewPath(testCfg, testPathID, comps)
				assert.NoError(t, err)

				p2, err := pipeline.NewPath(testCfg, testPathID2, comps)
				assert.NoError(t, err)

				assert.False(t, a.Mergable(p1, p2))
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
