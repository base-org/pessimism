package alert_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestPagerDutyMessage(t *testing.T) {

	a := core.Alert{
		HeuristicID: core.UUID{},
		Content:     "Test alert",
	}

	expected := "\n\tHeuristic Triggered: unknown\n\tNetwork: unknown\n\tAssessment: \n\tTest alert\n\t"
	actual := new(alert.Interpolator).PagerDutyMessage(a)
	assert.Equal(t, expected, actual)
}
