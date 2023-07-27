package core

// ComponentType ... Denotes the ETL component type
type ComponentType uint8

const (
	Oracle ComponentType = iota + 1
	Pipe
	Aggregator
)

// String ... Converts the component type to a string
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

// StringToPipelineType ... Converts a string to a pipeline type
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

// String ... Converts the pipeline type to a string
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
