package core

import "time"

// AlertDestination ... The destination for an alert
type AlertDestination uint8

const (
	Slack        AlertDestination = iota + 1
	CounterParty                  // 2
)

// AlertingPolicy ... The alerting policy for an invariant session
type AlertingPolicy struct {
	Destination AlertDestination
}

// Alert ...
type Alert struct {
	Dest      AlertDestination
	SUUID     InvSessionUUID
	Timestamp time.Time

	Content string
}

// NewAlert ... Alert initializer
func NewAlert(ts time.Time, sUUID InvSessionUUID, content string) Alert {
	return Alert{
		SUUID:     sUUID,
		Timestamp: ts,
		Content:   content,
	}
}
