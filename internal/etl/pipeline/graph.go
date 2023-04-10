package pipeline

import (
	"fmt"
	"log"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

type componentEntry struct {
	comp    component.Component
	edges   map[core.ComponentID]interface{}
	outType core.RegisterType
}

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

func newGraph() *cGraph {
	return &cGraph{
		edgeMap: make(map[core.ComponentID]*componentEntry, 0),
	}
}

func (graph *cGraph) componentExists(cID core.ComponentID) bool {
	_, exists := graph.edgeMap[cID]
	return exists
}

func (graph *cGraph) getComponent(cID core.ComponentID) (component.Component, error) {
	if graph.componentExists(cID) {
		return graph.edgeMap[cID].comp, nil
	}

	return nil, fmt.Errorf("Component with ID %s does not exist within pipeline graph", cID)
}

// AddEdge ... Adds edge between two preconstructed constructed component nodes
func (graph *cGraph) addEdge(cID1, cID2 core.ComponentID) error {
	entry1, found := graph.edgeMap[cID1]
	if !found {
		return fmt.Errorf("Could not find a valid component in mapping for cID: %s", cID1.String())
	}

	entry2, found := graph.edgeMap[cID2]
	if !found {
		return fmt.Errorf("Could not find a valid component in mapping for cID: %s", cID2.String())
	}

	log.Printf("Creating entrypoint for %s in component %s", entry1.outType, entry2.comp.ID().String())

	// Edge already exists case
	if _, exists := entry1.edges[entry2.comp.ID()]; exists {
		return fmt.Errorf("Edge already exists from (%s) to (%s)", cID1.String(), cID2.String())
	}

	entryChan, err := entry2.comp.GetEntryPoint(entry1.outType)
	if err != nil {
		return err
	}

	log.Printf("Adding directive between (%s) -> (%s)", cID1.String(), cID2.String())
	if err := entry1.comp.AddDirective(cID2, entryChan); err != nil {
		return err
	}

	// Update edge mapping with new link
	graph.edgeMap[cID1].edges[cID2] = nil

	return nil
}

func (graph *cGraph) removeEdge(cID1, cID2 core.ComponentID) error {
	// TODO
	return nil
}

func (graph *cGraph) addComponent(cID core.ComponentID, comp component.Component) error {

	if _, exists := graph.edgeMap[cID]; exists {
		return fmt.Errorf("Component with ID %s already exists", cID)
	}

	graph.edgeMap[cID] = newEntry(comp, comp.OutputType())

	return nil
}

func (graph *cGraph) string() string {
	str := ""

	for key, entry := range graph.edgeMap {
		var slice string = "["

		for edge := range entry.edges {
			slice += edge.String() + ", "
		}

		slice += "]"

		str += fmt.Sprintf("%s -> %s", key.String(), slice)
	}
	return str
}
