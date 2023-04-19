package pipeline

import (
	"fmt"
	"log"

	"github.com/base-org/pessimism/internal/core"
)

type pipeLineMap = map[core.PipelinePID][]pipeLineEntry

type pipeLineEntry struct {
	id core.PipelineID
	as ActivityState
	p  PipeLine
}

type pipeRegistry struct {
	pipeLines     pipeLineMap
	compPipelines map[core.ComponentID][]core.PipelineID
}

func newPipeRegistry() *pipeRegistry {
	return &pipeRegistry{
		pipeLines: make(pipeLineMap),
	}
}

/*
Note - PipelineIDs caan only conflict
       when whenpipeLineType = Live && activityState = Active
*/

func (pr *pipeRegistry) addPipeline(id core.PipelineID, pl PipeLine, pType core.PipelineType) {
	entry := pipeLineEntry{
		id: id,
		as: Booting,
		p:  pl,
	}

	if _, found := pr.pipeLines[id.PID]; !found {
		pr.pipeLines[id.PID] = make([]pipeLineEntry, 0)
	}

	pr.pipeLines[id.PID] = append(pr.pipeLines[id.PID], entry)
}

func (pr *pipeRegistry) addCompPipelineLink(cID core.ComponentID, pID core.PipelineID) {
	if _, found := pr.compPipelines[cID]; !found {
		pr.compPipelines[cID] = []core.PipelineID{pID}
		return
	}

	pr.compPipelines[cID] = append(pr.compPipelines[cID], pID)
	return
}

func (pr *pipeRegistry) getPipelineIDs(cID core.ComponentID) ([]core.PipelineID, error) {
	if _, found := pr.compPipelines[cID]; !found {
		return []core.PipelineID{}, fmt.Errorf("Could not find key for %s", cID)
	}

	return pr.compPipelines[cID], nil
}

func (pr *pipeRegistry) fetchCompPipelineIDs(cID core.ComponentID) ([]core.PipelineID, error) {
	if _, found := pr.compPipelines[cID]; !found {
		return []core.PipelineID{}, fmt.Errorf("Could not find key for %s", cID)
	}

	return pr.compPipelines[cID], nil
}

func (pr *pipeRegistry) getPipeline(pID core.PipelineID) (PipeLine, error) {
	if _, found := pr.pipeLines[pID.PID]; !found {
		return nil, fmt.Errorf("Could not find pipeline ID for %s", pID.String())
	}

	log.Printf("Current PID: %s", pID.String())
	for _, plEntry := range pr.pipeLines[pID.PID] {
		log.Printf("entry: %s", plEntry.id.String())
		if plEntry.id.UUID == pID.UUID {
			return plEntry.p, nil
		}
	}

	return nil, fmt.Errorf("Could not find matching UUID for pipeline entry")

}

// removePipeline
