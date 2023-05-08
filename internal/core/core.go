package core

import (
	"time"
)

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	Timestamp time.Time

	PType PipelineType
	Type  RegisterType
	Value any
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
