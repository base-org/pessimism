package core

import (
	"fmt"

	"github.com/google/uuid"
)

// CPID ... Component Primary ID
type CPID [4]byte

// Represents a non-deterministic ID that's assigned to
// every uniquely invoked component
type ComponentID struct {
	PID  CPID
	UUID uuid.UUID
}

/*
	NOTE: Pipelines that require a backfill will cause inaccurate collisions
	within the pipeline DAG.

*/

// NOTE - This is useful for error handling with functions that
// also return a ComponentID
// NilCompID ... Returns a zero'd out component ID
func NilCompID() ComponentID {
	return ComponentID{
		PID:  CPID{0},
		UUID: [16]byte{0},
	}
}

// MakeComponentID ... Constructs a component ID sequence
// provided all necessary encoding bytes
func MakeComponentID(pt PipelineType, ct ComponentType, rt RegisterType, n Network) ComponentID {
	cID := CPID{
		byte(n),
		byte(pt),
		byte(ct),
		byte(rt),
	}

	return ComponentID{
		PID:  cID,
		UUID: uuid.New(),
	}
}

func (cID CPID) String() string {
	return fmt.Sprintf("%s:%s:%s:%s",
		Network(cID[0]).String(),
		PipelineType(cID[1]).String(),
		ComponentType(cID[2]).String(),
		RegisterType(cID[3]).String(),
	)
}

// String ... Returns string representation of a component ID
func (cID ComponentID) String() string {
	return fmt.Sprintf("%s::%s",
		cID.PID.String(),
		cID.UUID.String(),
	)
}

// Type ... Returns component type from component ID
func (cID ComponentID) Type() ComponentType {
	return ComponentType(cID.PID[2])
}

// Used for local lookups to look for active collisions
type PipelinePID [9]byte

func (pid PipelinePID) String() string {
	pt := PipelineType(pid[0]).String()
	cID1 := CPID(*(*[4]byte)(pid[1:5])).String()
	cID2 := CPID(*(*[4]byte)(pid[5:9])).String()

	return fmt.Sprintf("%s::%s::%s", pt, cID1, cID2)
}

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
	cID1, cID2 := firstCID.PID, lastCID.PID

	pID := PipelinePID{
		byte(pt),
		cID1[0],
		cID1[1],
		cID1[2],
		cID1[3],
		cID2[0],
		cID2[1],
		cID2[2],
		cID2[3],
	}

	return PipelineID{
		PID:  pID,
		UUID: uuid.New(),
	}
}

func (pID PipelineID) String() string {
	return fmt.Sprintf("%s:::%s",
		pID.PID.String(), pID.UUID.String(),
	)
}
