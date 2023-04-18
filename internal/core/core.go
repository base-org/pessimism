package core

import (
	"time"
)

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value any
}

func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}
