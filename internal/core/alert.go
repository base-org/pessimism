package core

import "time"

// AlertDestination ... The destination for an alert
type AlertDestination uint8

const (
	Slack               AlertDestination = iota + 1
	CounterParty                         // 2
	UnknownAlertingDest                  // 3
)

// StringToAlertingDestType ... Converts a string to an alerting destination type
func StringToAlertingDestType(stringType string) AlertDestination {
	switch stringType {
	case "slack":
		return Slack

	case "counterparty":
		return CounterParty
	}

	return UnknownAlertingDest
}

// AlertingPolicy ... The alerting policy for an invariant session
type AlertingPolicy struct {
	Destination AlertDestination
}

// Alert ...
type Alert struct {
	Dest      AlertDestination
	PUUID     PipelineUUID
	SUUID     InvSessionUUID
	Timestamp time.Time
	Ptype     PipelineType

	Content string
}
