package engine_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/stretchr/testify/assert"
)

func TestSessionStore(t *testing.T) {
	id1 := core.NewUUID()
	id2 := core.NewUUID()

	var tests = []struct {
		name        string
		function    string
		constructor func() *engine.Store
		testFunc    func(t *testing.T, ss *engine.Store)
	}{
		{
			name: "Successful Retrieval",
			constructor: func() *engine.Store {
				ss := engine.NewStore()

				h := heuristic.New(core.TopicType(0), core.BalanceEnforcement)
				h.SetID(id1)

				_ = ss.AddSession(id1, core.PathID{}, h)

				return ss
			},
			testFunc: func(t *testing.T, ss *engine.Store) {
				// Ensure that the heuristic is retrievable
				h, err := ss.GetHeuristic(id1)
				assert.NoError(t, err)
				assert.Equal(t, h.ID(), id1)

				// Ensure that path UUIDs are retrievable
				ids, err := ss.GetIDs(core.PathID{})
				assert.NoError(t, err)
				assert.Equal(t, ids, []core.UUID{id1})
			},
		},
		{
			name: "Successful Retrieval with Multiple Heuristics",
			constructor: func() *engine.Store {
				ss := engine.NewStore()

				h := heuristic.New(core.TopicType(0), core.BalanceEnforcement)
				h.SetID(id1)

				_ = ss.AddSession(id1, core.PathID{}, h)

				h2 := heuristic.New(core.TopicType(0), core.BalanceEnforcement)
				h2.SetID(id2)

				_ = ss.AddSession(id2, core.PathID{}, h2)

				return ss
			},
			testFunc: func(t *testing.T, ss *engine.Store) {
				// Ensure that the first inserted heuristic is retrievable
				h, err := ss.GetHeuristic(id1)
				assert.NoError(t, err)
				assert.Equal(t, h.ID(), id1)

				// Ensure that the second inserted heuristic is retrievable
				h2, err := ss.GetHeuristic(id2)
				assert.NoError(t, err)
				assert.Equal(t, h2.ID(), id2)

				// Ensure that path UUIDs are retrievable
				ids, err := ss.GetIDs(core.PathID{})
				assert.NoError(t, err)
				assert.Equal(t, ids, []core.UUID{id1, id2})

				// Ensure that both heuristics are retrievable at once
				hs, err := ss.GetHeuristics([]core.UUID{id1, id2})
				assert.NoError(t, err)
				assert.Equal(t, hs, []heuristic.Heuristic{h, h2})
			},
		},
		{
			name: "Successful Retrieval",
			constructor: func() *engine.Store {
				ss := engine.NewStore()

				h := heuristic.New(core.TopicType(0), core.BalanceEnforcement)
				h.SetID(id1)

				_ = ss.AddSession(id1, core.PathID{}, h)
				return ss
			},
			testFunc: func(t *testing.T, ss *engine.Store) {
				// Ensure that the heuristic is retrievable
				h, err := ss.GetHeuristic(id1)
				assert.NoError(t, err)
				assert.Equal(t, h.ID(), id1)

				// Ensure that path UUIDs are retrievable
				ids, err := ss.GetIDs(core.PathID{})
				assert.NoError(t, err)
				assert.Equal(t, ids, []core.UUID{id1})
			},
		},
		{
			name: "Failed Retrieval",
			constructor: func() *engine.Store {
				ss := engine.NewStore()

				h := heuristic.New(core.TopicType(0), core.BalanceEnforcement)
				_ = ss.AddSession(id1, core.PathID{}, h)

				return ss
			},
			testFunc: func(t *testing.T, ss *engine.Store) {
				h, err := ss.GetHeuristic(id2)
				assert.Nil(t, h)
				assert.Error(t, err)
			},
		},
		{
			name: "Failed Add with Duplicate IDs",
			constructor: func() *engine.Store {
				ss := engine.NewStore()

				h := heuristic.New(core.TopicType(0), core.BalanceEnforcement)
				_ = ss.AddSession(id1, core.PathID{}, h)

				return ss
			},
			testFunc: func(t *testing.T, ss *engine.Store) {
				// Ensure that only one uuid can exist in the store
				err := ss.AddSession(id1, core.PathID{}, heuristic.New(core.TopicType(0), core.BalanceEnforcement))
				assert.Error(t, err)
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, test.name), func(t *testing.T) {
			ss := test.constructor()
			test.testFunc(t, ss)
		})
	}
}
