package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// pipeLineEntry ... value entry for some
// pipeline with necessary metadata
type pipeLineEntry struct {
	id core.PipelineUUID
	as ActivityState
	p  Pipeline
}

type pipeLineMap = map[core.PipelinePID][]pipeLineEntry

// pipeRegistry ... Stores critical pipeline information
//
//	pipeLines - Mapping used for storing all existing pipelines
//	compPipelines - Mapping used for storing all component-->[]PID entries
type pipeRegistry struct {
	pipeLines     pipeLineMap
	compPipeLines map[core.ComponentUUID][]core.PipelineUUID
}

// newPipeRegistry ... Initializer
func newPipeRegistry() *pipeRegistry {
	return &pipeRegistry{
		compPipeLines: make(map[core.ComponentUUID][]core.PipelineUUID),
		pipeLines:     make(pipeLineMap),
	}
}

/*
Note - PipelineUUIDs can only conflict
       when whenpipeLineType = Live && activityState = Active
*/

// addPipeline ... Creates and stores a new pipeline entry
func (pr *pipeRegistry) addPipeline(id core.PipelineUUID, pl Pipeline) {
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

// addComponentLink ... Creates an entry for some new C_UUID:P_UUID mapping
func (pr *pipeRegistry) addComponentLink(cID core.ComponentUUID, pID core.PipelineUUID) {
	// EDGE CASE - C_UUID:P_UUID pair already exists
	if _, found := pr.compPipeLines[cID]; !found { // Create slice
		pr.compPipeLines[cID] = make([]core.PipelineUUID, 0)
	}

	pr.compPipeLines[cID] = append(pr.compPipeLines[cID], pID)
}

// getPipelineUUIDs ... Returns all entried PIDs for some CID
func (pr *pipeRegistry) getPipelineUUIDs(cID core.ComponentUUID) ([]core.PipelineUUID, error) {
	pIDs, found := pr.compPipeLines[cID]

	if !found {
		return []core.PipelineUUID{}, fmt.Errorf("could not find key for %s", cID)
	}

	return pIDs, nil
}

// getPipelineByPID ... Returns pipeline provided some PID
func (pr *pipeRegistry) getPipeline(pID core.PipelineUUID) (Pipeline, error) {
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
