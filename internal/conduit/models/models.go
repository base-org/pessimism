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

type ComponentType int

const (
	Oracle   ComponentType = 0
	Pipe     ComponentType = 1
	Conveyor ComponentType = 2
)