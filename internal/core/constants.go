package core

import "github.com/base-org/pessimism/internal/logging"

// Network ... Represents the network for which a pipeline's oracle
// is subscribed to.
type Network uint8

const (
	Layer1 = iota + 1
	Layer2
)

const (
	UnknownType = "unknown"
)

func (n Network) String() string {
	switch n {
	case Layer1:
		return "layer1"

	case Layer2:
		return "layer2"
	}

	return UnknownType
}

// ComponentType ...
type ComponentType uint8

const (
	Oracle ComponentType = iota + 1
	Pipe
	Aggregator
)

func (ct ComponentType) String() string {
	switch ct {
	case Oracle:
		return "oracle"

	case Pipe:
		return "pipe"

	case Aggregator:
		return "aggregator"
	}

	return UnknownType
}

// PipelineType ...
type PipelineType uint8

const (
	Backtest PipelineType = iota + 1
	Live
	MockTest
)

func (pt PipelineType) String() string {
	switch pt {
	case Backtest:
		return "backtest"

	case Live:
		return "live"

	case MockTest:
		return "mocktest"
	}

	return UnknownType
}

type FetchType int

const (
	FetchHeader FetchType = 0
	FetchBlock  FetchType = 1
)

type Timeouts int

const (
	EthClientTimeout Timeouts = 20 // in seconds
)

type InvariantType int

const (
	ExampleInv InvariantType = iota
	TxCaller
)

func (it InvariantType) String() string {
	switch it {
	case ExampleInv:
		return "example"

	case TxCaller:
		return "tx_caller"

	default:
		return "unknown"
	}
}

const (
	CUUIDKey logging.LogKey = "cuuid"
	PUUIDKey logging.LogKey = "puuid"
	SUUIDKey logging.LogKey = "suuid"
)
