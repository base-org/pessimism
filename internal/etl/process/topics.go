package process

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type topics struct {
	relays map[core.TopicType]chan core.Event
}

func (p *topics) GetRelay(rt core.TopicType) (chan core.Event, error) {
	val, found := p.relays[rt]
	if !found {
		return nil, fmt.Errorf(topicNotFoundErr, rt.String())
	}

	return val, nil
}

func (p *topics) AddRelay(rt core.TopicType) error {
	if _, found := p.relays[rt]; found {
		return fmt.Errorf(topicExistsErr, rt.String())
	}

	p.relays[rt] = core.NewTransitChannel()

	return nil
}
