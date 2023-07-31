package core

import "time"

// AlertingPolicy ... The alerting policy for a heuristic session
// NOTE - This could be extended to support additional
// policy metadata like criticality, etc.
type AlertingPolicy struct {
	Destination AlertDestination
}

// Alert ... An alert
type Alert struct {
	Dest      AlertDestination
	PUUID     PUUID
	SUUID     SUUID
	Timestamp time.Time
	Ptype     PipelineType

	Content string
}
