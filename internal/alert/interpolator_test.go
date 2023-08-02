package alert_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_InterpolateSlackMessage(t *testing.T) {
	sUUID := core.NilSUUID()

	msg := "Friedrich Nietzsche"

	expected := "\n\t⚠️🚨 Pessimism Alert: unknown 🚨⚠️\n\n\t_Heuristic activation conditions met_\n\n\t_Network:_ unknown\n\t_Session UUID:_ unknown:unknown:unknown::000000000\n\n\t*Assessment Content:* \n\t```Friedrich Nietzsche```\t\n\t"

	actual := alert.NewInterpolator().
		InterpolateSlackMessage(sUUID, msg)

	assert.Equal(t, expected, actual, "should be equal")
}

func Test_InterpolatePagerdutyMessage(t *testing.T) {
	sUUID := core.NilSUUID()

	msg := "Test alert"

	expected := "\n\tHeuristic Triggered: unknown\n\tNetwork: unknown\n\tAssessment: \n\tTest alert\n\t"

	actual := alert.NewInterpolator().InterpolatePagerdutyMessage(sUUID, msg)

	assert.Equal(t, expected, actual, "should be equal")
}
