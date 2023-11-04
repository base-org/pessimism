package core

// ProcessType ... Denotes the ETL component type
type ProcessType uint8

const (
	Read ProcessType = iota + 1
	Subscribe
)

// String ... Converts the component type to a string
func (ct ProcessType) String() string {
	switch ct {
	case Read:
		return "reader"

	case Subscribe:
		return "subscriber"
	}

	return UnknownType
}

// PathType ...
type PathType uint8

const (
	Backtest PathType = iota + 1
	Live
	MockTest
)

// StringToPathType ... Converts a string to a pipeline type
func StringToPathType(stringType string) PathType {
	switch stringType {
	case "backtest":
		return Backtest

	case "live":
		return Live

	case "mocktest":
		return MockTest
	}

	return PathType(0)
}

// String ... Converts the pipeline type to a string
func (pt PathType) String() string {
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
