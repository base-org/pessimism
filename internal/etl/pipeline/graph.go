package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type ComponentGraph interface {
	ComponentExists(cID core.ComponentUUID) bool
	GetComponent(cID core.ComponentUUID) (component.Component, error)
	AddEdge(cUUID1, cUUID2 core.ComponentUUID) error
	AddComponent(cID core.ComponentUUID, comp component.Component) error
	AddComponents(cSlice []component.Component) error

	// TODO(#23): Manager DAG Component Removal Support
	RemoveEdge(_, _ core.ComponentUUID) error
	RemoveComponent(_ core.ComponentUUID) error
}

// componentEntry ... Used to store critical component graph entry data
type componentEntry struct {
	comp    component.Component
	edges   map[core.ComponentUUID]interface{}
	outType core.RegisterType
}

// newEntry ... Intitializer for graph node entry; stores critical routing information
// & component metadata
func newEntry(c component.Component, rt core.RegisterType) *componentEntry {
	return &componentEntry{
		comp:    c,
		outType: rt,
		edges:   make(map[core.ComponentUUID]interface{}),
	}
}

// cGraph ... Represents a directed acyclic component graph (DAG)
type cGraph struct {
	edgeMap map[core.ComponentUUID]*componentEntry
}

// newGraph ... Initializer
func newGraph() *cGraph {
	return &cGraph{
		edgeMap: make(map[core.ComponentUUID]*componentEntry, 0),
	}
}

// componentExists ... Returns true if component node already exists for UUID, false otherwise
func (graph *cGraph) ComponentExists(cID core.ComponentUUID) bool {
	_, exists := graph.edgeMap[cID]
	return exists
}

// getComponent ... Returns a component entry for some component ID
func (graph *cGraph) GetComponent(cID core.ComponentUUID) (component.Component, error) {
	if graph.ComponentExists(cID) {
		return graph.edgeMap[cID].comp, nil
	}

	return nil, fmt.Errorf(cUUIDNotFoundErr, cID)
}

/*
NOTE - There is no check to ensure that a cyclic edge is being added, meaning
	a caller could create an edge between B->A assuming edge A->B already exists.
	This would contradict the acyclic assumption of a DAG but is fortunatetly
	cirumnavigated since all components declare entrypoint register dependencies,
	meaning that component could only be susceptible to bipartite connectivity
	in the circumstance where a component declares inversal input->output of an
	existing component.
*/

// TODO(#30): Pipeline Collisions Occur When They Shouldn't
// addEdge ... Adds edge between two preconstructed constructed component nodes
func (graph *cGraph) AddEdge(cUUID1, cUUID2 core.ComponentUUID) error {
	entry1, found := graph.edgeMap[cUUID1]
	if !found {
		return fmt.Errorf(cUUIDNotFoundErr, cUUID1.String())
	}

	entry2, found := graph.edgeMap[cUUID2]
	if !found {
		return fmt.Errorf(cUUIDNotFoundErr, cUUID2.String())
	}

	logging.NoContext().
		Debug("Adding edge between components",
			zap.String("cuuid_1", entry1.comp.ID().String()),
			zap.String("cuuid_2", entry2.comp.ID().String()))

	// Edge already exists edgecase (No pun)
	if _, exists := entry1.edges[entry2.comp.ID()]; exists {
		return fmt.Errorf(edgeExistsErr, cUUID1.String(), cUUID2.String())
	}

	c2Ingress, err := entry2.comp.GetIngress(entry1.outType)
	if err != nil {
		return err
	}

	logging.NoContext().
		Debug("Adding edge", zap.String("From", cUUID1.String()), zap.String("To", cUUID2.String()))

	if err := entry1.comp.AddEgress(cUUID2, c2Ingress); err != nil {
		return err
	}

	// Update edge mapping with new link
	graph.edgeMap[cUUID1].edges[cUUID2] = nil

	return nil
}

// TODO(#23): Manager DAG Component Removal Support
// removeEdge ... Removes an edge from the graph
func (graph *cGraph) RemoveEdge(_, _ core.ComponentUUID) error {
	return nil
}

// TODO(#23): Manager DAG Component Removal Support
// removeComponent ... Removes a component from the graph
func (graph *cGraph) RemoveComponent(_ core.ComponentUUID) error {
	return nil
}

// addComponent ... Adds component node entry to graph
func (graph *cGraph) AddComponent(cID core.ComponentUUID, comp component.Component) error {
	if _, exists := graph.edgeMap[cID]; exists {
		return fmt.Errorf(cUUIDExistsErr, cID)
	}

	graph.edgeMap[cID] = newEntry(comp, comp.OutputType())

	return nil
}

func (graph *cGraph) AddComponents(cSlice []component.Component) error {
	for _, c := range cSlice {
		if err := graph.AddComponent(c.ID(), c); err != nil {
			return err
		}
	}
	return nil
}

// edges ...  Returns a representation of all graph edges between component IDs
func (graph *cGraph) edges() map[core.ComponentUUID][]core.ComponentUUID { //nolint:unused // will be leveraged soon
	idMap := make(map[core.ComponentUUID][]core.ComponentUUID, len(graph.edgeMap))

	for cID, cEntry := range graph.edgeMap {
		cEdges := make([]core.ComponentUUID, len(cEntry.edges))

		i := 0
		for edge := range cEntry.edges {
			cEdges[i] = edge
			i++
		}

		idMap[cID] = cEdges
	}

	return idMap
}
