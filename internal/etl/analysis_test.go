package etl_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl"
	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_Mergable(t *testing.T) {
	var tests = []struct {
		name         string
		description  string
		construction func() etl.Analyzer
		test         func(t *testing.T, a etl.Analyzer)
	}{
		{
			name:        "Successful Path Merge",
			description: "Mergable function should return true if paths are mergable",
			construction: func() etl.Analyzer {
				r := registry.New()
				return etl.NewAnalyzer(r)
			},
			test: func(t *testing.T, a etl.Analyzer) {
				// Setup test paths
				mockOracle, err := mocks.NewReader(context.Background(), core.BlockHeader)
				assert.NoError(t, err)

				processes := []process.Process{mockOracle}
				id1 := core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))
				id2 := core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))

				testCfg := &core.PathConfig{
					PathType:     core.Live,
					ClientConfig: &core.ClientConfig{},
				}

				p1, err := etl.NewPath(testCfg, id1, processes)
				assert.NoError(t, err)

				p2, err := etl.NewPath(testCfg, id2, processes)
				assert.NoError(t, err)

				assert.True(t, a.Mergable(p1, p2))
			},
		},
		{
			name:        "Failure Path Merge",
			description: "Mergable function should return false when PID's do not match",
			construction: func() etl.Analyzer {
				r := registry.New()
				return etl.NewAnalyzer(r)
			},
			test: func(t *testing.T, a etl.Analyzer) {
				// Setup test paths
				reader, err := mocks.NewReader(context.Background(), core.BlockHeader)
				assert.NoError(t, err)

				processes := []process.Process{reader}
				id1 := core.MakePathID(0, core.MakeProcessID(core.Live, 1, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))
				id2 := core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0))

				testCfg := &core.PathConfig{
					PathType:     core.Live,
					ClientConfig: &core.ClientConfig{},
				}

				p1, err := etl.NewPath(testCfg, id1, processes)
				assert.NoError(t, err)

				p2, err := etl.NewPath(testCfg, id2, processes)
				assert.NoError(t, err)

				assert.False(t, a.Mergable(p1, p2))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.construction()
			test.test(t, a)
		})
	}

}
