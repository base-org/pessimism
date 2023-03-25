package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/conduit/models"

	id "github.com/google/uuid"
)

type RouterOption func(*OutputRouter) error

func WithDirective(componentID id.UUID, outChan chan models.TransitData) RouterOption {
	return func(oo *OutputRouter) error {
		return oo.AddDirective(componentID, outChan)
	}
}

type OutputRouter struct {
	outChans map[id.UUID]chan models.TransitData
}

func NewOutputRouter(opts ...RouterOption) *OutputRouter {
	or := &OutputRouter{
		make(map[id.UUID]chan models.TransitData),
	}

	for _, opt := range opts {
		opt(or)
	}

	return or
}

func (router *OutputRouter) TransitOutput(data models.TransitData) {
	// TODO - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized

	for _, channel := range router.outChans {
		// Write that data
		channel <- data
	}
}

func (router *OutputRouter) AddDirective(componentID id.UUID, outChan chan models.TransitData) error {
	if _, found := router.outChans[componentID]; found {
		return fmt.Errorf("%s already exists within component router mapping", componentID.String())
	}

	router.outChans[componentID] = outChan
	return nil
}

func (router *OutputRouter) RemoveDirective(componentID id.UUID) error {
	if _, found := router.outChans[componentID]; !found {
		return fmt.Errorf("No key %s exists within component router mapping", componentID.String())
	}

	delete(router.outChans, componentID)
	return nil
}
