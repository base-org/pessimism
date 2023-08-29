package core_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestStringToSev(t *testing.T) {
	assert.Equal(t, core.StringToSev("low"), core.LOW)
	assert.Equal(t, core.StringToSev("medium"), core.MEDIUM)
	assert.Equal(t, core.StringToSev("high"), core.HIGH)
	assert.Equal(t, core.StringToSev("unknown"), core.UNKNOWN)
	assert.Equal(t, core.StringToSev(""), core.UNKNOWN)
}

func TestSeverity_String(t *testing.T) {
	assert.Equal(t, core.LOW.String(), "low")
	assert.Equal(t, core.MEDIUM.String(), "medium")
	assert.Equal(t, core.HIGH.String(), "high")
	assert.Equal(t, core.UNKNOWN.String(), "unknown")
}

func TestToPagerDutySev(t *testing.T) {

	assert.Equal(t, core.LOW.ToPagerDutySev(), core.PagerDutySeverity("warning"))
	assert.Equal(t, core.MEDIUM.ToPagerDutySev(), core.PagerDutySeverity("error"))
	assert.Equal(t, core.HIGH.ToPagerDutySev(), core.PagerDutySeverity("critical"))

}
