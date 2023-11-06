package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
)

type Store struct {
	ids         map[core.PathID][]core.UUID
	instanceMap map[core.UUID]heuristic.Heuristic // no duplicates
}

// NewStore ... Initializer
func NewStore() *Store {
	return &Store{
		instanceMap: make(map[core.UUID]heuristic.Heuristic),
		ids:         make(map[core.PathID][]core.UUID),
	}
}

func (s *Store) GetHeuristics(ids []core.UUID) ([]heuristic.Heuristic, error) {
	heuristics := make([]heuristic.Heuristic, len(ids))

	for i, id := range ids {
		session, err := s.GetHeuristic(id)
		if err != nil {
			return nil, err
		}

		heuristics[i] = session
	}

	return heuristics, nil
}

func (s *Store) GetHeuristic(id core.UUID) (heuristic.Heuristic, error) {
	if entry, found := s.instanceMap[id]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("heuristic UUID doesn't exists in store heuristic mapping")
}

func (s *Store) GetIDs(id core.PathID) ([]core.UUID, error) {
	if sessionIDs, found := s.ids[id]; found {
		return sessionIDs, nil
	}
	return nil, fmt.Errorf("path UUID doesn't exists in store heuristic mapping")
}

func (s *Store) AddSession(uuid core.UUID,
	id core.PathID, h heuristic.Heuristic) error {
	if _, found := s.instanceMap[uuid]; found {
		return fmt.Errorf("heuristic UUID already exists in store pid mapping")
	}

	if _, found := s.ids[id]; !found {
		s.ids[id] = make([]core.UUID, 0)
	}

	s.instanceMap[uuid] = h
	s.ids[id] = append(s.ids[id], uuid)
	return nil
}

func (s *Store) RemoveInvSession(_ core.UUID,
	_ core.PathID, _ heuristic.Heuristic) error {
	return nil
}
