package core

// RegisterType ... One byte register type enum
type RegisterType uint8

const (
	AccountBalance RegisterType = iota + 1
	BlackholeTX
	ContractCreateTX
	GethBlock
)

// String ... Returns string representation of a
// register enum
func (rt RegisterType) String() string {
	switch rt {
	case AccountBalance:
		return "account.balance"

	case ContractCreateTX:
		return "contract.create.tx"

	case BlackholeTX:
		return "blackhole.tx"

	case GethBlock:
		return "geth.block"
	}

	return UnknownType
}

// DataRegister ...Represents an ETL subsytem data type that
// can be produced and consumed by heterogenous components
type DataRegister struct {
	DataType             RegisterType
	ComponentType        ComponentType
	ComponentConstructor interface{}
	Dependencies         []*DataRegister
}
