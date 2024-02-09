package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestSNSMessagePayload_Marshal(t *testing.T) {
	alert := core.Alert{
		Net:         core.Layer1,
		HT:          core.BalanceEnforcement,
		Sev:         core.HIGH,
		PathID:      core.MakePathID(0, core.MakeProcessID(core.Live, 0, 0, 0), core.MakeProcessID(core.Live, 0, 0, 0)),
		HeuristicID: core.UUID{},
		Timestamp:   time.Time{},
		Content:     "test",
	}

	event := &AlertEventTrigger{
		Message: "test",
		Alert:   alert,
	}

	payload, err := event.ToSNSMessagePayload().Marshal()
	if err != nil {
		t.Fatal(err)
	}

	var snsPayload SNSMessage
	err = json.Unmarshal(payload, &snsPayload)
	if err != nil {
		t.Fatal(err)
	}

	var snsMsgPayload SNSMessagePayload
	err = json.Unmarshal([]byte(snsPayload.Default), &snsMsgPayload)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, core.Layer1.String(), snsMsgPayload.Network)
	assert.Equal(t, core.BalanceEnforcement.String(), snsMsgPayload.HeuristicType)
	assert.Equal(t, core.HIGH.String(), snsMsgPayload.Severity)
	assert.Equal(t, "test", snsMsgPayload.Content)
	assert.Equal(t, alert.PathID.String(), snsMsgPayload.PathID)
	assert.Equal(t, alert.HeuristicID.String(), snsMsgPayload.HeuristicID)
	assert.Equal(t, alert.Timestamp, snsMsgPayload.Timestamp)
}
