package core

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type TransitOption = func(TransitData) TransitData

// WithAddress ... Injects address to transit data
func WithAddress(address common.Address) TransitOption {
	return func(td TransitData) TransitData {
		td.Address = &address
		return td
	}
}

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	Timestamp time.Time

	Network Network
	PType   PipelineType
	Type    RegisterType

	Address *common.Address
	Value   any
}

func NewTransitData(rt RegisterType, val any, opts ...TransitOption) TransitData {
	td := TransitData{
		Timestamp: time.Now(),

		Type:  rt,
		Value: val,
	}

	for _, opt := range opts {
		opt(td)
	}

	return td
}

// NewTransitChannel ... Builds new tranit channel
func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}

// InvariantInput ... Standardized type used to supply
// the Risk Engine
type InvariantInput struct {
	PUUID PipelineUUID
	Input TransitData
}

// EngineInputRelay ... Represents a inter-subsystem
// relay used to bind final ETL pipeline outputs to risk engine inputs
type EngineInputRelay struct {
	pUUID   PipelineUUID
	outChan chan InvariantInput
}

// NewEngineRelay ... Initializer
func NewEngineRelay(pUUID PipelineUUID, outChan chan InvariantInput) *EngineInputRelay {
	return &EngineInputRelay{
		pUUID:   pUUID,
		outChan: outChan,
	}
}

// RelayTransitData ... Creates invariant input from transit data to send to risk engine
func (eir *EngineInputRelay) RelayTransitData(td TransitData) error {
	invInput := InvariantInput{
		PUUID: eir.pUUID,
		Input: td,
	}

	eir.outChan <- invInput
	return nil
}
