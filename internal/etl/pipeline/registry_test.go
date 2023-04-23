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
	nilPID = core.MakePipelineUUID(0, core.NilComponentUUID(), core.NilComponentUUID())

	cID1 = core.MakeComponentUUID(0, 0, 0, 0)
	cID2 = core.MakeComponentUUID(0, 0, 0, 0)
)

func getTestPipeLine(ctx context.Context) Pipeline {

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

	pipeLine, err := NewPipeLine(core.NilPipelineUUID(), comps)
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

				testRegistry.addPipeline(core.NilPipelineUUID(), testPipeLine)
				return testRegistry
			},
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				ctx := context.Background()
				testPipeLine := getTestPipeLine(ctx)

				pID2 := core.MakePipelineUUID(
					0,
					core.MakeComponentUUID(0, 0, 0, 1),
					core.MakeComponentUUID(0, 0, 0, 1),
				)

				pr.addPipeline(pID2, testPipeLine)

				for _, comp := range testPipeLine.Components() {
					pIDs, err := pr.getPipelineUUIDs(comp.ID())

					assert.NoError(t, err)
					assert.Len(t, pIDs, 2)
					assert.Equal(t, pIDs[0].PID, nilPID.PID)
					assert.Equal(t, pIDs[1].PID, pID2.PID)
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

				pID := core.MakePipelineUUID(
					0,
					core.MakeComponentUUID(0, 0, 0, 1),
					core.MakeComponentUUID(0, 0, 0, 1),
				)

				pr.addPipeline(pID, testPipeLine)

				for _, comp := range testPipeLine.Components() {
					pIDs, err := pr.getPipelineUUIDs(comp.ID())

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
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				pr.addComponentLink(cID, pID)

				expectedIDs := []core.PipelineUUID{pID}
				actualIDs, err := pr.getPipelineUUIDs(cID)

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
				cID := core.MakeComponentUUID(0, 0, 0, 0)

				_, err := pr.getPipelineUUIDs(cID)

				assert.Error(t, err)
			},
		},
		{
			name:        "Failed Retrieval When PID Is Invalid Key",
			function:    "getPipeline",
			description: "",

			constructionLogic: newPipeRegistry,
			testLogic: func(t *testing.T, pr *pipeRegistry) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

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
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				pLine := getTestPipeLine(context.Background())

				pr.addPipeline(pID, pLine)

				pID2 := core.MakePipelineUUID(0, cID, cID)
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
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

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
