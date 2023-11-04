package etl_test

import (
	"context"
	"fmt"
	"sync"
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
			name:     "Successful Construction",
			function: "NewPath",
			constructionLogic: func() pipeline.Pipeline {
				sub, _ := mocks.NewSubscriber(
					context.Background(),
					core.BlockHeader,
					core.Log)

				testO, _ := mocks.NewReader(
					context.Background(),
					core.BlockHeader)

				pl, err := pipeline.NewPath(
					nil,
					core.PathID{},
					[]component.Process{sub, testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl pipeline.Pipeline) {

				assert.Equal(t, pl.Components()[0].OutputType(), core.Log)
				assert.Equal(t, pl.Components()[1].OutputType(), core.BlockHeader)
			},
		},
		{
			name:     "Successful Run",
			function: "AddEngineRelay",
			constructionLogic: func() pipeline.Pipeline {

				testO, _ := mocks.NewReader(
					context.Background(),
					core.BlockHeader)

				pl, err := pipeline.NewPath(
					nil,
					core.PathID{},
					[]component.Process{testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl pipeline.Pipeline) {

				relay := make(chan core.HeuristicInput)
				err := pl.AddEngineRelay(relay)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Successful Run",
			function: "RunPipeline",
			constructionLogic: func() pipeline.Pipeline {

				testO, _ := mocks.NewReader(
					context.Background(),
					core.BlockHeader)

				pl, err := pipeline.NewPath(
					nil,
					core.PathID{},
					[]component.Process{testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl pipeline.Pipeline) {
				assert.Equal(t, pl.State(), pipeline.INACTIVE, "Pipeline should be inactive")

				wg := &sync.WaitGroup{}
				pl.Run(wg)

				assert.Equal(t, pl.State(), pipeline.ACTIVE, "Pipeline should be active")
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			subline := tc.constructionLogic()
			tc.testLogic(t, subline)
		})

	}
}
