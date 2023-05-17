package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// TODO(#48): Pipeline Analysis Functionality
// EtlStore ... Interface used to define all etl storage based functions
type EtlStore interface {
	AddComponentLink(cID core.ComponentUUID, pID core.PipelineUUID)
	AddPipeline(id core.PipelineUUID, pl Pipeline)
	GetAllPipelines() []Pipeline
	GetExistingPipelinesByPID(pPID core.PipelinePID) []core.PipelineUUID
	GetPipelineUUIDs(cID core.ComponentUUID) ([]core.PipelineUUID, error)
	GetPipelineFromPUUID(pUUID core.PipelineUUID) (Pipeline, error)
}

// pipelineEntry ... value entry for some
// pipeline with necessary metadata
type pipelineEntry struct {
	id core.PipelineUUID
	as ActivityState
	p  Pipeline
}

type pipelineMap = map[core.PipelinePID][]pipelineEntry

// etlStore ... Stores critical pipeline information
//
//	pipeLines - Mapping used for storing all existing pipelines
//	compPipelines - Mapping used for storing all component-->[]PID entries
type etlStore struct {
	pipelines     pipelineMap
	compPipelines map[core.ComponentUUID][]core.PipelineUUID
}

// newEtlStore ... Initializer
func newEtlStore() EtlStore {
	return &etlStore{
		compPipeLines: make(map[core.ComponentUUID][]core.PipelineUUID),
		pipeLines:     make(pipeLineMap),
	}
}

/*
Note - PipelineUUIDs can only conflict
       when whenpipeLineType = Live && activityState = Active
*/

// addComponentLink ... Creates an entry for some new C_UUID:P_UUID mapping
func (store *etlStore) AddComponentLink(cUUID core.ComponentUUID, pUUID core.PipelineUUID) {
	// EDGE CASE - C_UUID:P_UUID pair already exists
	if _, found := store.compPipeLines[cUUID]; !found { // Create slice
		store.compPipeLines[cUUID] = make([]core.PipelineUUID, 0)
	}

	store.compPipeLines[cUUID] = append(store.compPipeLines[cUUID], pUUID)
}

// addPipeline ... Creates and stores a new pipeline entry
func (store *etlStore) AddPipeline(pUUID core.PipelineUUID, pl Pipeline) {
	entry := pipeLineEntry{
		id: pUUID,
		as: Booting,
		p:  pl,
	}

	entrySlice, found := store.pipeLines[pUUID.PID]
	if !found {
		entrySlice = make([]pipeLineEntry, 0)
	}

	entrySlice = append(entrySlice, entry)

	store.pipeLines[pUUID.PID] = entrySlice

	for _, comp := range pl.Components() {
		store.AddComponentLink(comp.UUID(), pUUID)
	}
}

// GetPipelineUUIDs ... Returns all entried PIDs for some CID
func (store *etlStore) GetPipelineUUIDs(cID core.ComponentUUID) ([]core.PipelineUUID, error) {
	pIDs, found := store.compPipeLines[cID]

	if !found {
		return []core.PipelineUUID{}, fmt.Errorf("could not find key for %s", cID)
	}

	return pIDs, nil
}

// getPipelineByPID ... Returns pipeline storeovided some PID
func (store *etlStore) GetPipelineFromPUUID(pUUID core.PipelineUUID) (Pipeline, error) {
	if _, found := store.pipeLines[pUUID.PID]; !found {
		return nil, fmt.Errorf(pIDNotFoundErr, pUUID.String())
	}

	for _, plEntry := range store.pipeLines[pUUID.PID] {
		if plEntry.id.UUID == pUUID.UUID {
			return plEntry.p, nil
		}
	}

	return nil, fmt.Errorf(uuidNotFoundErr)
}

// GetExistingPipelinesByPID ... Returns existing pipelines for some PID value
func (store *etlStore) GetExistingPipelinesByPID(pPID core.PipelinePID) []core.PipelineUUID {
	entries, exists := store.pipeLines[pPID]
	if !exists {
		return []core.PipelineUUID{}
	}

	pUUIDs := make([]core.PipelineUUID, len(entries))

	for i, entry := range entries {
		pUUIDs[i] = entry.id
	}

	return pUUIDs
}

// GetAllPipelines ... Returns all existing/current pipelines
func (store *etlStore) GetAllPipelines() []Pipeline {
	pipeLines := make([]Pipeline, 0)

	for _, pLines := range store.pipeLines {
		for _, pipeLine := range pLines {
			pipeLines = append(pipeLines, pipeLine.p)
		}
	}

	return pipeLines
}
