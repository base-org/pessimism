package pipeline

import (
	"fmt"
	"log"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

// componentEntry ... Used to store critical component graph entry data
type componentEntry struct {
	comp    component.Component
	edges   map[core.ComponentID]interface{}
	outType core.RegisterType
}

// newEntry ... Intitializer for graph node entry; stores critical routing information
// & component metadata
func newEntry(c component.Component, rt core.RegisterType) *componentEntry {
	return &componentEntry{
		comp:    c,
		outType: rt,
		edges:   make(map[core.ComponentID]interface{}),
	}
}

// cGraph ... Represents a directed acyclic component graph (DAG)
type cGraph struct {
	edgeMap map[core.ComponentID]*componentEntry
}

// newGraph ... Initializer
func newGraph() *cGraph {
	return &cGraph{
		edgeMap: make(map[core.ComponentID]*componentEntry, 0),
	}
}

// componentExists ... Returns true if component node already exists for ID, false otherwise
func (graph *cGraph) componentExists(cID core.ComponentID) bool {
	_, exists := graph.edgeMap[cID]
	return exists
}

// getComponent ... Returns a component entry for some component ID
func (graph *cGraph) getComponent(cID core.ComponentID) (component.Component, error) {
	if graph.componentExists(cID) {
		return graph.edgeMap[cID].comp, nil
	}

	return nil, fmt.Errorf("component with ID %s does not exist within pipeline graph", cID)
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

// addEdge ... Adds edge between two preconstructed constructed component nodes
func (graph *cGraph) addEdge(cID1, cID2 core.ComponentID) error {
	entry1, found := graph.edgeMap[cID1]
	if !found {
		return fmt.Errorf("could not find a valid component in mapping for cID: %s", cID1.String())
	}

	entry2, found := graph.edgeMap[cID2]
	if !found {
		return fmt.Errorf("could not find a valid component in mapping for cID: %s", cID2.String())
	}

	log.Printf("creating entrypoint for %s in component %s", entry1.outType, entry2.comp.ID().String())

	// Edge already exists case edgecase (No pun)
	if _, exists := entry1.edges[entry2.comp.ID()]; exists {
		return fmt.Errorf("edge already exists from (%s) to (%s)", cID1.String(), cID2.String())
	}

	entryChan, err := entry2.comp.GetIngress(entry1.outType)
	if err != nil {
		return err
	}

	log.Printf("adding directive between (%s) -> (%s)", cID1.String(), cID2.String())
	if err := entry1.comp.AddEgress(cID2, entryChan); err != nil {
		return err
	}

	// Update edge mapping with new link
	graph.edgeMap[cID1].edges[cID2] = nil

	return nil
}

// TODO(#23): Manager DAG Component Removal Support
// removeEdge ... Removes an edge from the graph
func (graph *cGraph) removeEdge(_, _ core.ComponentID) error { //nolint:unused // will be implemented soon
	return nil
}

// TODO(#23): Manager DAG Component Removal Support
// removeComponent ... Removes a component from the graph
func (graph *cGraph) removeComponent(_ core.ComponentID) error { //nolint:unused // will be implemented soon
	return nil
}

// addComponent ... Adds component node entry to graph
func (graph *cGraph) addComponent(cID core.ComponentID, comp component.Component) error {
	if _, exists := graph.edgeMap[cID]; exists {
		return fmt.Errorf("component with ID %s already exists", cID)
	}

	graph.edgeMap[cID] = newEntry(comp, comp.OutputType())

	return nil
}

// edges ...  Returns a representation of all graph edges between component IDs
func (graph *cGraph) edges() map[core.ComponentID][]core.ComponentID { //nolint:unused // will be used soon
	idMap := make(map[core.ComponentID][]core.ComponentID, len(graph.edgeMap))

	for cID, cEntry := range graph.edgeMap {
		cEdges := make([]core.ComponentID, len(cEntry.edges))

		i := 0
		for edge := range cEntry.edges {
			cEdges[i] = edge
			i++
		}

		idMap[cID] = cEdges
	}

	return idMap
}
