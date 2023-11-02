package core

import (
	"math/big"
	"time"
)

// ClientConfig ... Configuration passed through to an reader component constructor
type ClientConfig struct {
	Network      Network
	PollInterval time.Duration
	NumOfRetries int
	StartHeight  *big.Int
	EndHeight    *big.Int
}

// SessionConfig ... Configuration passed through to a session constructor
type SessionConfig struct {
	Network     Network
	PT          PipelineType
	AlertPolicy *AlertPolicy
	Type        HeuristicType
	Params      *SessionParams
}

// PipelineConfig ... Configuration passed through to a pipeline constructor
type PipelineConfig struct {
	Network      Network
	DataType     RegisterType
	PipelineType PipelineType
	ClientConfig *ClientConfig
}

// Backfill ... Returns true if the reader is configured to backfill
func (oc *ClientConfig) Backfill() bool {
	return oc.StartHeight != nil
}
