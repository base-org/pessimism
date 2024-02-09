package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

func TestToPagerDutyEvent(t *testing.T) {
	alert := &client.AlertEventTrigger{
		Message: "test",
		Alert: core.Alert{
			Sev:    core.HIGH,
			PathID: core.PathID{},
		},
	}

	sPathID := alert.Alert.PathID.String()
	res := alert.ToPagerdutyEvent()
	assert.Equal(t, core.Critical, res.Severity)
	assert.Equal(t, "test", res.Message)
	assert.Equal(t, sPathID, res.DedupKey)

	alert.Alert.Sev = core.MEDIUM
	res = alert.ToPagerdutyEvent()
	assert.Equal(t, core.Error, res.Severity)

	alert.Alert.Sev = core.LOW
	res = alert.ToPagerdutyEvent()
	assert.Equal(t, core.Warning, res.Severity)
}
