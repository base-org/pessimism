package core

type FilePath string

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
	Local       Env = "local"
)

type CtxKey uint8

const (
	Logger CtxKey = iota
	Metrics
	State
	L1Client
	L2Client
	L2Geth
	L2RawGeth
	L1RawGeth
)

// Network ... Represents the network for which a pipeline's oracle
// is subscribed to.
type Network uint8

const (
	Layer1 Network = iota + 1
	Layer2
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

	default:
		return UnknownType
	}
}

// StringToNetwork ... Converts a string to a network
func StringToNetwork(stringType string) Network {
	switch stringType {
	case "layer1":
		return Layer1

	case "layer2":
		return Layer2

	default:
		return Network(0)
	}
}

type ChainSubscription uint8

const (
	OnlyLayer1 ChainSubscription = iota + 1
	OnlyLayer2
	BothNetworks
)

type FetchType int

const (
	FetchHeader FetchType = 0
	FetchBlock  FetchType = 1
)

type Timeouts int

const (
	EthClientTimeout Timeouts = 20 // in seconds
)

// HeuristicType ... Represents the type of heuristic
type HeuristicType uint8

const (
	BalanceEnforcement HeuristicType = iota + 1
	ContractEvent
	WithdrawalEnforcement
	FaultDetector
)

// String ... Converts an heuristic type to a string
func (it HeuristicType) String() string {
	switch it {
	case BalanceEnforcement:
		return "balance_enforcement"

	case ContractEvent:
		return "contract_event"

	case WithdrawalEnforcement:
		return "withdrawal_enforcement"

	case FaultDetector:
		return "fault_detector"

	default:
		return "unknown"
	}
}

// StringToHeuristicType ... Converts a string to an heuristic type
func StringToHeuristicType(stringType string) HeuristicType {
	switch stringType {
	case "balance_enforcement":
		return BalanceEnforcement

	case "contract_event":
		return ContractEvent

	case "withdrawal_enforcement":
		return WithdrawalEnforcement

	case "fault_detector":
		return FaultDetector

	default:
		return HeuristicType(0)
	}
}

// AlertPolicy ... The alerting policy for a heuristic session
type AlertPolicy struct {
	Dest string `json:"destination"`
	Msg  string `json:"message"`
}

// Message ... Returns the message for an alert
func (ap *AlertPolicy) Message() string {
	return ap.Msg
}

// Destination ... Returns the destination for an alert
func (ap *AlertPolicy) Destination() AlertDestination {
	return StringToAlertingDestType(ap.Dest)
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
