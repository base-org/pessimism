package core

type RegisterType uint8

const (
	GethBlock RegisterType = iota + 1
	ContractCreateTX
	BlackholeTX
)

func (rt RegisterType) String() string {
	switch rt {
	case GethBlock:
		return "geth.block"

	case ContractCreateTX:
		return "contract.create.tx"

	case BlackholeTX:
		return "blackhole.tx"
	}

	return UnknownType
}

type DataRegister struct {
	DataType             RegisterType
	ComponentType        ComponentType
	ComponentConstructor interface{}
	Dependencies         []*DataRegister
}
