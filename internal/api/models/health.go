package models

import (
	"time"
)

// HealthCheck ... Returns health status of server
// Currently just returns True
type HealthCheck struct {
	Timestamp time.Time
	Healthy   bool
}
