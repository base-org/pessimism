package core

import (
	"time"
)

// PagerDutySeverity represents the severity of an event
type PagerDutySeverity string

const (
	Critical PagerDutySeverity = "critical"
	Error    PagerDutySeverity = "error"
	Warning  PagerDutySeverity = "warning"
	Info     PagerDutySeverity = "info"
)

// AlertStatus ... A standardized response status for alert clients
type AlertStatus string

const (
	SuccessStatus AlertStatus = "success"
	FailureStatus AlertStatus = "failure"
)

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

// String ... Converts a severity to a string
func (s Severity) String() string {
	switch s {
	case LOW:
		return "low"
	case MEDIUM:
		return "medium"
	case HIGH:
		return "high"

	case UNKNOWN:
		return UnknownType

	default:
		return UnknownType
	}
}

// ToPagerDutySev ... Converts a severity to a pagerduty severity
// This mapping is opinionated as follows:
//   - LOW -> Warning
//   - MEDIUM -> Error
//   - HIGH -> Critical
//   - UNKNOWN -> Error
func (s Severity) ToPagerDutySev() PagerDutySeverity {
	switch s {
	case LOW:
		return Warning
	case MEDIUM:
		return Error
	case HIGH:
		return Critical

	case UNKNOWN:
		return Error

	default:
		return Error
	}
}

// Alert ... An alert
type Alert struct {
	Criticality Severity
	PUUID       PUUID
	SUUID       SUUID
	Timestamp   time.Time
	Ptype       PipelineType

	Content string
}

// AlertRoutingParams ... The routing parameters for alerts
type AlertRoutingParams struct {
	AlertRoutes *SeverityMap `yaml:"alertRoutes"`
}

// SeverityMap ... A map of severity to alert client config
type SeverityMap struct {
	Low    *AlertClientCfg `yaml:"low"`
	Medium *AlertClientCfg `yaml:"medium"`
	High   *AlertClientCfg `yaml:"high"`
}

// AlertClientCfg ... The alert client config
type AlertClientCfg struct {
	Slack     map[string]*Config `yaml:"slack"`
	PagerDuty map[string]*Config `yaml:"pagerduty"`
}

// Config ... The config for an alert client
type Config struct {
	URL            string `yaml:"url"`
	Channel        string `yaml:"channel"`
	IntegrationKey string `yaml:"integration_key"`
}
