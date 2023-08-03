package core

import "time"

// Severity ... The severity of an alert
type Severity uint8

const (
	UNKNOWN Severity = iota

	LOW
	MEDIUM
	HIGH
)

// StringToSev ... Converts a string to a severity
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

func (s Severity) String() string {
	switch s {
	case LOW:
		return "low"
	case MEDIUM:
		return "medium"
	case HIGH:
		return "high"
	default:
		return "unknown"
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
