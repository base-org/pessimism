package models

import (
	"strconv"
	"time"
)

type RegisterType string

type ComponentID = int

// NOTE - In the future this should take in a SessionID, a RegisterType, and a pipeline type
// where ID = int(sessionID, registerType, pipelineType)
func StringToComponentID(in string) (*ComponentID, error) {
	asInt, err := strconv.Atoi(in)
	if err != nil {
		return nil, err
	}

	return &asInt, nil
}

type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value any
}

type TransitChannel = chan TransitData

func NewTransitChannel() TransitChannel {
	return make(chan TransitData)
}
