package alert_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_InterpolatePagerDutyMessage(t *testing.T) {
	sUUID := core.NilSUUID()

	msg := "Test alert"

	expected := "\n\tHeuristic Triggered: unknown\n\tNetwork: unknown\n\tAssessment: \n\tTest alert\n\t"

	actual := alert.NewInterpolator().InterpolatePagerDutyMessage(sUUID, msg)

	assert.Equal(t, expected, actual, "should be equal")
}
