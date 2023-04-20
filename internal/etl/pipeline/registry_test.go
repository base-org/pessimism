package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/stretchr/testify/assert"
)

var (
	nilPID = core.MakePipelineID(0, core.NilCompID(), core.NilCompID())

	cID1 = core.MakeComponentID(0, 0, 0, 0)
	cID2 = core.MakeComponentID(0, 0, 0, 0)
)

func getTestPipeLine(ctx context.Context) PipeLine {

	blackHolePipe, err := registry.NewBlackHoleTxPipe(ctx, component.WithID(cID1))
	if err != nil {
		panic(err)
	}

	contractCreatePipe, err := registry.NewCreateContractTxPipe(ctx, component.WithID(cID2))
	if err != nil {
		panic(err)
	}

	comps := []component.Component{
		blackHolePipe,
		contractCreatePipe,
	}

	pipeLine, err := NewPipeLine(nilPID, comps)
	if err != nil {
		panic(err)
	}

	return pipeLine
}

func Test_PipeRegistry(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() *pipeRegistry
		testLogic         func(*testing.T, *pipeRegistry)
	}{
		{
			name:        "Successful Add When PID Already Exists",
			function:    "addPipeline",
			description: "",

			constructionLogic: func() *pipeRegistry {
				ctx := context.Background()

				testRegistry := newPipeRegistry()
				testPipeLine := getTestPipeLine(ctx)

				testRegistry.addPipeline(nilPID, testPipeLine)
				return testRegistry
			},
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				ctx := context.Background()
				testPipeLine := getTestPipeLine(ctx)

				pID2 := core.MakePipelineID(
					0,
					core.MakeComponentID(0, 0, 0, 1),
					core.MakeComponentID(0, 0, 0, 1),
				)

				pr.addPipeline(pID2, testPipeLine)

				for _, comp := range testPipeLine.Components() {
					pIDs, err := pr.getPipeLineIDs(comp.ID())

					assert.NoError(t, err)
					assert.Len(t, pIDs, 2)
					assert.Equal(t, pIDs[0], nilPID)
					assert.Equal(t, pIDs[1], pID2)
				}

			},
		},
		{
			name:        "Successful Add When PID Does Not Exists",
			function:    "addPipeline",
			description: "",

			constructionLogic: func() *pipeRegistry {
				pr := newPipeRegistry()
				return pr
			},
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				ctx := context.Background()
				testPipeLine := getTestPipeLine(ctx)

				pID := core.MakePipelineID(
					0,
					core.MakeComponentID(0, 0, 0, 1),
					core.MakeComponentID(0, 0, 0, 1),
				)

				pr.addPipeline(pID, testPipeLine)

				for _, comp := range testPipeLine.Components() {
					pIDs, err := pr.getPipeLineIDs(comp.ID())

					assert.NoError(t, err)
					assert.Len(t, pIDs, 1)
					assert.Equal(t, pIDs[0], pID)
				}

			},
		},
		{
			name:        "Successful Retrieval When CID Is Valid Key",
			function:    "getPipeLineIDs",
			description: "",

			constructionLogic: newPipeRegistry,
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				cID := core.MakeComponentID(0, 0, 0, 0)
				pID := core.MakePipelineID(0, cID, cID)

				pr.addComponentLink(cID, pID)

				expectedIDs := []core.PipelineID{pID}
				actualIDs, err := pr.getPipeLineIDs(cID)

				assert.NoError(t, err)
				assert.Equal(t, expectedIDs, actualIDs)

			},
		},
		{
			name:        "Failed Retrieval When CID Is Invalid Key",
			function:    "getPipeLineIDs",
			description: "",

			constructionLogic: newPipeRegistry,
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				cID := core.MakeComponentID(0, 0, 0, 0)

				_, err := pr.getPipeLineIDs(cID)

				assert.Error(t, err)
			},
		},
		{
			name:        "Failed Retrieval When PID Is Invalid Key",
			function:    "getPipeline",
			description: "",

			constructionLogic: newPipeRegistry,
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				cID := core.MakeComponentID(0, 0, 0, 0)
				pID := core.MakePipelineID(0, cID, cID)

				_, err := pr.getPipeline(pID)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), fmt.Sprintf(pIDNotFoundErr, pID))

			},
		}, {
			name:        "Failed Retrieval When Matching UUID Cannot Be Found",
			function:    "getPipeline",
			description: "",

			constructionLogic: func() *pipeRegistry {
				pr := newPipeRegistry()
				return pr
			},
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				cID := core.MakeComponentID(0, 0, 0, 0)
				pID := core.MakePipelineID(0, cID, cID)

				pLine := getTestPipeLine(context.Background())

				pr.addPipeline(pID, pLine)

				pID2 := core.MakePipelineID(0, cID, cID)
				_, err := pr.getPipeline(pID2)

				assert.Error(t, err)
				assert.Equal(t, err.Error(), uuidNotFoundErr)

			},
		}, {
			name:        "Successful Retrieval",
			function:    "getPipeline",
			description: "",

			constructionLogic: func() *pipeRegistry {
				pr := newPipeRegistry()
				return pr
			},
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				cID := core.MakeComponentID(0, 0, 0, 0)
				pID := core.MakePipelineID(0, cID, cID)

				expectedPline := getTestPipeLine(context.Background())

				pr.addPipeline(pID, expectedPline)

				actualPline, err := pr.getPipeline(pID)

				assert.NoError(t, err)
				assert.Equal(t, expectedPline, actualPline)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testRouter := tc.constructionLogic()
			tc.testLogic(t, testRouter)
		})

	}
}
