package core

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// PagerDutySeverity represents the severity of an event
type PagerDutySeverity string

const (
	Critical PagerDutySeverity = "critical"
	Error    PagerDutySeverity = "error"
	Warning  PagerDutySeverity = "warning"
	Info     PagerDutySeverity = "info"
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

func (s Severity) ToPagerdutySev() PagerDutySeverity {
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

type AlertRoute string

const (
	AlertRouteSlack     AlertRoute = "slack"
	AlertRoutePagerDuty AlertRoute = "pagerduty"
)

type AlertRoutesTable struct {
	AlertRoutes map[string]AlertRouteMap `yaml:"alertRoutes"`
}

type AlertRouteMap map[string][]map[string]Config
type Config struct {
	URL            string `yaml:"url"`
	Channel        string `yaml:"channel"`
	IntegrationKey string `yaml:"integration_key"`
}

func ParseAlertConfig(path string) (*AlertRoutesTable, error) {
	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	d := &AlertRoutesTable{}
	err = yaml.Unmarshal(f, &d)

	if err != nil {
		return nil, err
	}

	return d, nil
}
