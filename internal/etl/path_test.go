package etl_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl"
	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	var tests = []struct {
		name     string
		function string

		constructionLogic func() etl.Path
		testLogic         func(t *testing.T, path etl.Path)
	}{
		{
			name:     "Successful Construction",
			function: "NewPath",
			constructionLogic: func() etl.Path {
				sub, _ := mocks.NewSubscriber(
					context.Background(),
					core.BlockHeader,
					core.Log)

				testO, _ := mocks.NewReader(
					context.Background(),
					core.BlockHeader)

				path, err := etl.NewPath(
					nil,
					core.PathID{},
					[]process.Process{sub, testO})

				if err != nil {
					panic(err)
				}

				return path
			},
			testLogic: func(t *testing.T, path etl.Path) {

				assert.Equal(t, path.Processes()[0].EmitType(), core.Log)
				assert.Equal(t, path.Processes()[1].EmitType(), core.BlockHeader)
			},
		},
		{
			name:     "Successful Run",
			function: "AddEngineRelay",
			constructionLogic: func() etl.Path {

				testO, _ := mocks.NewReader(
					context.Background(),
					core.BlockHeader)

				pl, err := etl.NewPath(
					nil,
					core.PathID{},
					[]process.Process{testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl etl.Path) {

				relay := make(chan core.HeuristicInput)
				err := pl.AddEngineRelay(relay)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Successful Run",
			function: "RunPath",
			constructionLogic: func() etl.Path {

				testO, _ := mocks.NewReader(
					context.Background(),
					core.BlockHeader)

				pl, err := etl.NewPath(
					nil,
					core.PathID{},
					[]process.Process{testO})

				if err != nil {
					panic(err)
				}

				return pl
			},
			testLogic: func(t *testing.T, pl etl.Path) {
				assert.Equal(t, pl.State(), etl.INACTIVE, "Path should be inactive")

				wg := &sync.WaitGroup{}
				pl.Run(wg)

				assert.Equal(t, pl.State(), etl.ACTIVE, "Path should be active")
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			path := tc.constructionLogic()
			tc.testLogic(t, path)
		})

	}
}
