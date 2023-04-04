package models

import (
	"time"

	"github.com/google/uuid"
)

type RegisterType string

type ComponentID = uuid.UUID

type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value any
}

func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}
