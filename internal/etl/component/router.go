package component

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type RouterOption func(*router) error

func WithDirective(componentID core.ComponentID, outChan chan core.TransitData) RouterOption {
	return func(r *router) error {
		return r.AddDirective(componentID, outChan)
	}
}

// Router ... Used as a lookup for components to know where to send output data to and where to read data from
// Adding and removing directives in the equivalent of adding an edge between two nodes using standard graph theory
type router struct {
	outChans map[core.ComponentID]chan core.TransitData
}

// newRouter ... Initializer
func newRouter(opts ...RouterOption) (*router, error) {
	router := &router{
		outChans: make(map[core.ComponentID]chan core.TransitData),
	}

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, err
		}
	}

	return router, nil
}

// TransitOutput ... Sends single piece of transitData to all innner mapping value channels
func (router *router) TransitOutput(td core.TransitData) error {
	if len(router.outChans) == 0 {
		return fmt.Errorf("received transit request with 0 out channels to write to")
	}

	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, channel := range router.outChans {
		channel <- td
	}

	return nil
}

// TransitOutput ... Sends slice of transitData to all innner mapping value channels
func (router *router) TransitOutputs(dataSlice []core.TransitData) error {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, data := range dataSlice {
		if err := router.TransitOutput(data); err != nil {
			return err
		}
	}

	return nil
}

// AddDirective ... Inserts a new output directive given an ID and channel; fail on key collision
func (router *router) AddDirective(componentID core.ComponentID, outChan chan core.TransitData) error {
	if _, found := router.outChans[componentID]; found {
		return fmt.Errorf(dirAlreadyExistsErr, componentID.String())
	}

	router.outChans[componentID] = outChan
	return nil
}

// RemoveDirective ... Removes an output directive given an ID; fail if no key found
func (router *router) RemoveDirective(componentID core.ComponentID) error {
	if _, found := router.outChans[componentID]; !found {
		return fmt.Errorf(dirNotFoundErr, componentID.String())
	}

	delete(router.outChans, componentID)
	return nil
}
