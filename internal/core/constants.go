package core

import "github.com/base-org/pessimism/internal/logging"

type CtxKey uint8

const (
	Logger CtxKey = iota
	Metrics
	State
	L1Client
	L2Client
)

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

// String ... Converts a network to a string
func (n Network) String() string {
	switch n {
	case Layer1:
		return "layer1"

	case Layer2:
		return "layer2"
	}

	return UnknownType
}

// StringToNetwork ... Converts a string to a network
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

// InvariantType ... Represents the type of invariant
type InvariantType uint8

const (
	BalanceEnforcement = iota + 1
	ContractEvent
	WithdrawalEnforcement
)

// String ... Converts an invariant type to a string
func (it InvariantType) String() string {
	switch it {
	case BalanceEnforcement:
		return "balance_enforcement"

	case ContractEvent:
		return "contract_event"

	case WithdrawalEnforcement:
		return "withdrawal_enforcement"

	default:
		return "unknown"
	}
}

// StringToInvariantType ... Converts a string to an invariant type
func StringToInvariantType(stringType string) InvariantType {
	switch stringType {
	case "balance_enforcement":
		return BalanceEnforcement

	case "contract_event":
		return ContractEvent

	case "withdrawal_enforcement":
		return WithdrawalEnforcement

	default:
		return InvariantType(0)
	}
}

// AlertDestination ... The destination for an alert
type AlertDestination uint8

const (
	Slack      AlertDestination = iota + 1
	ThirdParty                  // 2
)

// String ... Converts an alerting destination type to a string
func (ad AlertDestination) String() string {
	switch ad {
	case Slack:
		return "slack"
	case ThirdParty:
		return "third_party"
	default:
		return "unknown"
	}
}

// StringToAlertingDestType ... Converts a string to an alerting destination type
func StringToAlertingDestType(stringType string) AlertDestination {
	switch stringType {
	case "slack":
		return Slack

	case "third_party":
		return ThirdParty
	}

	return AlertDestination(0)
}

// ID keys used for logging
const (
	AddrKey logging.LogKey = "address"

	CUUIDKey logging.LogKey = "cuuid"
	PUUIDKey logging.LogKey = "puuid"
	SUUIDKey logging.LogKey = "suuid"
)
