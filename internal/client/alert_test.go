package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

func TestToPagerDutyEvent(t *testing.T) {
	alert := &client.AlertEventTrigger{
		Message:  "test",
		Severity: core.HIGH,
		DedupKey: core.PathID{},
	}

	sPathID := alert.DedupKey.String()
	res := alert.ToPagerdutyEvent()
	assert.Equal(t, core.Critical, res.Severity)
	assert.Equal(t, "test", res.Message)
	assert.Equal(t, sPathID, res.DedupKey)

	alert.Severity = core.MEDIUM
	res = alert.ToPagerdutyEvent()
	assert.Equal(t, core.Error, res.Severity)

	alert.Severity = core.LOW
	res = alert.ToPagerdutyEvent()
	assert.Equal(t, core.Warning, res.Severity)
}
