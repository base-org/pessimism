package core

import (
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/retry"
)

func RetryStrategy() *retry.ExponentialStrategy {
	return &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
}

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
	PT          PathType
	AlertPolicy *AlertPolicy
	Type        HeuristicType
	Params      *SessionParams
}

// PathConfig ... Configuration passed through to a pipeline constructor
type PathConfig struct {
	Network      Network
	DataType     TopicType
	PathType     PathType
	ClientConfig *ClientConfig
}

// Backfill ... Returns true if the reader is configured to backfill
func (oc *ClientConfig) Backfill() bool {
	return oc.StartHeight != nil
}
