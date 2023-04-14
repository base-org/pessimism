package core

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
