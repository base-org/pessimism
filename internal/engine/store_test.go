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
	sUUID1 := core.MakeSUUID(core.Layer1, core.Live, core.HeuristicType(0))
	sUUID2 := core.MakeSUUID(core.Layer2, core.Live, core.HeuristicType(0))
	pUUID1 := core.NilPUUID()

	var tests = []struct {
		name        string
		function    string
		constructor func() engine.SessionStore
		testFunc    func(t *testing.T, ss engine.SessionStore)
	}{
		{
			name: "Successful Retrieval",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				h := heuristic.NewBaseHeuristic(core.RegisterType(0))
				h.SetSUUID(sUUID1)

				_ = ss.AddSession(sUUID1, pUUID1, h)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that the heuristic is retrievable
				h, err := ss.GetInstanceByUUID(sUUID1)
				assert.NoError(t, err)
				assert.Equal(t, h.SUUID(), sUUID1)

				// Ensure that pipeline UUIDs are retrievable
				sUUIDs, err := ss.GetSUUIDsByPUUID(pUUID1)
				assert.NoError(t, err)
				assert.Equal(t, sUUIDs, []core.SUUID{sUUID1})
			},
		},
		{
			name: "Successful Retrieval with Multiple Heuristics",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				h := heuristic.NewBaseHeuristic(core.RegisterType(0))
				h.SetSUUID(sUUID1)

				_ = ss.AddSession(sUUID1, pUUID1, h)

				h2 := heuristic.NewBaseHeuristic(core.RegisterType(0))
				h2.SetSUUID(sUUID2)

				_ = ss.AddSession(sUUID2, pUUID1, h2)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that the first inserted heuristic is retrievable
				h, err := ss.GetInstanceByUUID(sUUID1)
				assert.NoError(t, err)
				assert.Equal(t, h.SUUID(), sUUID1)

				// Ensure that the second inserted heuristic is retrievable
				h2, err := ss.GetInstanceByUUID(sUUID2)
				assert.NoError(t, err)
				assert.Equal(t, h2.SUUID(), sUUID2)

				// Ensure that pipeline UUIDs are retrievable
				sUUIDs, err := ss.GetSUUIDsByPUUID(pUUID1)
				assert.NoError(t, err)
				assert.Equal(t, sUUIDs, []core.SUUID{sUUID1, sUUID2})

				// Ensure that both heuristics are retrievable at once
				hs, err := ss.GetInstancesByUUIDs([]core.SUUID{sUUID1, sUUID2})
				assert.NoError(t, err)
				assert.Equal(t, hs, []heuristic.Heuristic{h, h2})
			},
		},
		{
			name: "Successful Retrieval",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				h := heuristic.NewBaseHeuristic(core.RegisterType(0))
				h.SetSUUID(sUUID1)

				_ = ss.AddSession(sUUID1, pUUID1, h)
				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that the heuristic is retrievable
				h, err := ss.GetInstanceByUUID(sUUID1)
				assert.NoError(t, err)
				assert.Equal(t, h.SUUID(), sUUID1)

				// Ensure that pipeline UUIDs are retrievable
				sUUIDs, err := ss.GetSUUIDsByPUUID(pUUID1)
				assert.NoError(t, err)
				assert.Equal(t, sUUIDs, []core.SUUID{sUUID1})
			},
		},
		{
			name: "Failed Retrieval",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				h := heuristic.NewBaseHeuristic(core.RegisterType(0))
				_ = ss.AddSession(sUUID1, pUUID1, h)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				h, err := ss.GetInstanceByUUID(sUUID2)
				assert.Nil(t, h)
				assert.Error(t, err)
			},
		},
		{
			name: "Failed Add with Duplicate SUUIDs",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				h := heuristic.NewBaseHeuristic(core.RegisterType(0))
				_ = ss.AddSession(sUUID1, pUUID1, h)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that only one suuid can exist in the store
				err := ss.AddSession(sUUID1, pUUID1, heuristic.NewBaseHeuristic(core.RegisterType(0)))
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
