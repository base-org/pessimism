package registry_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/engine/registry"
)

func Test_Event_Log_Invariant(t *testing.T) {
	var tests = []struct {
		name         string
		cfgConstruct func() *registry.EventInvConfig
		function     func(t *testing.T, cfg *registry.EventInvConfig)
	}{
		{
			name: "Successful Invalidation",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventInvariant(&registry.EventInvConfig{
					Address:      "0x123",
					ContractName: "0x69",
					Sigs:         []string{"0x420"},
				})

				hash := crypto.Keccak256Hash([]byte("0x420"))

				td := &registry.TransitData{
					Type: registry.EventLog,
					Address: "0x123",
					Value: types.Log{
						Topics: []common.Hash{
					},
			},
		},
	}

}
