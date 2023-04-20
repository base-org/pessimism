package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type pipeLineMap = map[core.PipelinePID][]pipeLineEntry

type pipeLineEntry struct {
	id core.PipelineID
	as ActivityState
	p  PipeLine
}

// pipeRegistry ... Stores critical pipeline information
//
//	pipeLines - Mapping used for storing all existing pipelines
//	compPipelines - Mapping used for storing all component-->[]PID entries
type pipeRegistry struct {
	pipeLines     pipeLineMap
	compPipeLines map[core.ComponentID][]core.PipelineID
}

// newPipeRegistry ... Initializer
func newPipeRegistry() *pipeRegistry {
	return &pipeRegistry{
		compPipeLines: make(map[core.ComponentID][]core.PipelineID),
		pipeLines:     make(pipeLineMap),
	}
}

/*
Note - PipelineIDs can only conflict
       when whenpipeLineType = Live && activityState = Active
*/

// addPipeline ... Creates and stores a new pipeline entry
func (pr *pipeRegistry) addPipeline(id core.PipelineID, pl PipeLine) {
	entry := pipeLineEntry{
		id: id,
		as: Booting,
		p:  pl,
	}

	entrySlice, found := pr.pipeLines[id.PID]
	if !found {
		entrySlice = make([]pipeLineEntry, 0)
	}

	entrySlice = append(entrySlice, entry)

	pr.pipeLines[id.PID] = entrySlice

	for _, comp := range pl.Components() {
		pr.addComponentLink(comp.ID(), id)
	}
}

// addComponentLink ... Creates an entry for some new CID:PID mapping
func (pr *pipeRegistry) addComponentLink(cID core.ComponentID, pID core.PipelineID) {
	// EDGE CASE - CID:PID pair already exists
	if _, found := pr.compPipeLines[cID]; !found { // Create slice
		pr.compPipeLines[cID] = make([]core.PipelineID, 0)
	}

	pr.compPipeLines[cID] = append(pr.compPipeLines[cID], pID)
}

// getPipeLineIDs ... Returns all entried PIDs for some CID
func (pr *pipeRegistry) getPipeLineIDs(cID core.ComponentID) ([]core.PipelineID, error) {
	pIDs, found := pr.compPipeLines[cID]

	if !found {
		return []core.PipelineID{}, fmt.Errorf("could not find key for %s", cID)
	}

	return pIDs, nil
}

// getPipelineByPID ... Returns pipeline provided some PID
func (pr *pipeRegistry) getPipeline(pID core.PipelineID) (PipeLine, error) {
	if _, found := pr.pipeLines[pID.PID]; !found {
		return nil, fmt.Errorf(pIDNotFoundErr, pID.String())
	}

	for _, plEntry := range pr.pipeLines[pID.PID] {
		if plEntry.id.UUID == pID.UUID {
			return plEntry.p, nil
		}
	}

	return nil, fmt.Errorf(uuidNotFoundErr)
}
