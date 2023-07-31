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

func Test_Event_Log_Heuristic(t *testing.T) {
	var tests = []struct {
		name     string
		function func(t *testing.T, cfg *registry.EventInvConfig)
	}{
		{
			name: "Successful Activation",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventHeuristic(
					&registry.EventInvConfig{
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

				outcome, activated, err := ei.Assess(td)

				assert.NoError(t, err)
				assert.True(t, activated)
				assert.NotNil(t, outcome)
			},
		},
		{
			name: "Error Activation Due to Mismatched Addresses",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventHeuristic(
					&registry.EventInvConfig{
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

				outcome, activated, err := ei.Assess(td)

				assert.Error(t, err)
				assert.False(t, activated)
				assert.Nil(t, outcome)
			},
		},
		{
			name: "No Activation Due to Missing Signature",
			function: func(t *testing.T, cfg *registry.EventInvConfig) {
				ei := registry.NewEventHeuristic(
					&registry.EventInvConfig{
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

				outcome, activated, err := ei.Assess(td)

				assert.NoError(t, err)
				assert.False(t, activated)
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
