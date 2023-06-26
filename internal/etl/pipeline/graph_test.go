package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"

	"github.com/stretchr/testify/assert"
)

var (
	testCUUID1 = core.MakeCUUID(69, 69, 69, 69)
	testCUUID2 = core.MakeCUUID(42, 42, 42, 42)
)

func Test_Graph(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() ComponentGraph
		testLogic         func(*testing.T, ComponentGraph)
	}{
		{
			name:        "Successful Component Node Insertion",
			function:    "AddComponent",
			description: "When a component is added to the graph, it should persist within the graph's edge mapping",

			constructionLogic: NewComponentGraph,
			testLogic: func(t *testing.T, g ComponentGraph) {
				cUUID := core.MakeCUUID(69, 69, 69, 69)

				component, err := mocks.NewDummyPipe(context.Background(), core.GethBlock, core.AccountBalance)
				assert.NoError(t, err)

				err = g.AddComponent(cUUID, component)
				assert.NoError(t, err, "Component addition should resolve to Nil")

				actualComponent, err := g.GetComponent(cUUID)
				assert.NoError(t, err, "Component retrieval should resolve to Nil")

				assert.Equal(t, component, actualComponent)

				edges := g.Edges()

				assert.Contains(t, edges, cUUID)

				assert.Len(t, edges[cUUID], 0, "No edges should exist yet")

			},
		},
		{
			name:        "Failed Cyclic Edge Addition",
			function:    "addEdge",
			description: "When an edge between two components already exists (A->B), then an inverted edge (B->A) should not be possible",

			constructionLogic: func() ComponentGraph {
				g := NewComponentGraph()

				comp1, err := mocks.NewDummyOracle(context.Background(), core.GethBlock)
				if err != nil {
					panic(err)
				}

				if err = g.AddComponent(testCUUID1, comp1); err != nil {
					panic(err)
				}

				comp2, err := mocks.NewDummyPipe(context.Background(), core.GethBlock, core.AccountBalance)
				if err != nil {
					panic(err)
				}

				if err = g.AddComponent(testCUUID2, comp2); err != nil {
					panic(err)
				}

				if err = g.AddEdge(testCUUID1, testCUUID2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g ComponentGraph) {
				err := g.AddEdge(testCUUID2, testCUUID1)
				assert.Error(t, err)

			},
		},
		{
			name:        "Failed Duplicate Edge Addition",
			function:    "AddEdge",
			description: "When a unique edge exists between two components (A->B), a new edge should not be possible",

			constructionLogic: func() ComponentGraph {
				g := NewComponentGraph()

				comp1, err := mocks.NewDummyOracle(context.Background(), core.GethBlock)
				if err != nil {
					panic(err)
				}

				if err = g.AddComponent(testCUUID1, comp1); err != nil {
					panic(err)
				}

				comp2, err := mocks.NewDummyPipe(context.Background(), core.GethBlock, core.AccountBalance)
				if err != nil {
					panic(err)
				}

				if err = g.AddComponent(testCUUID2, comp2); err != nil {
					panic(err)
				}

				if err = g.AddEdge(testCUUID1, testCUUID2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g ComponentGraph) {
				err := g.AddEdge(testCUUID1, testCUUID2)
				assert.Error(t, err)

			},
		},
		{
			name:        "Successful Edge Addition",
			function:    "AddEdge",
			description: "When two components are inserted, an edge should be possible between them",

			constructionLogic: func() ComponentGraph {
				g := NewComponentGraph()

				comp1, err := mocks.NewDummyOracle(context.Background(), core.GethBlock)
				if err != nil {
					panic(err)
				}

				if err = g.AddComponent(testCUUID1, comp1); err != nil {
					panic(err)
				}

				comp2, err := mocks.NewDummyPipe(context.Background(), core.GethBlock, core.AccountBalance)
				if err != nil {
					panic(err)
				}

				if err = g.AddComponent(testCUUID2, comp2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g ComponentGraph) {
				comp1, _ := g.GetComponent(testCUUID1)

				err := g.AddEdge(testCUUID1, testCUUID2)
				assert.NoError(t, err)

				err = comp1.AddEgress(testCUUID2, core.NewTransitChannel())
				assert.Error(t, err, "Error should be returned when trying to add existing outgress of component2 to component1 ingress")

				assert.True(t, g.ComponentExists(testCUUID1))
				assert.True(t, g.ComponentExists(testCUUID2))

				edgeMap := g.Edges()
				assert.Contains(t, edgeMap[testCUUID1], testCUUID2, "ID1 should have a mapped edge to ID2")

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
