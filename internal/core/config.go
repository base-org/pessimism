package core

import (
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/retry"
)

func RetryStrategy() *retry.ExponentialStrategy {
	return &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
}

type ClientConfig struct {
	Network      Network
	PollInterval time.Duration
	NumOfRetries int
	StartHeight  *big.Int
	EndHeight    *big.Int
}

type SessionConfig struct {
	Network     Network
	PT          PathType
	AlertPolicy *AlertPolicy
	Type        HeuristicType
	Params      *SessionParams
}

type PathConfig struct {
	Network      Network
	DataType     TopicType
	PathType     PathType
	ClientConfig *ClientConfig
}

func (oc *ClientConfig) Backfill() bool {
	return oc.StartHeight != nil
}
