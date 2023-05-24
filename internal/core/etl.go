package core

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

func StringToPipelineType(stringType string) PipelineType {
	switch stringType {
	case "backtest":
		return Backtest

	case "live":
		return Live

	case "mocktest":
		return MockTest
	}

	return PipelineType(0)
}

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
