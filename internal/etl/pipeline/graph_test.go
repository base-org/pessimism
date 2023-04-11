package pipeline

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/stretchr/testify/assert"
)

var (
	testID1 = core.MakeComponentID(69, 69, 69, 69)
	testID2 = core.MakeComponentID(42, 42, 42, 42)
)

func Test_Graph(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() *cGraph
		testLogic         func(*testing.T, *cGraph)
	}{
		{
			name:        "Successful Component Node Insertion",
			function:    "addComponent",
			description: "When a component is added to the graph, it should persist as an elongated value within the graph's internal edge mapping",

			constructionLogic: newGraph,
			testLogic: func(t *testing.T, g *cGraph) {
				testID := core.MakeComponentID(69, 69, 69, 69)

				component, err := registry.NewBlackHoleTxPipe(context.Background(), component.WithID(testID))
				assert.NoError(t, err)

				err = g.addComponent(testID, component)
				assert.NoError(t, err, "Component addition should resolve to Nil")

				nodeEntry := g.edgeMap[testID]
				assert.Equal(t, nodeEntry.edges, make(map[core.ComponentID]interface{}, 0), "No edges should exist for component entry")
				assert.Equal(t, nodeEntry.outType, component.OutputType(), "Output types should match")
				assert.Equal(t, nodeEntry.comp.ID(), testID, "IDs should match for component entry")
			},
		},
		{
			name:        "Failed Cyclic Edge Addition",
			function:    "addEdge",
			description: "When an edge between two components already exists (A->B), then an inversed edge (B->A) should not be possible",

			constructionLogic: func() *cGraph {
				g := newGraph()

				comp1, err := registry.NewMockOracle(context.Background(), core.GethBlock)
				if err != nil {
					panic(err)
				}

				if err = g.addComponent(testID1, comp1); err != nil {
					panic(err)
				}

				comp2, err := registry.NewCreateContractTxPipe(context.Background(), component.WithID(testID1))
				if err != nil {
					panic(err)
				}

				if err = g.addComponent(testID2, comp2); err != nil {
					panic(err)
				}

				if err = g.addEdge(testID1, testID2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g *cGraph) {
				err := g.addEdge(testID2, testID1)
				assert.Error(t, err)

			},
		},
		{
			name:        "Failed Duplicate Edge Addition",
			function:    "addEdge",
			description: "When a unique edge exists between two components (A->B), a new edge should not be possible",

			constructionLogic: func() *cGraph {
				g := newGraph()

				comp1, err := registry.NewMockOracle(context.Background(), core.GethBlock)
				if err != nil {
					panic(err)
				}

				if err = g.addComponent(testID1, comp1); err != nil {
					panic(err)
				}

				comp2, err := registry.NewCreateContractTxPipe(context.Background(), component.WithID(testID1))
				if err != nil {
					panic(err)
				}

				if err = g.addComponent(testID2, comp2); err != nil {
					panic(err)
				}

				if err = g.addEdge(testID1, testID2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g *cGraph) {
				err := g.addEdge(testID1, testID2)
				assert.Error(t, err)

			},
		},
		{
			name:        "Successful Edge Addition",
			function:    "addEdge",
			description: "When two components are inserted, an edge should be possible between them",

			constructionLogic: func() *cGraph {
				g := newGraph()

				comp1, err := registry.NewMockOracle(context.Background(), core.GethBlock)
				if err != nil {
					panic(err)
				}

				if err = g.addComponent(testID1, comp1); err != nil {
					panic(err)
				}

				comp2, err := registry.NewCreateContractTxPipe(context.Background(), component.WithID(testID1))
				if err != nil {
					panic(err)
				}

				if err = g.addComponent(testID2, comp2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g *cGraph) {
				comp1, _ := g.getComponent(testID1)

				err := g.addEdge(testID1, testID2)
				assert.NoError(t, err)

				err = comp1.AddDirective(testID2, core.NewTransitChannel())
				assert.Error(t, err, "Error should be returned when trying to add existing directive of component2 to component1")

				assert.True(t, strings.Contains(err.Error(), "directive key already exists within component router mapping"))
				assert.True(t, g.componentExists(testID1))
				assert.True(t, g.componentExists(testID2))

				_, exists := g.edgeMap[testID1].edges[testID2]
				assert.True(t, exists)

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
