package component

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

// egressHandler ... Used to route transit data from a component to it's respective edge components.
// Also used to manage egresses or "edge routes" for some component.
type egressHandler struct {
	egresses map[core.ComponentPID]chan core.TransitData

	relay *core.EngineInputRelay
}

// newEgress ... Initializer
func newEgressHandler() *egressHandler {
	return &egressHandler{
		egresses: make(map[core.ComponentPID]chan core.TransitData),
		relay:    nil,
	}
}

// Send ... Sends single piece of transitData to all innner mapping value channels
func (eh *egressHandler) Send(td core.TransitData) error {
	if len(eh.egresses) == 0 && !eh.HasEngineEgress() {
		return fmt.Errorf(egressNotExistErr)
	}

	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, channel := range eh.egresses {
		channel <- td
	}

	return nil
}

// SendBatch ... Sends slice of transitData to all innner mapping value channels
func (eh *egressHandler) SendBatch(dataSlice []core.TransitData) error {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged chanel buffers is recognized
	for _, data := range dataSlice {
		// NOTE - Does it make sense to fail loudly here?

		if err := eh.Send(data); err != nil {
			return err
		}

		if eh.HasEngineEgress() {
			if err := eh.relay.RelayTransitData(data); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddEgress ... Inserts a new egress given an ID and channel; fail on key collision
func (eh *egressHandler) AddEgress(componentID core.ComponentUUID, outChan chan core.TransitData) error {
	if _, found := eh.egresses[componentID.PID]; found {
		return fmt.Errorf(egressAlreadyExistsErr, componentID.String())
	}

	eh.egresses[componentID.PID] = outChan
	return nil
}

// RemoveEgress ... Removes an egress given an ID; fail if no key found
func (eh *egressHandler) RemoveEgress(componentID core.ComponentUUID) error {
	if _, found := eh.egresses[componentID.PID]; !found {
		return fmt.Errorf(egressNotFoundErr, componentID.PID.String())
	}

	delete(eh.egresses, componentID.PID)

	return nil
}

func (eh *egressHandler) HasEngineEgress() bool {
	return eh.relay != nil
}

func (eh *egressHandler) AddRelay(relay *core.EngineInputRelay) error {
	logging.NoContext().Debug("Adding egress to risk engine")

	if eh.HasEngineEgress() {
		return fmt.Errorf("engine egress already exists")
	}

	eh.relay = relay

	return nil
}
