package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// EventInvConfig  ... Configuration for the event heuristic
type EventInvConfig struct {
	ContractName string   `json:"contract_name"`
	Address      string   `json:"address"`
	Sigs         []string `json:"args"`
}

// Unmarshal ... Converts a general config to an event heuristic config
func (eic *EventInvConfig) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &eic)
}

// EventHeuristic ...
type EventHeuristic struct {
	cfg  *EventInvConfig
	sigs []common.Hash

	heuristic.Heuristic
}

// eventReportMsg ... Message to be sent to the alerting system
const eventReportMsg = `
	_Monitored Event Triggered_

	Contract Name: %s
	Contract Address: %s 
	Transaction Hash: %s
	Event: %s
`

// NewEventHeuristic ... Initializer
func NewEventHeuristic(cfg *EventInvConfig) heuristic.Heuristic {
	var sigs []common.Hash

	for _, sig := range cfg.Sigs {
		sigs = append(sigs, crypto.Keccak256Hash([]byte(sig)))
	}

	return &EventHeuristic{
		cfg:  cfg,
		sigs: sigs,

		Heuristic: heuristic.NewBaseHeuristic(core.EventLog),
	}
}

// Assess ... Checks if the balance is within the bounds
// specified in the config
func (ei *EventHeuristic) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	// 1. Validate and extract the log event from the transit data
	err := ei.ValidateInput(td)
	if err != nil {
		return nil, err
	}

	if td.Address != common.HexToAddress(ei.cfg.Address) {
		return nil, fmt.Errorf(invalidAddrErr, ei.cfg.Address, td.Address.String())
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	// 2. Check if the log event signature is in the list of signatures
	activated := false
	var sigHit string
	for i, sig := range ei.sigs {
		if log.Topics[0] == sig {
			activated = true
			sigHit = ei.cfg.Sigs[i]
			break
		}
	}

	if !activated {
		return heuristic.NoActivations(), nil
	}

	return heuristic.NewActivationSet().Add(&heuristic.Activation{
		TimeStamp: time.Now(),
		Message:   fmt.Sprintf(eventReportMsg, ei.cfg.ContractName, log.Address, log.TxHash.Hex(), sigHit),
	}), nil
}
