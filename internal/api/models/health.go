package models

import (
	"time"
)

// HealthCheck ... Returns health status of server
// Currently just returns True
type HealthCheck struct {
	Timestamp             time.Time
	Healthy               bool
	ChainConnectionStatus *ChainConnectionStatus
}

// ChainConnectionStatus ... Used to display health status of each node connection
type ChainConnectionStatus struct {
	IsL1Healthy bool
	IsL2Healthy bool
}
