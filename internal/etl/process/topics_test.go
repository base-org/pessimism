package process

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestAddRemoveTopic(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		construct func() *topics
		test      func(*testing.T, *topics)
	}{
		{
			name:        "Successful Add Test",
			description: "When a register type is passed to AddIngress function, it should be successfully added to handler ingress mapping",

			construct: func() *topics {
				return &topics{
					relays: make(map[core.TopicType]chan core.Event),
				}
			},

			test: func(t *testing.T, p *topics) {

				err := p.AddRelay(core.BlockHeader)
				assert.NoError(t, err, "geth.block register should added as an egress")

			},
		},
		{
			name:        "Failed Add Test",
			description: "When the same register type is added twice to AddIngress function, the second add should fail with key collisions",

			construct: func() *topics {
				p := &topics{
					relays: make(map[core.TopicType]chan core.Event),
				}

				if err := p.AddRelay(core.BlockHeader); err != nil {
					panic(err)
				}
				return p
			},

			test: func(t *testing.T, p *topics) {
				err := p.AddRelay(core.BlockHeader)

				assert.Error(t, err, "geth.block register should fail to be added")
				assert.Equal(t, err.Error(), fmt.Sprintf(topicExistsErr, core.BlockHeader.String()))

			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			publisher := tc.construct()
			tc.test(t, publisher)
		})

	}
}
