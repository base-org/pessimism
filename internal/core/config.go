package core

import (
	"math/big"
	"time"
)

// ClientConfig ... Configuration passed through to an oracle component constructor
type ClientConfig struct {
	Network      Network
	PollInterval time.Duration
	NumOfRetries int
	StartHeight  *big.Int
	EndHeight    *big.Int
}

// SessionConfig ... Configuration passed through to a session constructor
type SessionConfig struct {
	AlertDest AlertDestination
	Type      InvariantType
	Params    InvSessionParams
}

// PipelineConfig ... Configuration passed through to a pipeline constructor
type PipelineConfig struct {
	Network      Network
	DataType     RegisterType
	PipelineType PipelineType
	ClientConfig *ClientConfig
}

// Backfill ... Returns true if the oracle is configured to backfill
func (oc *ClientConfig) Backfill() bool {
	return oc.StartHeight != nil
}

// Backtest ... Returns true if the oracle is configured to backtest
func (oc *ClientConfig) Backtest() bool {
	return oc.EndHeight != nil
}
