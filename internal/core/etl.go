package core

// ProcessType ... Denotes the ETL process type
type ProcessType uint8

const (
	Read ProcessType = iota + 1
	Subscribe
)

// String ... Converts the process type to a string
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
