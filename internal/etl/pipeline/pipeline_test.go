package pipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_Pipeline(t *testing.T) {
	var tests = []struct {
		name     string
		function string

		constructionLogic func() pipeline.Pipeline
		testLogic         func(t *testing.T, pl pipeline.Pipeline)
	}{
		{
			name:     "Successful Run",
			function: "RunPipeline",
			constructionLogic: func() pipeline.Pipeline {
				testPipe, _ := mocks.NewDummyPipe(
					context.Background(),
					core.GethBlock,
					core.EventLog)

				testO, _ := mocks.NewDummyOracle(
					context.Background(),
					core.GethBlock)

				pl, err := pipeline.NewPipeline(
					nil,
					core.NilPUUID(),
					[]component.Component{testPipe, testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl pipeline.Pipeline) {

				assert.Equal(t, pl.Components()[0].OutputType(), core.EventLog)
				assert.Equal(t, pl.Components()[1].OutputType(), core.GethBlock)
			},
		},
		{
			name:     "Successful Run",
			function: "Add Engine Relay",
			constructionLogic: func() pipeline.Pipeline {

				testO, _ := mocks.NewDummyOracle(
					context.Background(),
					core.GethBlock)

				pl, err := pipeline.NewPipeline(
					nil,
					core.NilPUUID(),
					[]component.Component{testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl pipeline.Pipeline) {

				relay := make(chan core.InvariantInput)
				err := pl.AddEngineRelay(relay)
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
