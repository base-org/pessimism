package component

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// egress ... Used to route transit data from a component to it's respective edge components.
// Also used to manage egresses or "edge routes" for some component.
type egressHandler struct {
	egresses map[core.ComponentID]chan core.TransitData
}

// newEgress ... Initializer
func newEgressHandler() *egressHandler {
	return &egressHandler{
		egresses: make(map[core.ComponentID]chan core.TransitData),
	}
}

// TransitOutput ... Sends single piece of transitData to all innner mapping value channels
func (eh *egressHandler) TransitOutput(td core.TransitData) error {
	if len(eh.egresses) == 0 {
		return fmt.Errorf("received transit request with 0 out channels to write to")
	}

	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, channel := range eh.egresses {
		channel <- td
	}

	return nil
}

// TransitOutput ... Sends slice of transitData to all innner mapping value channels
func (eh *egressHandler) TransitOutputs(dataSlice []core.TransitData) error {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, data := range dataSlice {
		if err := eh.TransitOutput(data); err != nil {
			return err
		}
	}

	return nil
}

// AddEgress ... Inserts a new egress given an ID and channel; fail on key collision
func (eh *egressHandler) AddEgress(componentID core.ComponentID, outChan chan core.TransitData) error {
	if _, found := eh.egresses[componentID]; found {
		return fmt.Errorf(egressAlreadyExistsErr, componentID.String())
	}

	eh.egresses[componentID] = outChan
	return nil
}

// RemoveEGress ... Removes an egress given an ID; fail if no key found
func (egress *egressHandler) RemoveEgress(componentID core.ComponentID) error {
	if _, found := egress.egresses[componentID]; !found {
		return fmt.Errorf(egressNotFoundErr, componentID.String())
	}

	delete(egress.egresses, componentID)
	return nil
}
