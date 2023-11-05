package etl

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type Entry struct {
	id core.PathID
	p  Path
}

// etlStore ... Stores critical path information
//
//	paths - Mapping used for storing all existing paths
//	procToPath - Mapping used for storing all process-->[]PID entries
type EtlStore struct {
	paths      map[core.PathIdentifier][]Entry
	procToPath map[core.ProcessID][]core.PathID
}

// NewEtlStore ... Initializer
func NewEtlStore() EtlStore {
	return EtlStore{
		procToPath: make(map[core.ProcessID][]core.PathID),
		paths:      make(map[core.PathIdentifier][]Entry),
	}
}

/*
Note - PathIDs can only conflict
       when PathType = Live && activityState = Active
*/

// Link ... Creates an entry for some new C_UUID:P_UUID mapping
func (store *EtlStore) Link(id1 core.ProcessID, id2 core.PathID) {
	// EDGE CASE - C_UUID:P_UUID pair already exists
	if _, found := store.procToPath[id1]; !found { // Create slice
		store.procToPath[id1] = make([]core.PathID, 0)
	}

	store.procToPath[id1] = append(store.procToPath[id1], id2)
}

func (store *EtlStore) AddPath(id core.PathID, path Path) {
	entry := Entry{
		id: id,
		p:  path,
	}

	entrySlice, found := store.paths[id.ID]
	if !found {
		entrySlice = make([]Entry, 0)
	}

	entrySlice = append(entrySlice, entry)

	store.paths[id.ID] = entrySlice

	for _, p := range path.Processes() {
		store.Link(p.ID(), id)
	}
}

// GetPathIDs ... Returns all entry PIDs for some CID
func (store *EtlStore) GetPathIDs(cID core.ProcessID) ([]core.PathID, error) {
	pIDs, found := store.procToPath[cID]

	if !found {
		return []core.PathID{}, fmt.Errorf("could not find key for %s", cID)
	}

	return pIDs, nil
}

// getPathByPID ... Returns path store provided some PID
func (store *EtlStore) GetPathByID(id core.PathID) (Path, error) {
	if _, found := store.paths[id.ID]; !found {
		return nil, fmt.Errorf(pIDNotFoundErr, id.String())
	}

	for _, plEntry := range store.paths[id.ID] {
		if plEntry.id.UUID == id.UUID {
			return plEntry.p, nil
		}
	}

	return nil, fmt.Errorf(uuidNotFoundErr)
}

func (store *EtlStore) GetExistingPaths(id core.PathID) []core.PathID {
	entries, exists := store.paths[id.ID]
	if !exists {
		return []core.PathID{}
	}

	PathIDs := make([]core.PathID, len(entries))

	for i, entry := range entries {
		PathIDs[i] = entry.id
	}

	return PathIDs
}

// Count ... Returns the number of active paths
func (store *EtlStore) ActiveCount() int {
	count := 0

	for _, entrySlice := range store.paths {
		for _, entry := range entrySlice {
			if entry.p.State() == ACTIVE {
				count++
			}
		}
	}

	return count
}

func (store *EtlStore) Paths() []Path {
	paths := make([]Path, 0)

	for _, entrySlice := range store.paths {
		for _, entry := range entrySlice {
			paths = append(paths, entry.p)
		}
	}

	return paths
}
