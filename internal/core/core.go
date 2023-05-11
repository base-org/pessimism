package core

import (
	"time"

	"golang.org/x/net/context"
)

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	Timestamp time.Time

	Newtwork Network
	PType    PipelineType
	Type     RegisterType
	Value    any
}

// NewTransitChannel ... Builds new tranit channel
func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}

func (td TransitData) GetRegisterPID() RegisterPID {
	return MakeRegisterPID(
		td.PType,
		td.Type,
	)
}

// InvariantInput ... Standardized type used to supply
// the Risk Engine
type InvariantInput struct {
	PUUID PipelineUUID
	Input TransitData
}
type InvariantInputFunc func(TransitData) InvariantInput

type EngineInputRelay struct {
	ctx context.Context

	pUUID   PipelineUUID
	outChan chan InvariantInput
}

func NewEngineRelay(pUUID PipelineUUID, outChan chan InvariantInput) *EngineInputRelay {
	return &EngineInputRelay{}
}

func (eir *EngineInputRelay) RelayTransitData(td TransitData) error {
	invInput := InvariantInput{
		PUUID: eir.pUUID,
		Input: td,
	}

	eir.outChan <- invInput
	return nil
}

func (eir *EngineInputRelay) Close() error {
	close(eir.outChan)
	return nil
}
