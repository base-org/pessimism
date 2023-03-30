package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/conduit/models"
)

const (
	alreadyExistsErr = "%d already exists within component router mapping"
	notFoundErr      = "no key %d exists within component router mapping"
)

type RouterOption func(*OutputRouter) error

func WithDirective(componentID int, outChan chan models.TransitData) RouterOption {
	return func(r *OutputRouter) error {
		return r.AddDirective(componentID, outChan)
	}
}

type OutputRouter struct {
	outChans map[int]chan models.TransitData
}

// NewOutputRouter ... Initializer
func NewOutputRouter(opts ...RouterOption) (*OutputRouter, error) {
	router := &OutputRouter{
		make(map[int]chan models.TransitData),
	}

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, err
		}
	}

	return router, nil
}

// TransitOutput ... Sends single piece of transitData to all innner mapping value channels
func (router *OutputRouter) TransitOutput(data models.TransitData) {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, channel := range router.outChans {
		channel <- data
	}
}

// TransitOutput ... Sends slice of transitData to all innner mapping value channels
func (router *OutputRouter) TransitOutputs(dataSlice []models.TransitData) {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, data := range dataSlice {
		router.TransitOutput(data)
	}
}

// AddDirective ... Inserts a new output directive given an ID and channel; fail on key collision
func (router *OutputRouter) AddDirective(componentID int, outChan chan models.TransitData) error {
	if _, found := router.outChans[componentID]; found {
		return fmt.Errorf(alreadyExistsErr, componentID)
	}

	router.outChans[componentID] = outChan
	return nil
}

// RemoveDirective ... Removes an output directive given an ID; fail if no key found
func (router *OutputRouter) RemoveDirective(componentID int) error {
	if _, found := router.outChans[componentID]; !found {
		return fmt.Errorf(notFoundErr, componentID)
	}

	delete(router.outChans, componentID)
	return nil
}
