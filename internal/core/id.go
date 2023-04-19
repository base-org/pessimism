package core

import (
	"fmt"

	"github.com/google/uuid"
)

// ComponentID ... Represents a deterministic ID that's assigned
// to all ETL components
type ComponentID [4]byte

/*
	NOTE: Pipelines that require a backfill will cause inaccurate collisions
	within the pipeline DAG.

*/

// NOTE - This is useful for error handling with functions that
// also return a ComponentID
// NilCompID ... Returns a zero'd out component ID
func NilCompID() ComponentID {
	return ComponentID{0}
}

// MakeComponentID ... Constructs a component ID sequence
// provided all necessary encoding bytes
func MakeComponentID(pt PipelineType, ct ComponentType, rt RegisterType, n Network) ComponentID {
	return ComponentID{
		byte(n),
		byte(pt),
		byte(ct),
		byte(rt),
	}
}

// String ... Returns string representation of a component ID
func (cID ComponentID) String() string {
	return fmt.Sprintf("%s:%s:%s:%s",
		Network(cID[0]).String(),
		PipelineType(cID[1]).String(),
		ComponentType(cID[2]).String(),
		RegisterType(cID[3]).String(),
	)
}

// Type ... Returns component type from component ID
func (cID ComponentID) Type() ComponentType {
	return ComponentType(cID[2])
}

// Used for local lookups to look for active collisions
type PipelinePID = [9]byte

type PipelineID struct {
	PID  PipelinePID
	UUID uuid.UUID
}

func NilPipelineID() PipelineID {
	return PipelineID{
		PID:  PipelinePID{0},
		UUID: [16]byte{0},
	}
}

func MakePipelineID(pt PipelineType, firstCID, lastCID ComponentID) PipelineID {
	pID := PipelinePID{
		byte(pt),
		firstCID[0],
		firstCID[1],
		firstCID[2],
		firstCID[3],
		lastCID[0],
		lastCID[1],
		lastCID[2],
		lastCID[3],
	}

	return PipelineID{
		PID:  pID,
		UUID: uuid.New(),
	}
}

func (pID PipelineID) String() string {
	pid := pID.PID

	pt := PipelineType(pid[0]).String()
	cID1 := ComponentID(*(*[4]byte)(pid[1:5])).String()
	cID2 := ComponentID(*(*[4]byte)(pid[5:9])).String()

	return fmt.Sprintf("%s::%s::%s:::%s",
		pt, cID1, cID2, pID.UUID.String(),
	)
}
