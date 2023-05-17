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
	BalanceEnforcement
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

// ID keys used for logging
const (
	AddrKey logging.LogKey = "address"

	CUUIDKey logging.LogKey = "cuuid"
	PUUIDKey logging.LogKey = "puuid"
	SUUIDKey logging.LogKey = "suuid"
)
