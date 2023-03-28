package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/rs/xid"
)

type RouterOption func(*OutputRouter) error

func WithDirective(componentID xid.ID, outChan chan models.TransitData) RouterOption {
	return func(r *OutputRouter) error {
		return r.AddDirective(componentID, outChan)
	}
}

type OutputRouter struct {
	outChans map[xid.ID]chan models.TransitData
}

func NewOutputRouter(opts ...RouterOption) *OutputRouter {

	router := &OutputRouter{
		make(map[xid.ID]chan models.TransitData),
	}

	for _, opt := range opts {
		opt(router)
	}

	return router
}

func (router *OutputRouter) TransitOutput(data models.TransitData) {
	// TODO - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized

	for _, channel := range router.outChans {
		channel <- data
	}
}

func (router *OutputRouter) AddDirective(componentID xid.ID, outChan chan models.TransitData) error {
	if _, found := router.outChans[componentID]; found {
		return fmt.Errorf("%s already exists within component router mapping", componentID.String())
	}

	router.outChans[componentID] = outChan
	return nil
}

func (router *OutputRouter) RemoveDirective(componentID xid.ID) error {
	if _, found := router.outChans[componentID]; !found {
		return fmt.Errorf("No key %s exists within component router mapping", componentID.String())
	}

	delete(router.outChans, componentID)
	return nil
}
