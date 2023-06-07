package core

import (
	"math/big"
	"time"
)

// ClientConfig ... Configuration passed through to an oracle component constructor
type ClientConfig struct {
	RPCEndpoint  string
	PollInterval time.Duration
	NumOfRetries int
	StartHeight  *big.Int
	EndHeight    *big.Int
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
