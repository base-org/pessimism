package core

import "time"

type Severity int

const (
	LOW = iota
	MEDIUM
	HIGH

	UNKNOWN
)

func StringToSev(stringType string) Severity {
	switch stringType {
	case "low":
		return LOW
	case "medium":
		return MEDIUM
	case "high":
		return HIGH
	default:
		return UNKNOWN
	}
}

// Alert ... An alert
type Alert struct {
	Criticality Severity
	Dest        AlertDestination
	PUUID       PUUID
	SUUID       SUUID
	Timestamp   time.Time
	Ptype       PipelineType

	Content string
}
