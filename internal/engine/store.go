package engine

import (
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

type InvariantStore struct {
	s map[core.RegisterType][]invariant.Invariant
}

func NewInvariantStore() *InvariantStore {
	return &InvariantStore{
		s: make(map[core.RegisterType][]invariant.Invariant),
	}
}

func (is *InvariantStore) GetInvariants(rk core.RegisterType) ([]invariant.Invariant, error) {
	return is.s[rk], nil
}

func (is *InvariantStore) AddInvariant(rk core.RegisterType, inv invariant.Invariant) error {
	if _, found := is.s[rk]; !found {
		is.s[rk] = make([]invariant.Invariant, 0)
	}

	is.s[rk] = append(is.s[rk], inv)
	return nil
}
