package core

import (
	"fmt"
)

type ComponentID [4]byte

func MakeComponentID(pt PipelineType, ct ComponentType, rt RegisterType, n Network) ComponentID {
	return [4]byte{
		byte(n),
		byte(pt),
		byte(ct),
		byte(rt),
	}
}

func (cID ComponentID) String() string {
	return fmt.Sprintf("%s:%s:%s:%s",
		Network(cID[0]).String(),
		PipelineType(cID[0]).String(),
		ComponentType(cID[2]).String(),
		RegisterType(cID[3]).String(),
	)
}

func NilCompID() ComponentID {
	return [4]byte{0}
}

type PipelineID [9]byte

func NilPipelineID() PipelineID {
	return [9]byte{0}
}

func MakePipelineID(pt PipelineType, firstCID, lastCID ComponentID) PipelineID {
	return [9]byte{
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
}

func (pID PipelineID) String() string {
	pt := PipelineType(pID[0]).String()
	cID1 := ComponentID(*(*[4]byte)(pID[1:5])).String()
	cID2 := ComponentID(*(*[4]byte)(pID[5:9])).String()

	return fmt.Sprintf("%s::%s::%s",
		pt, cID1, cID2,
	)
}
