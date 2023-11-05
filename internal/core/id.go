package core

import (
	"fmt"

	"github.com/google/uuid"
)

// UUID ... third-party wrapper struct for
// https://pkg.go.dev/github.com/google/UUID
type UUID struct {
	uuid.UUID
}

func NewUUID() UUID {
	return UUID{
		uuid.New(),
	}
}

// ShortString ... Short string representation for easier
// debugging and ensuring conformance with pessimism specific abstractions
// https://pkg.go.dev/github.com/google/UUID#UUID.String
func (id UUID) ShortString() string {
	uid := id.UUID
	// Only render first 8 bytes instead of entire sequence
	return fmt.Sprintf("%d%d%d%d%d%d%d%d%d",
		uid[0],
		uid[1],
		uid[2],
		uid[2],
		uid[3],
		uid[4],
		uid[5],
		uid[6],
		uid[7])
}

type ProcIdentifier [4]byte

// Represents a non-deterministic ID that's assigned to
// every uniquely constructed ETL process
type ProcessID struct {
	ID   ProcIdentifier
	UUID UUID
}

// Used for local lookups to look for active collisions
type PathIdentifier [9]byte

// Represents a non-deterministic ID that's assigned to
// every uniquely constructed ETL path
type PathID struct {
	ID   PathIdentifier
	UUID UUID
}

func (id PathID) Equal(other PathID) bool {
	return id.ID == other.ID
}

func (id PathID) PathType() PathType {
	return PathType(id.ID[0])
}

func (id PathID) NetworkType() Network {
	return Network(id.ID[1])
}

// MakeProcessID ...
func MakeProcessID(pt PathType, ct ProcessType, tt TopicType, n Network) ProcessID {
	cID := ProcIdentifier{
		byte(n),
		byte(pt),
		byte(ct),
		byte(tt),
	}

	return ProcessID{
		ID:   cID,
		UUID: NewUUID(),
	}
}

// MakePathID ... Constructs a path PID sequence & random UUID
func MakePathID(pt PathType, proc1, proc2 ProcessID) PathID {
	id1, id2 := proc1.ID, proc2.ID

	id := PathIdentifier{
		byte(pt),
		id1[0],
		id1[1],
		id1[2],
		id1[3],
		id2[0],
		id2[1],
		id2[2],
		id2[3],
	}

	return PathID{
		ID:   id,
		UUID: NewUUID(),
	}
}

// String ... Returns string representation of a process PID
func (pid ProcIdentifier) String() string {
	return fmt.Sprintf("%s:%s:%s:%s",
		Network(pid[0]).String(),
		PathType(pid[1]).String(),
		ProcessType(pid[2]).String(),
		TopicType(pid[3]).String(),
	)
}

func (id ProcessID) String() string {
	return fmt.Sprintf("%s",
		id.UUID.ShortString(),
	)
}

func (id ProcessID) Identifier() string {
	return fmt.Sprintf("%s",
		id.ID.String(),
	)
}
func (id ProcessID) Type() ProcessType {
	return ProcessType(id.ID[2])
}

func (id PathIdentifier) String() string {
	pt := PathType(id[0]).String()
	first := ProcIdentifier(*(*[4]byte)(id[1:5])).String()
	last := ProcIdentifier(*(*[4]byte)(id[5:9])).String()

	return fmt.Sprintf("%s::%s::%s", pt, first, last)
}

func (id PathID) Network() Network {
	return Network(id.ID[1])
}

func (id PathID) String() string {
	return id.UUID.ShortString()
}

func (id PathID) Identifier() string {
	return id.ID.String()
}

type SessionID struct {
	PathID      PathID
	ProcID      ProcessID
	HeuristicID UUID
}
