package etl_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"

	"github.com/base-org/pessimism/internal/etl"
	"github.com/base-org/pessimism/internal/mocks"

	"github.com/stretchr/testify/assert"
)

var (
	id1 = core.MakeProcessID(69, 69, 69, 69)
	id2 = core.MakeProcessID(42, 42, 42, 42)
)

func Test_Graph(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() *etl.Graph
		testLogic         func(*testing.T, *etl.Graph)
	}{
		{
			name:        "Successful Process Node Insertion",
			function:    "AddProcess",
			description: "When a process is added to the graph, it should persist within the graph's edge mapping",

			constructionLogic: etl.NewGraph,
			testLogic: func(t *testing.T, g *etl.Graph) {
				cUUID := core.MakeProcessID(69, 69, 69, 69)

				process, err := mocks.NewSubscriber(context.Background(), core.BlockHeader, core.BlockHeader)
				assert.NoError(t, err)

				err = g.Add(cUUID, process)
				assert.NoError(t, err, "Process addition should resolve to Nil")

				actualProcess, err := g.GetProcess(cUUID)
				assert.NoError(t, err, "Process retrieval should resolve to Nil")

				assert.Equal(t, process, actualProcess)

				edges := g.Edges()

				assert.Contains(t, edges, cUUID)

				assert.Len(t, edges[cUUID], 0, "No edges should exist yet")

			},
		},
		{
			name:        "Failed Cyclic Edge Addition",
			function:    "addEdge",
			description: "When an edge between two processes already exists (A->B), then an inverted edge (B->A) should not be possible",

			constructionLogic: func() *etl.Graph {
				g := etl.NewGraph()

				comp1, err := mocks.NewReader(context.Background(), core.BlockHeader)
				if err != nil {
					panic(err)
				}

				if err = g.Add(id1, comp1); err != nil {
					panic(err)
				}

				comp2, err := mocks.NewSubscriber(context.Background(), core.BlockHeader, core.BlockHeader)
				if err != nil {
					panic(err)
				}

				if err = g.Add(id2, comp2); err != nil {
					panic(err)
				}

				if err = g.Subscribe(id1, id2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g *etl.Graph) {
				err := g.Subscribe(id2, id1)
				assert.Error(t, err)

			},
		},
		{
			name:        "Failed Duplicate Edge Addition",
			function:    "AddEdge",
			description: "When a unique edge exists between two processes (A->B), a new edge should not be possible",

			constructionLogic: func() *etl.Graph {
				g := etl.NewGraph()

				comp1, err := mocks.NewReader(context.Background(), core.BlockHeader)
				if err != nil {
					panic(err)
				}

				if err = g.Add(id1, comp1); err != nil {
					panic(err)
				}

				comp2, err := mocks.NewSubscriber(context.Background(), core.BlockHeader, core.BlockHeader)
				if err != nil {
					panic(err)
				}

				if err = g.Add(id2, comp2); err != nil {
					panic(err)
				}

				if err = g.Subscribe(id1, id2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g *etl.Graph) {
				err := g.Subscribe(id1, id2)
				assert.Error(t, err)

			},
		},
		{
			name:        "Successful Edge Addition",
			function:    "AddEdge",
			description: "When two processes are inserted, an edge should be possible between them",

			constructionLogic: func() *etl.Graph {
				g := etl.NewGraph()

				comp1, err := mocks.NewReader(context.Background(), core.BlockHeader)
				if err != nil {
					panic(err)
				}

				if err = g.Add(id1, comp1); err != nil {
					panic(err)
				}

				comp2, err := mocks.NewSubscriber(context.Background(), core.BlockHeader, core.BlockHeader)
				if err != nil {
					panic(err)
				}

				if err = g.Add(id2, comp2); err != nil {
					panic(err)
				}

				return g
			},

			testLogic: func(t *testing.T, g *etl.Graph) {
				comp1, _ := g.GetProcess(id1)

				err := g.Subscribe(id1, id2)
				assert.NoError(t, err)

				err = comp1.AddSubscriber(id2, core.NewTransitChannel())
				assert.Error(t, err)

				assert.True(t, g.Exists(id1))
				assert.True(t, g.Exists(id2))

				edgeMap := g.Edges()
				assert.Contains(t, edgeMap[id1], id2, "ID1 should have a mapped edge to ID2")

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
