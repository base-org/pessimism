package registry_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func Test_Event_Log_Invariant(t *testing.T) {
	var tests = []struct {
		name     string
		function func(t *testing.T, cfg *registry.EventInvConfig)
	}{
		{
			name: "Successful Invalidation",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventInvariant(&registry.EventInvConfig{
					Address:      "0x0000000000000000000000000000000000000420",
					ContractName: "0x69",
					Sigs:         []string{"0x420"},
				})

				hash := crypto.Keccak256Hash([]byte("0x420"))

				td := core.TransitData{
					Type:    core.EventLog,
					Address: common.HexToAddress("0x0000000000000000000000000000000000000420"),
					Value: types.Log{
						Topics: []common.Hash{hash},
					},
				}

				outcome, invalid, err := ei.Invalidate(td)

				assert.NoError(t, err)
				assert.True(t, invalid)
				assert.NotNil(t, outcome)
			},
		},
		{
			name: "Error Invalidation Due to Mismatched Addresses",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventInvariant(&registry.EventInvConfig{
					Address:      "0x0000000000000000000000000000000000000420",
					ContractName: "0x69",
					Sigs:         []string{"0x420"},
				})

				hash := crypto.Keccak256Hash([]byte("0x420"))

				td := core.TransitData{
					Type:    core.EventLog,
					Address: common.HexToAddress("0x0000000000000000000000000000000000000069"),
					Value: types.Log{
						Topics: []common.Hash{hash},
					},
				}

				outcome, invalid, err := ei.Invalidate(td)

				assert.Error(t, err)
				assert.False(t, invalid)
				assert.Nil(t, outcome)
			},
		},
		{
			name: "No Invalidation Due to Missing Signature",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventInvariant(&registry.EventInvConfig{
					Address:      "0x0000000000000000000000000000000000000420",
					ContractName: "0x69",
					Sigs:         []string{"0x424"},
				})

				hash := crypto.Keccak256Hash([]byte("0x420"))

				td := core.TransitData{
					Type:    core.EventLog,
					Address: common.HexToAddress("0x0000000000000000000000000000000000000420"),
					Value: types.Log{
						Topics: []common.Hash{hash},
					},
				}

				outcome, invalid, err := ei.Invalidate(td)

				assert.NoError(t, err)
				assert.False(t, invalid)
				assert.Nil(t, outcome)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.function(t, nil)
		})
	}

}
