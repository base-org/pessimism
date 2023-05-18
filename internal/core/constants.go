package core

import "github.com/base-org/pessimism/internal/logging"

// Network ... Represents the network for which a pipeline's oracle
// is subscribed to.
type Network uint8

const (
	Layer1 = iota + 1
	Layer2

	UnknownNetwork
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

func StringToNetwork(stringType string) Network {
	switch stringType {
	case "layer1":
		return Layer1

	case "layer2":
		return Layer2
	}

	return UnknownNetwork
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

type InvariantType uint8

const (
	ExampleInv InvariantType = iota
	TxCaller
	BalanceEnforcement

	UnknownInvariant
)

func (it InvariantType) String() string {
	switch it {
	case ExampleInv:
		return "example"

	case TxCaller:
		return "tx_caller"

	case BalanceEnforcement:
		return "balance_enforcement"

	default:
		return "unknown"
	}
}

func StringToInvariantType(stringType string) InvariantType {
	switch stringType {
	case "example":
		return ExampleInv

	case "tx_caller":
		return TxCaller

	case "balance_enforcement":
		return BalanceEnforcement

	default:
		return UnknownInvariant
	}
}

// ID keys used for logging
const (
	AddrKey logging.LogKey = "address"

	CUUIDKey logging.LogKey = "cuuid"
	PUUIDKey logging.LogKey = "puuid"
	SUUIDKey logging.LogKey = "suuid"
)
