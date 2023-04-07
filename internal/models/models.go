package models

import (
	"time"
)

type RegisterType string

type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value any
}

func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}
