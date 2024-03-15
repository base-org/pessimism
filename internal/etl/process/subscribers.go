package process

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type subscribers struct {
	subs map[core.ProcIdentifier]chan core.Event

	relay *core.ExecInputRelay
}

func (s *subscribers) None() bool {
	return len(s.subs) == 0 && s.HasEngineRelay()
}

func (s *subscribers) Publish(e core.Event) error {
	if len(s.subs) == 0 && !s.HasEngineRelay() {
		return fmt.Errorf(noSubErr)
	}

	if s.HasEngineRelay() {
		if err := s.relay.RelayEvent(e); err != nil {
			return err
		}
	}

	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged channel buffers is recognized
	for _, channel := range s.subs {
		channel <- e
	}

	return nil
}

func (s *subscribers) PublishBatch(dataSlice []core.Event) error {
	// NOTE - Consider introducing a fail safe timeout to ensure that freezing on clogged channel buffers is recognized
	for _, data := range dataSlice {
		// NOTE - Does it make sense to fail loudly here?

		if err := s.Publish(data); err != nil {
			return err
		}
	}

	return nil
}

func (s *subscribers) AddSubscriber(id core.ProcessID, topic chan core.Event) error {
	if _, found := s.subs[id.ID]; found {
		return fmt.Errorf(subExistsErr, id.String())
	}

	s.subs[id.ID] = topic
	return nil
}

func (s *subscribers) RemoveSubscriber(id core.ProcessID) error {
	if _, found := s.subs[id.ID]; !found {
		return fmt.Errorf(subNotFound, id.ID.String())
	}

	delete(s.subs, id.ID)
	return nil
}

func (s *subscribers) HasEngineRelay() bool {
	return s.relay != nil
}

func (s *subscribers) AddEngineRelay(relay *core.ExecInputRelay) error {
	if s.HasEngineRelay() {
		return fmt.Errorf(engineRelayExists)
	}

	s.relay = relay
	return nil
}
