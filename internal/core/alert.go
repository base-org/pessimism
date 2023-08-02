package core

import "time"

// Alert ... An alert
type Alert struct {
	Dest      AlertDestination
	PUUID     PUUID
	SUUID     SUUID
	Timestamp time.Time
	Ptype     PipelineType

	Content string
}
