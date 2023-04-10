package core

import (
	"time"
)

type Network uint8

const (
	Layer1 = iota + 1
	Layer2
)

func (n Network) String() string {
	switch n {
	case Layer1:
		return "layer1"

	case Layer2:
		return "layer2"
	}

	return "unknown"
}

type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value any
}

func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}
