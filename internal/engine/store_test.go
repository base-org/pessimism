package engine_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/stretchr/testify/assert"
)

func TestSessionStore(t *testing.T) {
	sUUID1 := core.MakeSUUID(core.Layer1, core.Live, core.InvariantType(0))
	sUUID2 := core.MakeSUUID(core.Layer2, core.Live, core.InvariantType(0))
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

				inv := invariant.NewBaseInvariant(core.RegisterType(0))
				inv.SetSUUID(sUUID1)

				_ = ss.AddInvSession(sUUID1, pUUID1, inv)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that the invariant is retrievable
				inv, err := ss.GetInstanceByUUID(sUUID1)
				assert.NoError(t, err)
				assert.Equal(t, inv.SUUID(), sUUID1)

				// Ensure that pipeline UUIDs are retrievable
				sUUIDs, err := ss.GetSUUIDsByPUUID(pUUID1)
				assert.NoError(t, err)
				assert.Equal(t, sUUIDs, []core.SUUID{sUUID1})
			},
		},
		{
			name: "Successful Retrieval with Multiple Invariants",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				inv := invariant.NewBaseInvariant(core.RegisterType(0))
				inv.SetSUUID(sUUID1)

				_ = ss.AddInvSession(sUUID1, pUUID1, inv)

				inv2 := invariant.NewBaseInvariant(core.RegisterType(0))
				inv2.SetSUUID(sUUID2)

				_ = ss.AddInvSession(sUUID2, pUUID1, inv2)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that the first inserted invariant is retrievable
				inv, err := ss.GetInstanceByUUID(sUUID1)
				assert.NoError(t, err)
				assert.Equal(t, inv.SUUID(), sUUID1)

				// Ensure that the second inserted invariant is retrievable
				inv2, err := ss.GetInstanceByUUID(sUUID2)
				assert.NoError(t, err)
				assert.Equal(t, inv2.SUUID(), sUUID2)

				// Ensure that pipeline UUIDs are retrievable
				sUUIDs, err := ss.GetSUUIDsByPUUID(pUUID1)
				assert.NoError(t, err)
				assert.Equal(t, sUUIDs, []core.SUUID{sUUID1, sUUID2})

				// Ensure that both invariants are retrievable at once
				invs, err := ss.GetInstancesByUUIDs([]core.SUUID{sUUID1, sUUID2})
				assert.NoError(t, err)
				assert.Equal(t, invs, []invariant.Invariant{inv, inv2})
			},
		},
		{
			name: "Successful Retrieval",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				inv := invariant.NewBaseInvariant(core.RegisterType(0))
				inv.SetSUUID(sUUID1)

				_ = ss.AddInvSession(sUUID1, pUUID1, inv)
				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that the invariant is retrievable
				inv, err := ss.GetInstanceByUUID(sUUID1)
				assert.NoError(t, err)
				assert.Equal(t, inv.SUUID(), sUUID1)

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

				inv := invariant.NewBaseInvariant(core.RegisterType(0))
				_ = ss.AddInvSession(sUUID1, pUUID1, inv)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				inv, err := ss.GetInstanceByUUID(sUUID2)
				assert.Nil(t, inv)
				assert.Error(t, err)
			},
		},
		{
			name: "Failed Add with Duplicate SUUIDs",
			constructor: func() engine.SessionStore {
				ss := engine.NewSessionStore()

				inv := invariant.NewBaseInvariant(core.RegisterType(0))
				_ = ss.AddInvSession(sUUID1, pUUID1, inv)

				return ss
			},
			testFunc: func(t *testing.T, ss engine.SessionStore) {
				// Ensure that only one suuid can exist in the store
				err := ss.AddInvSession(sUUID1, pUUID1, invariant.NewBaseInvariant(core.RegisterType(0)))
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
