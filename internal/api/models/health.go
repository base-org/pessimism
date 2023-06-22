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

// NodeConnectionStatus ... Used to display health status of each node connection
type NodeConnectionStatus struct {
	IsL1Healthy bool
	IsL2Healthy bool
}
