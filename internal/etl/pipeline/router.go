package pipeline

import (
	"fmt"

	"github.com/base-org/pessimism/internal/models"
)

type routerOption func(*router) error

func WithDirective(componentID models.ComponentID, outChan chan models.TransitData) routerOption {
	return func(r *router) error {
		return r.AddDirective(componentID, outChan)
	}
}

// Router ... Used as a lookup for components to know where to send output data to and where to read data from
// Adding and removing directives in the equivalent of adding an edge between two nodes using standard graph theory
type router struct {
	outChans map[models.ComponentID]chan models.TransitData
}

// newRouter ... Initializer
func newRouter(opts ...routerOption) (*router, error) {
	router := &router{
		outChans: make(map[models.ComponentID]chan models.TransitData),
	}

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, err
		}
	}

	return router, nil
}

// TransitOutput ... Sends single piece of transitData to all innner mapping value channels
func (router *router) TransitOutput(data models.TransitData) error {
	if len(router.outChans) == 0 {
		return fmt.Errorf("Received transit request with 0 out channels to write to")
	}

	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, channel := range router.outChans {
		channel <- data
	}

	return nil
}

// TransitOutput ... Sends slice of transitData to all innner mapping value channels
func (router *router) TransitOutputs(dataSlice []models.TransitData) error {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, data := range dataSlice {
		if err := router.TransitOutput(data); err != nil {
			return err
		}
	}

	return nil
}

// AddDirective ... Inserts a new output directive given an ID and channel; fail on key collision
func (router *router) AddDirective(componentID models.ComponentID, outChan chan models.TransitData) error {
	if _, found := router.outChans[componentID]; found {
		return fmt.Errorf(dirAlreadyExistsErr, componentID)
	}

	router.outChans[componentID] = outChan
	return nil
}

// RemoveDirective ... Removes an output directive given an ID; fail if no key found
func (router *router) RemoveDirective(componentID models.ComponentID) error {
	if _, found := router.outChans[componentID]; !found {
		return fmt.Errorf(dirNotFoundErr, componentID)
	}

	delete(router.outChans, componentID)
	return nil
}
