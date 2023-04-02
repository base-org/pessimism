package models

import (
	"time"
)

type RegisterType string

type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value interface{}
}

type TransitChannel = chan TransitData

func NewTransitChannel() TransitChannel {
	return make(chan TransitData)
}
