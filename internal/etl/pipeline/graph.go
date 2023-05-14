package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// ComponentGraph ...
type ComponentGraph interface {
	ComponentExists(cUIID core.ComponentUUID) bool
	GetComponent(cUIID core.ComponentUUID) (component.Component, error)
	AddEdge(cUUID1, cUUID2 core.ComponentUUID) error
	AddComponent(cUIID core.ComponentUUID, comp component.Component) error
	AddComponents(cSlice []component.Component) error

	Edges() map[core.ComponentUUID][]core.ComponentUUID // Useful for testing

	// TODO(#23): Manager DAG Component Removal Support
	RemoveEdge(_, _ core.ComponentUUID) error
	RemoveComponent(_ core.ComponentUUID) error
}

// cNode ... Used to store critical component graph entry data
type cNode struct {
	comp    component.Component
	edges   map[core.ComponentUUID]interface{}
	outType core.RegisterType
}

// newNode ... Intitializer for graph node entry; stores critical routing information
// & component metadata
func newNode(c component.Component, rt core.RegisterType) *cNode {
	return &cNode{
		comp:    c,
		outType: rt,
		edges:   make(map[core.ComponentUUID]interface{}),
	}
}

// cGraph ... Represents a directed acyclic component graph (DAG)
type cGraph struct {
	edgeMap map[core.ComponentUUID]*cNode
}

// NewComponentGraph ... Initializer
func NewComponentGraph() ComponentGraph {
	return &cGraph{
		edgeMap: make(map[core.ComponentUUID]*cNode, 0),
	}
}

// componentExists ... Returns true if component node already exists for UUID, false otherwise
func (graph *cGraph) ComponentExists(cUIID core.ComponentUUID) bool {
	_, exists := graph.edgeMap[cUIID]
	return exists
}

// getComponent ... Returns a component entry for some component ID
func (graph *cGraph) GetComponent(cUIID core.ComponentUUID) (component.Component, error) {
	if graph.ComponentExists(cUIID) {
		return graph.edgeMap[cUIID].comp, nil
	}

	return nil, fmt.Errorf(cUUIDNotFoundErr, cUIID)
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
			zap.String("from", entry1.comp.UUID().String()),
			zap.String("to", entry2.comp.UUID().String()))

	// Edge already exists edgecase (No pun)
	if _, exists := entry1.edges[entry2.comp.UUID()]; exists {
		return fmt.Errorf(edgeExistsErr, cUUID1.String(), cUUID2.String())
	}

	c2Ingress, err := entry2.comp.GetIngress(entry1.outType)
	if err != nil {
		return err
	}

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

// addComponent ... Adds component node entry to edge mapping
func (graph *cGraph) AddComponent(cUIID core.ComponentUUID, comp component.Component) error {
	if _, exists := graph.edgeMap[cUIID]; exists {
		return fmt.Errorf(cUUIDExistsErr, cUIID)
	}

	graph.edgeMap[cUIID] = newNode(comp, comp.OutputType())

	return nil
}

// AddComponents ... Inserts all components from some slice into edge mapping
func (graph *cGraph) AddComponents(cSlice []component.Component) error {
	for _, c := range cSlice {
		if err := graph.AddComponent(c.UUID(), c); err != nil {
			return err
		}
	}
	return nil
}

// Edges ...  Returns a representation of all graph edges between component UUIDs
func (graph *cGraph) Edges() map[core.ComponentUUID][]core.ComponentUUID {
	uuidMap := make(map[core.ComponentUUID][]core.ComponentUUID, len(graph.edgeMap))

	for cUIID, cEntry := range graph.edgeMap {
		cEdges := make([]core.ComponentUUID, len(cEntry.edges))

		i := 0
		for edge := range cEntry.edges {
			cEdges[i] = edge
			i++
		}

		uuidMap[cUIID] = cEdges
	}

	return uuidMap
}
