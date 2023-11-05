package heuristic_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/stretchr/testify/assert"
)

func Test_BaseHeuristic(t *testing.T) {
	expected := core.NewUUID()
	h := heuristic.New(core.TopicType(0), core.BalanceEnforcement)

	h.SetID(expected)
	actual := h.ID()

	assert.Equal(t, expected, actual)

	tt := h.TopicType()
	assert.Equal(t, core.TopicType(0), tt)

	// Test validate

	err := h.Validate(core.Event{
		Type: core.TopicType(0),
	})

	assert.Nil(t, err)

	err = h.Validate(core.Event{
		Type: core.TopicType(1),
	})

	assert.NotNil(t, err)

	ht := h.Type()
	assert.Equal(t, core.BalanceEnforcement, ht)
}
