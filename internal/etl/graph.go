package etl

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/process"
)

type node struct {
	p       process.Process
	edges   map[core.ProcessID]interface{}
	outType core.TopicType
}

func newNode(c process.Process, rt core.TopicType) *node {
	return &node{
		p:       c,
		outType: rt,
		edges:   make(map[core.ProcessID]interface{}),
	}
}

func NewGraph() *Graph {
	return &Graph{
		edgeMap: make(map[core.ProcessID]*node),
	}
}

// Represents a directed acyclic process graph (DAG)
type Graph struct {
	edgeMap map[core.ProcessID]*node
}

func (graph *Graph) Exists(id core.ProcessID) bool {
	_, exists := graph.edgeMap[id]
	return exists
}

func (graph *Graph) GetProcess(id core.ProcessID) (process.Process, error) {
	if graph.Exists(id) {
		return graph.edgeMap[id].p, nil
	}

	return nil, fmt.Errorf(procNotFoundErr, id)
}

/*
NOTE - There is no check to ensure that a cyclic edge is being added, meaning
	a caller could create an edge between B->A assuming edge A->B already exists.
	This would contradict the acyclic assumption of a DAG but is fortunately
	circumnavigated since all processes declare entrypoint register dependencies,
	meaning that process could only be susceptible to bipartite connectivity
	in the circumstance where a process declares inverse input->output of an
	existing process.
*/

// Adds subscription or edge between two preconstructed constructed process nodes
func (graph *Graph) Subscribe(from, to core.ProcessID) error {
	fromNode, found := graph.edgeMap[from]
	if !found {
		return fmt.Errorf(procNotFoundErr, from.String())
	}

	toNode, found := graph.edgeMap[to]
	if !found {
		return fmt.Errorf(procNotFoundErr, to.String())
	}

	if _, exists := fromNode.edges[toNode.p.ID()]; exists {
		return fmt.Errorf(edgeExistsErr, from.String(), to.String())
	}

	relay, err := toNode.p.GetRelay(fromNode.outType)
	if err != nil {
		return err
	}

	if err := fromNode.p.AddSubscriber(to, relay); err != nil {
		return err
	}

	// Update edge mapping with new link
	graph.edgeMap[from].edges[to] = nil

	return nil
}

// TODO(#23): Manager DAG process Removal Support
// removeEdge ... Removes an edge from the graph
func (graph *Graph) RemoveEdge(_, _ core.ProcessID) error {
	return nil
}

// TODO(#23): Manager DAG process Removal Support
// Remove ... Removes a process from the graph
func (graph *Graph) Remove(_ core.ProcessID) error {
	return nil
}

func (graph *Graph) Add(id core.ProcessID, p process.Process) error {
	if _, exists := graph.edgeMap[id]; exists {
		return fmt.Errorf(procExistsErr, id)
	}

	graph.edgeMap[id] = newNode(p, p.EmitType())

	return nil
}

func (graph *Graph) AddMany(processes []process.Process) error {
	// Add all process entries to graph
	for _, p := range processes {
		if err := graph.Add(p.ID(), p); err != nil {
			return err
		}
	}

	// Add edges between processes
	for i := 1; i < len(processes); i++ {
		err := graph.Subscribe(processes[i].ID(), processes[i-1].ID())
		if err != nil {
			return err
		}
	}
	return nil
}

func (graph *Graph) Edges() map[core.ProcessID][]core.ProcessID {
	uuidMap := make(map[core.ProcessID][]core.ProcessID, len(graph.edgeMap))

	for id, cEntry := range graph.edgeMap {
		cEdges := make([]core.ProcessID, len(cEntry.edges))

		i := 0
		for edge := range cEntry.edges {
			cEdges[i] = edge
			i++
		}

		uuidMap[id] = cEdges
	}

	return uuidMap
}
