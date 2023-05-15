package models

import (
	"encoding/json"
	"time"
)

// HealthCheck ... Returns health status of server
// Currently just returns True
type HealthCheck struct {
	Timestamp time.Time
	Healthy   bool
}

func (hc *HealthCheck) UnmarshalJson(blob []byte) error {
	return json.Unmarshal(blob, hc)
}
