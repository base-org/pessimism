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
	Live PathType = iota + 1
)

// StringToPathType ... Converts a string to a pipeline type
func StringToPathType(stringType string) PathType {
	switch stringType {

	case "live":
		return Live

	}

	return PathType(0)
}

// String ... Converts the pipeline type to a string
func (pt PathType) String() string {
	switch pt {

	case Live:
		return "live"

	}

	return UnknownType
}
